//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

func TestEBPFPeerProcess(t *testing.T) { testpeer.Run(t) }
func TestNativeBidirectionalHandoverAndCleanup(t *testing.T) {
	testNativeBidirectionalHandover(t, false)
}
func TestNativeRemotePeerHandoverAndCleanup(t *testing.T) {
	testNativeBidirectionalHandover(t, true)
}
func testNativeBidirectionalHandover(t *testing.T, changeRemote bool) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	steps := []string{"1", "2", "1"}
	if changeRemote {
		steps[1] = "2-remote"
	}
	join := testpeer.Start(t, "TestEBPFPeerProcess", steps...)
	endpoint, device, err := userspace.NewTUN("pr-ue")
	require.NoError(t, err)
	defer endpoint.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 1456))
	require.NoError(t, netlink.AddrAdd(device, &netlink.Addr{IPNet: testpeer.Network("10.60.0.1/32")}))
	require.NoError(t, netlink.LinkSetUp(device))
	r := NewRegistry()
	cfg := Config{Local: netip.MustParseAddr("10.88.0.1"), Remote: netip.MustParseAddr("10.88.0.2"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 1001, DownlinkTEID: 2001, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 1456}
	session, err := r.Open(cfg)
	require.NoError(t, err)
	// Reinstallation replaces session; cleanup must own the current instance.
	defer func() { _ = session.Close() }()
	rule := netlink.NewRule()
	rule.Priority = 100
	rule.Table = 1600
	rule.Src = testpeer.Network("10.60.0.1/32")
	require.NoError(t, netlink.RuleAdd(rule))
	defer netlink.RuleDel(rule)
	encap := &netlink.BpfEncap{}
	require.NoError(t, encap.SetProg(nl.LWT_BPF_XMIT, session.ProgramFD(), "packetrusher_gtpu"))
	require.NoError(t, encap.SetXmitHeadroom(44))
	route := &netlink.Route{Dst: testpeer.Network("0.0.0.0/0"), LinkIndex: device.Attrs().Index, Scope: netlink.SCOPE_LINK, Table: 1600, Src: net.ParseIP("10.60.0.1"), Encap: encap}
	require.NoError(t, netlink.RouteAdd(route))
	defer netlink.RouteDel(route)
	app, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.60.0.1")}, &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 9000})
	require.NoError(t, err)
	defer app.Close()
	exchange := func(label string) {
		t.Helper()
		require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
		_, e := app.Write([]byte(label))
		require.NoError(t, e)
		b := make([]byte, 1500)
		n, e := app.Read(b)
		if e != nil {
			ul, dl, dropped, se := r.Stats()
			t.Logf("failed exchange kernel stats: UL=%d DL=%d ingressDrops=%d statsErr=%v peers=%v", ul, dl, dropped, se, r.peers.Load())
		}
		require.NoError(t, e)
		require.Equal(t, label, string(b[:n]))
	}
	exchange("initial-real-bpf")
	next := cfg
	next.Local = netip.MustParseAddr("10.88.0.3")
	if changeRemote {
		next.Remote = netip.MustParseAddr("10.88.0.4")
	}
	next.UplinkTEID = 1002
	next.DownlinkTEID = 2002
	require.NoError(t, session.Update(next))
	exchange("same-socket-after-n3-teid-handover")
	if changeRemote {
		_, _, dropped, err := r.Stats()
		require.NoError(t, err)
		require.GreaterOrEqual(t, dropped, uint64(5), "wrong TEID/UE twice and the previous UPF must be dropped")
	}
	oldSocket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.1"), Port: 2152})
	require.NoError(t, err)
	require.NoError(t, oldSocket.Close())
	plain, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.3")}, &net.UDPAddr{IP: net.ParseIP("10.88.0.2"), Port: 9999})
	require.NoError(t, err)
	require.NoError(t, plain.SetDeadline(time.Now().Add(5*time.Second)))
	_, err = plain.Write([]byte("unrelated-udp"))
	require.NoError(t, err)
	b := make([]byte, 100)
	n, err := plain.Read(b)
	require.NoError(t, err)
	require.Equal(t, "unrelated-udp", string(b[:n]))
	plain.Close()
	// The caller owns the LWT route; retire it before dropping its program/maps.
	require.NoError(t, netlink.RouteDel(route))
	require.NoError(t, session.Close())
	require.Empty(t, r.links)
	require.Empty(t, r.locals)
	targetSocket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.3"), Port: 2152})
	require.NoError(t, err)
	targetSocket.Close()
	// Reinstall proves owned ingress links and tuple maps were fully retired.
	session, err = r.Open(cfg)
	require.NoError(t, err)
	encap = &netlink.BpfEncap{}
	require.NoError(t, encap.SetProg(nl.LWT_BPF_XMIT, session.ProgramFD(), "packetrusher_gtpu"))
	require.NoError(t, encap.SetXmitHeadroom(44))
	route.Encap = encap
	require.NoError(t, netlink.RouteAdd(route))
	exchange("reinstalled-real-bpf")
	ul, dl, dropped, err := r.Stats()
	require.NoError(t, err)
	require.Equal(t, uint64(1), ul)
	require.Equal(t, uint64(1), dl)
	require.GreaterOrEqual(t, dropped, uint64(2))
	t.Logf("actual kernel counters after reinstall: uplink=%d downlink=%d owned-drop=%d", ul, dl, dropped)
	require.NoError(t, netlink.RouteDel(route))
	require.NoError(t, session.Close())
	join()
	require.Empty(t, r.sessions)
	require.Empty(t, r.links)
	require.Empty(t, r.locals)
	require.Nil(t, r.collection)
}
