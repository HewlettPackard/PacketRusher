// SPDX-License-Identifier: Apache-2.0
package context

import (
	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

func TestPDUAddressFamiliesAndDowngrade(t *testing.T) {
	iid := []byte{0, 0, 0, 0, 0, 0, 0, 7}
	for _, tc := range []struct {
		requested, selected uint8
		address             *ie.PDUAddr
		valid               bool
	}{
		{1, 1, &ie.PDUAddr{IPv4: []byte{10, 0, 0, 2}}, true},
		{2, 2, &ie.PDUAddr{IPv6IfId: iid}, true},
		{3, 3, &ie.PDUAddr{IPv4: []byte{10, 0, 0, 2}, IPv6IfId: iid}, true},
		{3, 1, &ie.PDUAddr{IPv4: []byte{10, 0, 0, 2}}, true},
		{3, 2, &ie.PDUAddr{IPv6IfId: iid}, true},
		{1, 2, &ie.PDUAddr{IPv6IfId: iid}, false},
		{2, 1, &ie.PDUAddr{IPv4: []byte{10, 0, 0, 2}}, false},
		{3, 3, &ie.PDUAddr{IPv6IfId: iid}, false},
		{2, 2, &ie.PDUAddr{IPv6IfId: iid, IPv4: []byte{10, 0, 0, 2}}, false},
		{2, 2, &ie.PDUAddr{IPv6IfId: []byte{1}}, false},
		{2, 2, &ie.PDUAddr{IPv6IfId: iid, SMFIPv6LLA: netip.MustParseAddr("2001:db8::1").AsSlice()}, false},
	} {
		session := &UEPDUSession{requestedSessionType: tc.requested}
		err := session.SetPDUAddress(tc.selected, tc.address)
		if !tc.valid {
			require.Error(t, err)
			require.Empty(t, session.GetIp())
			continue
		}
		require.NoError(t, err)
		_, hasIPv6 := session.GetIPv6InterfaceID()
		require.Equal(t, tc.selected != 1, hasIPv6)
		require.Equal(t, tc.selected != 2, session.GetIPv4().IsValid())
		if hasIPv6 {
			require.False(t, session.GetIPv6().IsValid(), "NAS must not invent a global prefix")
			require.Error(t, session.SetIPv6(netip.MustParseAddr("2001:db8::8")))
			require.NoError(t, session.SetIPv6(netip.MustParseAddr("2001:db8::7")))
			require.Equal(t, "2001:db8::7", session.GetIPv6().String())
		}
	}
}
