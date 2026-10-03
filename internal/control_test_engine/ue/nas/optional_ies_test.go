// SPDX-License-Identifier: Apache-2.0
package nas

import (
	"reflect"
	"testing"

	"github.com/free5gc/nas/ie"
	message "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/auth"
	"my5G-RANTester/internal/common/sidf"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"my5G-RANTester/internal/testutil/naswire"
)

func TestPreviouslySupportedOptionalIEsOnIndependentProtectedWire(t *testing.T) {
	for _, fixture := range naswire.OptionalIEs {
		t.Run(fixture.Name, func(t *testing.T) {
			plain := naswire.Hex(fixture.Wire)
			ue := securityTestUE()
			ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
			ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt = [16]byte{5, 6, 7, 8}, [16]byte{1, 2, 3, 4}
			header, count := byte(2), uint32(0x0207)
			if plain[0] == 0x7e && plain[2] == 0x5d {
				header, count = 3, 0
				require.NoError(t, auth.AlgorithmKeyDerivation(2, ue.UeSecurity.Kamf, &ue.UeSecurity.KnasEnc, 2, &ue.UeSecurity.KnasInt))
			}
			inner := plain
			nested := plain[0] == 0x2e
			if nested {
				plain = naswire.WrapSM(inner)
			}
			packet := naswire.Protect(plain, ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt, count, 1, header)
			before := ue.UeSecurity
			bad := append([]byte(nil), packet...)
			bad[2] ^= 0x80
			msg, err := DecodeNAS(ue, bad)
			require.Error(t, err)
			require.Nil(t, msg)
			require.Equal(t, before, ue.UeSecurity, "invalid MAC must not commit security state with optional IEs")
			msg, err = DecodeNAS(ue, packet)
			require.NoError(t, err)
			require.Equal(t, count, ue.UeSecurity.DLCount.Get())
			require.Equal(t, before.ULCount, ue.UeSecurity.ULCount)
			if nested {
				current := ue.UeSecurity
				direct, directErr := DecodeNAS(ue, inner)
				require.NoError(t, directErr, "5GSM session IDs are not security header types")
				require.NotNil(t, direct)
				require.Equal(t, current, ue.UeSecurity)
				msg = nas_control.GetNasPduFromPduAccept(msg.(*message.DLNASTransport))
			}
			require.NotNil(t, msg, "nested SM warning must not discard a supported message")
			require.False(t, reflect.ValueOf(msg).Elem().FieldByName(fixture.Field).IsNil(), "native optional field must be preserved")
			wire, err := msg.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, inner, wire, "decoded IE must preserve the independent wire value")
			if header != 3 {
				plainMsg, err := DecodeNAS(ue, plain)
				require.NoError(t, err, "supported ordinary plaintext optional IEs remain accepted")
				require.NotNil(t, plainMsg)
			}
		})
	}
}

func TestRegistrationWithOptionalIEStillDispatchesRegistrationComplete(t *testing.T) {
	for _, suffix := range []string{"a1", "b0", "34030201f1"} {
		t.Run(suffix, func(t *testing.T) {
			scenarioUpdates := make(chan scenario.ScenarioMessage, 1)
			outgoing := make(chan gnbcontext.UEMessage, 1)
			ue := &context.UEContext{}
			ue.NewRanUeContext("001002086", &ie.UESecCapability{Length: 2, EA2_128_5G: true, IA2_128_5G: true}, "", "", "", "8000", "000000000000", "208", "93", sidf.HomeNetworkPublicKey{ProtectionScheme: "0", PublicKeyID: "0"}, "0", "internet", 1, "010203", config.TunnelMode(0), scenarioUpdates, nil, 1)
			ue.SetGnbRx(outgoing)
			ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt = [16]byte{5, 6, 7, 8}, [16]byte{1, 2, 3, 4}
			packet := naswire.Protect(naswire.Hex("7e00420101"+suffix), ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt, 7, 1, 2)
			core := ue.NASSecurityContext().Clone()
			core.Side = message.CoreNetworkSide
			DispatchNas(ue, packet)
			require.Equal(t, context.MM5G_REGISTERED, ue.GetStateMM())
			select {
			case sent := <-outgoing:
				require.True(t, sent.IsNas)
				reply, err := message.Parse(sent.Nas, core)
				require.NoError(t, err)
				require.IsType(t, &message.RegComplete{}, reply)
			default:
				t.Fatal("Registration Complete was not sent")
			}
		})
	}
}

func TestMalformedOptionalIEsCannotCommitProtectedCounter(t *testing.T) {
	// Declared TLV/TLV-E lengths exceed the bytes on the wire, including a
	// duplicate IE that upstream used to skip without checking its envelope.
	for _, wire := range []string{
		"7e0042010134040201f1", "7e004201017a00050001f100",
		"7e0042010160032000", "7e0042010134020201",
		"7e0042010173001000000000000000000000000000000000",
		"7e004201016002200060042000",
		"7e0068010000", "7e00680100062e0101d11f", "7e00680100052e0101d11f2402aa",
	} {
		t.Run(wire, func(t *testing.T) {
			ue := securityTestUE()
			ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
			before := ue.UeSecurity
			plain := naswire.Hex(wire)
			_, err := DecodeNAS(ue, plain)
			require.Error(t, err)
			packet := naswire.Protect(plain, ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt, 0x0207, 1, 2)
			_, err = DecodeNAS(ue, packet)
			require.Error(t, err)
			require.Equal(t, before, ue.UeSecurity, "valid MAC does not make a malformed IE acceptable")
		})
	}
	inner := naswire.Hex("2e0101c211000601000320ff01060103e80103e875000550000180")
	require.Nil(t, nas_control.GetNasPduFromPduAccept(&message.DLNASTransport{PayloadCntr: &ie.PayloadCntr{Contents: inner}}))
	// A newer unsupported IE retains an explicit error, including nested SM.
	_, err := DecodeNAS(securityTestUE(), naswire.Hex("7e004201011b0100"))
	require.Error(t, err)
	inner = naswire.Hex("2e0101c211000601000320ff01060103e80103e8c1")
	require.Nil(t, nas_control.GetNasPduFromPduAccept(&message.DLNASTransport{PayloadCntr: &ie.PayloadCntr{Contents: inner}}))
}
