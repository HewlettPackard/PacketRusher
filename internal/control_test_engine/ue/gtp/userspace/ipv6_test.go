/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package userspace

import (
	"bytes"
	"encoding/hex"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouterDiscovery(t *testing.T) {
	// fe80::7 asks all routers; fe80::1 advertises 2001:db8:1:2::/64 to all nodes.
	solicitation, _ := hex.DecodeString("6000000000083aff" + "fe800000000000000000000000000007" +
		"ff020000000000000000000000000002" + "85007d3000000000")
	advertisement, _ := hex.DecodeString("6000000000303aff" + "fe800000000000000000000000000001" +
		"ff020000000000000000000000000001" + "8600c386400007080000000000000000" +
		"030440c0ffffffffffffffff00000000" + "20010db8000100020000000000000000")
	require.Equal(t, solicitation, routerSolicitation(netip.MustParseAddr("fe80::7"))[Headroom:])

	change := func(index int, value byte) []byte {
		packet := bytes.Clone(advertisement)
		packet[index] = value
		return packet
	}
	fragment := append(change(6, 44)[:40], 58, 0, 0, 1, 0, 0, 0, 7) // more fragments follow
	none := netip.Prefix{}
	for name, tc := range map[string]struct {
		packet        []byte
		prefix        netip.Prefix
		advertisement bool
	}{
		"Router Advertisement":   {advertisement, netip.MustParsePrefix("2001:db8:1:2::/64"), true},
		"wrong checksum":         {change(87, 1), none, true},
		"Neighbor Advertisement": {change(40, 136), none, false},
		"fragment of ICMPv6":     {append(fragment, advertisement[40:]...), none, false},
		"IPv6 header only":       {advertisement[:40], none, false},
	} {
		prefix, isAdvertisement := advertisedPrefix(tc.packet)
		require.Equal(t, tc.advertisement, isAdvertisement, name)
		require.Equal(t, tc.prefix, prefix, name)
	}
}
