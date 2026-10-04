/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package nas

import (
	"encoding/hex"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

// The PDU address of a PDU Session Establishment Accept is an IPv4 address, an IPv6
// interface identifier, of which the UE makes its link-local address, or both
// (TS 24.501 §9.11.4.10).
func TestPDUSessionAcceptAddress(t *testing.T) {
	for name, tc := range map[string]struct {
		sessionType, address, ipv4 string
		ipv6                       netip.Addr
	}{
		"IPv4":   {"11", "2905010a000002", "10.0.0.2", netip.Addr{}},
		"IPv6":   {"12", "2909020000000000000007", "", netip.MustParseAddr("fe80::7")},
		"IPv4v6": {"13", "290d0300000000000000070a000002", "10.0.0.2", netip.MustParseAddr("fe80::7")},
	} {
		ue := securityTestUE()
		ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
		session, err := ue.CreatePDUSession()
		require.NoError(t, err, name)
		session.SetStateSM_PDU_SESSION_PENDING()
		// PDU session 1, its type with SSC mode 1, QoS rules and session AMBR.
		accept, err := hex.DecodeString("2e0101c2" + tc.sessionType + "000601000320ff01060103e80103e8" + tc.address)
		require.NoError(t, err, name)
		// The DL NAS Transport carrying it as the N1 SM container of PDU session 1.
		transport := append([]byte{0x7e, 0, 0x68, 1, byte(len(accept) >> 8), byte(len(accept))}, accept...)
		DispatchNas(ue, protect(ue, append(transport, 0x12, 1), 0x0207))
		require.Equal(t, context.SM5G_PDU_SESSION_ACTIVE, session.GetStateSM(), name)
		require.Equal(t, tc.ipv4, session.GetIp(), name)
		require.Equal(t, tc.ipv6, session.GetIPv6(), name)
	}
}
