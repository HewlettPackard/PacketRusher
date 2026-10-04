/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package ebpfgtp

import (
	"net"
	"net/netip"
	"os"
	"testing"

	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

// The kernel's verifier accepts the programs and they attach to a tunnel's devices.
func TestLoadAndAttach(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("set PACKETRUSHER_TUN_TEST=1 in a private network namespace with CAP_NET_ADMIN and CAP_BPF")
	}
	n3 := netip.MustParseAddr("192.0.2.1")
	link := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "n3test"}}
	require.NoError(t, netlink.LinkAdd(link))
	t.Cleanup(func() { netlink.LinkDel(link) })
	require.NoError(t, netlink.AddrAdd(link, &netlink.Addr{IPNet: &net.IPNet{IP: n3.AsSlice(), Mask: net.CIDRMask(24, 32)}}))
	require.NoError(t, netlink.LinkSetUp(link))
	require.NoError(t, Load())
	t.Cleanup(Close)

	config := userspace.Config{UPF: netip.MustParseAddr("192.0.2.2"), UplinkTEID: 1, DownlinkTEID: 2, QFI: 9}
	session, err := userspace.Open("tuntest", n3, config)
	require.NoError(t, err)
	t.Cleanup(session.Close)
	require.NoError(t, Attach("tuntest", n3, config))
	config.DownlinkTEID = 3 // as a handover does
	require.NoError(t, Attach("tuntest", n3, config))

	var tun uint32
	require.Error(t, objects.Maps["downlinks"].Lookup(downlinkKey{Local: n3.As4(), TEID: bigEndian(2)}, &tun))
	require.NoError(t, objects.Maps["downlinks"].Lookup(downlinkKey{Local: n3.As4(), TEID: bigEndian(3)}, &tun))
	Detach("tuntest")
	require.Empty(t, decaps)
}
