// SPDX-License-Identifier: Apache-2.0
package nas

import (
	"bytes"
	"testing"

	"github.com/free5gc/nas/ie"
	message "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/common/auth"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func TestSecurityModeCommandCommitsOnlyAfterIntegrityValidation(t *testing.T) {
	ue := securityTestUE()
	packet := protectedSecurityModeCommand(t, ue, ie.IntegrityAlgo_1285GIA2)
	bad := append([]byte(nil), packet...)
	bad[2] ^= 0x80
	before := ue.UeSecurity
	_, err := DecodeNAS(ue, bad)
	require.Error(t, err)
	require.Equal(t, before, ue.UeSecurity, "bad MAC must not install algorithms, keys or counters")
	msg, err := DecodeNAS(ue, packet)
	require.NoError(t, err)
	require.IsType(t, &message.SecModeCmd{}, msg)
	require.Equal(t, uint8(2), ue.UeSecurity.CipheringAlg)
	require.Equal(t, uint8(2), ue.UeSecurity.IntegrityAlg)
	require.NotEqual(t, before.KnasInt, ue.UeSecurity.KnasInt)
}

func TestSecurityModeCommandRejectsUnadvertisedNullIntegrity(t *testing.T) {
	ue := securityTestUE()
	packet := protectedSecurityModeCommand(t, ue, ie.IntegrityAlgo_5GIA0)
	before := ue.UeSecurity
	_, err := DecodeNAS(ue, packet)
	require.ErrorContains(t, err, "did not offer")
	require.Equal(t, before, ue.UeSecurity)
}

func TestPlainSecurityModeCommandIsRejectedWithoutChangingSecurity(t *testing.T) {
	ue := securityTestUE()
	protected := protectedSecurityModeCommand(t, ue, ie.IntegrityAlgo_1285GIA2)
	plain := protected[message.SecHdrLen:]
	decoded, err := message.Parse(plain, nil)
	require.NoError(t, err)
	require.IsType(t, &message.SecModeCmd{}, decoded, "the regression must use a valid plaintext SMC")
	before := ue.UeSecurity
	msg, err := DecodeNAS(ue, plain)
	require.ErrorContains(t, err, "new NAS security context")
	require.Nil(t, msg)
	require.Equal(t, before, ue.UeSecurity, "plaintext SMC must not change key identifiers, algorithms, keys or counters")
}

func TestExistingContextSecurityModeCommandIsRejectedAfterValidMAC(t *testing.T) {
	for _, header := range []message.SecHdrType{message.SecHdrTypeIntegrityProtected, message.SecHdrTypeIntegrityProtectedAndCiphered} {
		t.Run(header.String(), func(t *testing.T) {
			ue := securityTestUE()
			ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
			require.NoError(t, auth.AlgorithmKeyDerivation(2, ue.UeSecurity.Kamf, &ue.UeSecurity.KnasEnc, 2, &ue.UeSecurity.KnasInt))
			protected := protectedSecurityModeCommand(t, ue, ie.IntegrityAlgo_1285GIA2)
			command, err := message.Parse(protected[message.SecHdrLen:], nil)
			require.NoError(t, err)
			core := ue.NASSecurityContext().Clone()
			core.Side = message.CoreNetworkSide
			core.DownlinkCount.Set(2, 7)
			packet, err := message.Marshal(command, core, header)
			require.NoError(t, err)
			verified, err := message.Parse(packet, ue.NASSecurityContext().Clone())
			require.NoError(t, err, "the packet must have a valid MAC under the current context")
			require.IsType(t, &message.SecModeCmd{}, verified)
			before := ue.UeSecurity
			msg, err := DecodeNAS(ue, packet)
			require.ErrorContains(t, err, "new NAS security context")
			require.Nil(t, msg)
			require.Equal(t, before, ue.UeSecurity, "rejected SMC must not commit the verified receive counter")
		})
	}
}

func TestOtherPlainNASRetainsCurrentContextBehavior(t *testing.T) {
	ue := securityTestUE()
	plain, err := (&message.IdReq{IdType: &ie.IdType5GS{IdType: ie.IdType_5GS_SUCI}}).MarshalBinary()
	require.NoError(t, err)
	before := ue.UeSecurity
	msg, err := DecodeNAS(ue, plain)
	require.NoError(t, err)
	require.IsType(t, &message.IdReq{}, msg)
	require.Equal(t, before, ue.UeSecurity)
}

func TestMalformedNASDoesNotMutateSecurity(t *testing.T) {
	ue := securityTestUE()
	before := ue.UeSecurity
	for _, packet := range [][]byte{nil, {}, {0x7e}, {0x7e, 3}, {0x7e, 3, 0, 0, 0, 0}, {0x7e, 4, 0, 0, 0, 0, 0}, {0x7e, 15, 0, 0, 0, 0, 0}} {
		_, err := DecodeNAS(ue, packet)
		require.Error(t, err)
		require.Equal(t, before, ue.UeSecurity)
	}
}

func securityTestUE() *context.UEContext {
	ue := &context.UEContext{}
	ue.UeSecurity.Kamf = bytes.Repeat([]byte{0x42}, 32)
	ue.UeSecurity.IntegrityAlg = 1
	ue.UeSecurity.ULCount.Set(1, 5)
	ue.UeSecurity.DLCount.Set(2, 6)
	ue.UeSecurity.UeSecurityCapability = &ie.UESecCapability{Length: 2, EA05G: true, EA2_128_5G: true, IA2_128_5G: true}
	return ue
}

func protectedSecurityModeCommand(t *testing.T, ue *context.UEContext, integrity ie.AlgIntegrity) []byte {
	t.Helper()
	var enc, intKey [16]byte
	require.NoError(t, auth.AlgorithmKeyDerivation(2, ue.UeSecurity.Kamf, &enc, uint8(integrity), &intKey))
	core := message.NewSecCtx(message.CoreNetworkSide, message.Bearer3GPP, ie.EncAlgo_1285GEA2, integrity, enc[:], intKey[:])
	msg := &message.SecModeCmd{
		SelectedNASSecAlgos:       &ie.NASSecAlgos{CipheringAlgo: ie.EncAlgo_1285GEA2, MsgIntAlgo: integrity},
		Ngksi:                     &ie.NASKeySetId{Ksi: 1, Tsc: ie.SecCtxTypeNative},
		ReplayedUESecCapabilities: ue.GetUeSecurityCapability(),
	}
	packet, err := message.Marshal(msg, core, message.SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx)
	require.NoError(t, err)
	return packet
}
