//go:build linux

// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Run after namespace preparation with CAP_BPF/SYS_ADMIN/PERFMON dropped,
// retaining NET_ADMIN for the portable fallback's real TUN/routing lifecycle.
func TestNativeDefaultFallbackWithoutBPFCapability(t *testing.T) {
	if os.Getenv("PACKETRUSHER_FALLBACK_TEST") != "1" {
		t.Skip("owned namespace and missing BPF capability required")
	}
	status, err := os.ReadFile("/proc/self/status")
	require.NoError(t, err)
	var effective uint64
	for _, line := range strings.Split(string(status), "\n") {
		if text, ok := strings.CutPrefix(line, "CapEff:\t"); ok {
			effective, err = strconv.ParseUint(text, 16, 64)
			require.NoError(t, err)
		}
	}
	for _, bit := range []uint{21, 38, 39} {
		require.Zero(t, effective&(1<<bit))
	}
	require.NotZero(t, effective&(1<<12))
	t.Logf("CapEff=%016x; NET_ADMIN present, BPF/SYS_ADMIN/PERFMON absent", effective)
	oldRegistry, oldUserspace, oldTables := ebpfRegistry, userspaceRegistry, sessionRoutingTables
	ebpfRegistry = ebpfgtp.NewRegistry()
	userspaceRegistry = userspace.NewRegistry()
	sessionRoutingTables = newRoutingTableAllocator(netlink.RouteAdd, netlink.RouteDel, routingTableInUse)
	t.Cleanup(func() { ebpfRegistry, userspaceRegistry, sessionRoutingTables = oldRegistry, oldUserspace, oldTables })
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		a, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, a))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
	}
	peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.88.4.9"), Port: 2152})
	require.NoError(t, err)
	defer peer.Close()
	ue, pdu := sharedSetupUE(t, 1)
	ue.TunnelBackend = ""
	pdu.SetIp([12]uint8{10, 60, 0, 1})
	msg := userspaceTestMessage(t, "127.88.4.1", 60)
	SetupGtpInterface(ue, msg)
	require.NotNil(t, pdu.GetTunInterface())
	backend, cause := pdu.TunnelSelection()
	require.Equal(t, config.TunnelBackendUserspace, backend)
	require.Contains(t, cause, "eBPF backend unavailable")
	device := pdu.GetTunInterface()
	table := pdu.GetTunRoute().Table
	require.Equal(t, int(firstRoutingTable), table, "clean eBPF rollback released its original routing claim")
	app, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.60.0.1")}, &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 9000})
	require.NoError(t, err)
	userspaceRoundTrip(t, app, peer, msg.GnbIp, 60, 61)
	require.NoError(t, app.Close())
	pdu.ReleaseTunnel()
	_, err = netlink.LinkByName(device.Attrs().Name)
	require.Error(t, err)
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes)
	port, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.88.4.1"), Port: 2152})
	require.NoError(t, err)
	require.NoError(t, port.Close())
	// Exactly the same capability failure is still an error for explicit eBPF;
	// it cannot publish a userspace tunnel or leave failed staging behind.
	ue.TunnelBackend = config.TunnelBackendEBPF
	SetupGtpInterface(ue, msg)
	require.Nil(t, pdu.GetTunInterface())
	backend, cause = pdu.TunnelSelection()
	require.Empty(t, backend)
	require.Empty(t, cause)
	_, err = netlink.LinkByName(device.Attrs().Name)
	require.Error(t, err)
	routes, err = netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes)
	require.Empty(t, sessionRoutingTables.sessions)
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	require.NoError(t, err)
	for _, rule := range rules {
		require.NotEqual(t, table, rule.Table)
	}
	// An actual retained policy/blackhole claim blocks fallback, even though the
	// initiating eBPF failure is the same missing capability as above.
	ue.TunnelBackend = config.TunnelBackendAuto
	oldDelete, oldFallback := ruleDel, setupUserspaceBackend
	fallbackCalls := 0
	ruleDel = func(*netlink.Rule) error { return unix.EIO }
	setupUserspaceBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
		fallbackCalls++
		return nil
	}
	t.Cleanup(func() { ruleDel, setupUserspaceBackend = oldDelete, oldFallback })
	SetupGtpInterface(ue, msg)
	ruleDel, setupUserspaceBackend = oldDelete, oldFallback
	require.Zero(t, fallbackCalls, "incomplete owned rollback must never attempt another backend")
	require.Nil(t, pdu.GetTunInterface())
	var retained *ebpfTunnel
	quarantinedEBPFEndpoints.Range(func(key, value any) bool {
		tunnel := key.(*ebpfTunnel)
		if tunnel.table != nil && tunnel.table.session == pdu {
			retained = tunnel
			require.ErrorIs(t, value.(error), unix.EIO)
		}
		return true
	})
	require.NotNil(t, retained)
	t.Cleanup(func() {
		// Only the fixture can prove the injected deletion failure is now gone.
		require.NoError(t, netlink.RuleDel(retained.rule))
		require.NoError(t, netlink.RouteDel(retained.table.claim))
		quarantinedEBPFEndpoints.Delete(retained)
	})
	require.True(t, errors.Is(retained.release(), ebpfgtp.ErrCleanupIncomplete))
	require.Empty(t, sessionRoutingTables.free)
	routes, err = netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Len(t, routes, 1, "the actual blackhole claim remains reserved")
	rules, err = netlink.RuleList(netlink.FAMILY_ALL)
	require.NoError(t, err)
	policyRetained := false
	for _, rule := range rules {
		policyRetained = policyRetained || rule.Table == table
	}
	require.True(t, policyRetained, "the original source policy is retained")
	_, err = netlink.LinkByName(device.Attrs().Name)
	require.Error(t, err, "the endpoint retired despite the retained routing claim")
}
