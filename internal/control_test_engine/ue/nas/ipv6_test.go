// SPDX-License-Identifier: Apache-2.0
package nas

import (
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/testutil/naswire"
	"testing"
)

func TestProtectedIPv6PDUSessionAcceptFromIndependentWire(t *testing.T) {
	for _, fixture := range []struct {
		requested  uint8
		wire, ipv4 string
	}{
		{2, "2e0101c212000601000320ff01060103e80103e82909020000000000000007", ""},
		{3, "2e0101c213000601000320ff01060103e80103e8290d0300000000000000070a000002", "10.0.0.2"},
		{3, "2e0101c211000601000320ff01060103e80103e82905010a000002", "10.0.0.2"},
	} {
		ue := securityTestUE()
		ue.PDUSessionType = config.PDUSessionType(fixture.requested)
		ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
		session, err := ue.CreatePDUSession()
		require.NoError(t, err)
		session.SetStateSM_PDU_SESSION_PENDING()
		packet := naswire.Protect(naswire.WrapSM(naswire.Hex(fixture.wire)), ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt, 0x0207, 1, 2)
		bad := append([]byte(nil), packet...)
		bad[2] ^= 0x80
		before := ue.UeSecurity.DLCount
		DispatchNas(ue, bad)
		require.Equal(t, before, ue.UeSecurity.DLCount)
		require.Equal(t, context.SM5G_PDU_SESSION_ACTIVE_PENDING, session.GetStateSM())
		DispatchNas(ue, packet)
		require.Equal(t, context.SM5G_PDU_SESSION_ACTIVE, session.GetStateSM())
		require.Equal(t, fixture.ipv4, session.GetIp())
		iid, has6 := session.GetIPv6InterfaceID()
		if fixture.wire[9] == '1' {
			require.False(t, has6)
		} else {
			require.True(t, has6)
			require.Equal(t, [8]byte{0, 0, 0, 0, 0, 0, 0, 7}, iid)
		}
		require.False(t, session.GetIPv6().IsValid(), "NAS carries IID; user-plane RA supplies prefix")
	}
}

func TestMalformedIPv6PDUAddressCannotActivateSession(t *testing.T) {
	for _, suffix := range []string{"29080200000000000007", "290a02000000000000000700", "290d0300000000000000070a000002"} {
		ue := securityTestUE()
		ue.PDUSessionType = config.PDUSessionType(2)
		session, err := ue.CreatePDUSession()
		require.NoError(t, err)
		session.SetStateSM_PDU_SESSION_PENDING()
		DispatchNas(ue, naswire.WrapSM(naswire.Hex("2e0101c212000601000320ff01060103e80103e8"+suffix)))
		require.Equal(t, context.SM5G_PDU_SESSION_ACTIVE_PENDING, session.GetStateSM())
		_, has6 := session.GetIPv6InterfaceID()
		require.False(t, has6)
	}
}
