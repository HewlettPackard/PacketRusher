// SPDX-License-Identifier: Apache-2.0
package gtp

import (
	"errors"
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

func TestPayloadMTU(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		underlay, configured, want int
		bad                        bool
	}{
		{"reported Ethernet case", 1500, 0, 1456, false},
		{"jumbo frames", 9000, 0, 8956, false},
		{"IPv4 length limit on loopback", 65536, 0, 65491, false},
		{"UPF override", 1500, 1400, 1400, false},
		{"override exceeds outer MTU", 1500, 1464, 0, true},
		{"negative override", 1500, -1, 0, true},
		{"IPv4 minimum", 1500, 67, 0, true},
		{"underlay too small", 100, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := payloadMTU(tc.underlay, tc.configured)
			if tc.bad {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestSetTunnelMTUFindsN3Alias(t *testing.T) {
	tunnel := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "valgnb", Index: 3, MTU: 1464}}
	n3 := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "n3", Index: 2, MTU: 1500}}
	other := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "other", Index: 1, MTU: 9000}}
	set := 0
	ops := mtuOperations{
		links: func() ([]netlink.Link, error) { return []netlink.Link{tunnel, other, n3}, nil },
		addresses: func(link netlink.Link, family int) ([]netlink.Addr, error) {
			require.Equal(t, netlink.FAMILY_V4, family)
			if link == n3 {
				return []netlink.Addr{testAddress("10.0.0.1"), testAddress("10.0.0.2")}, nil
			}
			return []netlink.Addr{testAddress("192.0.2.1")}, nil
		},
		set: func(link netlink.Link, mtu int) error {
			require.Same(t, tunnel, link)
			set = mtu
			return nil
		},
	}
	require.NoError(t, setTunnelMTU(tunnel, netip.MustParseAddr("::ffff:10.0.0.2"), 0, ops))
	require.Equal(t, 1456, set)
	require.Equal(t, 1456, tunnel.Attrs().MTU)
	require.NoError(t, setTunnelMTU(tunnel, netip.MustParseAddr("10.0.0.2"), 1400, ops))
	require.Equal(t, 1400, set)
	require.ErrorContains(t, setTunnelMTU(tunnel, netip.MustParseAddr("10.0.0.2"), 1464, ops), "between 68 and 1456")
	require.ErrorContains(t, setTunnelMTU(tunnel, netip.MustParseAddr("10.0.0.3"), 0, ops), "not assigned")
	require.Equal(t, 1400, set, "invalid configuration must not change the interface")
	ops.set = func(netlink.Link, int) error { return errors.New("permission denied") }
	require.ErrorContains(t, setTunnelMTU(tunnel, netip.MustParseAddr("10.0.0.2"), 0, ops), "permission denied")
	require.Equal(t, 1400, tunnel.Attrs().MTU, "failed updates must preserve the cached MTU")
}

func testAddress(ip string) netlink.Addr {
	return netlink.Addr{IPNet: &net.IPNet{IP: net.ParseIP(ip), Mask: net.CIDRMask(24, 32)}}
}
