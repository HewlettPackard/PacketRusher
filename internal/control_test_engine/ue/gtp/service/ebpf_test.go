// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
)

func TestEBPFServicePeerProcess(t *testing.T) { testpeer.Run(t) }
func ebpfTestMessage(t *testing.T, local string, ul, dl uint32) gnbContext.UEMessage {
	t.Helper()
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, "10.88.0.2", "01", "000000", 0, 9, 8, 9, ul, dl)
	require.NoError(t, err)
	var sessions [16]*gnbContext.GnbPDUSession
	sessions[0] = pdu
	return gnbContext.UEMessage{GNBPduSessions: sessions, GnbIp: netip.MustParseAddr(local)}
}
func TestEBPFUpdateFailureReleasesOnlyWhenRollbackFails(t *testing.T) {
	for _, rollbackFails := range []bool{false, true} {
		t.Run(fmt.Sprintf("rollbackFails=%t", rollbackFails), func(t *testing.T) {
			ue, pdu := sharedSetupUE(t, 1)
			ue.TunnelBackend = config.TunnelBackendEBPF
			source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
			target := ebpfTestMessage(t, "10.88.0.3", 1002, 2002)
			link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: "existing", MTU: 1456}}
			pdu.SetTunInterface(link)
			pdu.SetGnbIp(source.GnbIp)
			pdu.GnbPduSession = source.GNBPduSessions[0]
			released := 0
			pdu.SetTunnelCleanup(func(bool) { released++ })
			pdu.SetTunnelUpdate(func(*gnbContext.GnbPDUSession, netip.Addr) error {
				if rollbackFails {
					return fmt.Errorf("%w: restore failed", errTunnelRollback)
				}
				return errors.New("occupied target port")
			})
			SetupGtpInterface(ue, target)
			require.Equal(t, source.GnbIp, pdu.GetGnbIp())
			require.Same(t, source.GNBPduSessions[0], pdu.GnbPduSession)
			if rollbackFails {
				require.Equal(t, 1, released)
				require.Nil(t, pdu.GetTunInterface())
			} else {
				require.Zero(t, released)
				require.Same(t, link, pdu.GetTunInterface())
			}
			pdu.ReleaseTunnel()
			require.Equal(t, 1, released)
		})
	}
}
func TestNativeEBPFServiceRoutingHandoverRollbackAndRelease(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	join := testpeer.Start(t, "TestEBPFServicePeerProcess", "1", "1", "2", "1")
	previous := ebpfRegistry
	ebpfRegistry = ebpfgtp.NewRegistry()
	t.Cleanup(func() { ebpfRegistry = previous })
	ue, pdu := sharedSetupUE(t, 1)
	ue.TunnelBackend = config.TunnelBackendEBPF
	pdu.SetIp([12]uint8{10, 60, 0, 1})
	source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
	target := ebpfTestMessage(t, "10.88.0.3", 1002, 2002)
	SetupGtpInterface(ue, source)
	require.NotNil(t, pdu.GetTunInterface())
	require.NotNil(t, pdu.GetTunRoute())
	require.NotNil(t, pdu.GetTunRule())
	device := pdu.GetTunInterface()
	route := pdu.GetTunRoute()
	rule := pdu.GetTunRule()
	table := route.Table
	require.Equal(t, 1456, device.Attrs().MTU)
	_, err := netlink.LinkByName("gtp-gnb")
	require.Error(t, err, "eBPF must not create a kernel GTP device")
	app, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.60.0.1")}, &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 9000})
	require.NoError(t, err)
	defer app.Close()
	exchange := func(label string) {
		t.Helper()
		require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
		_, err := app.Write([]byte(label))
		require.NoError(t, err)
		b := make([]byte, 100)
		n, err := app.Read(b)
		require.NoError(t, err)
		require.Equal(t, label, string(b[:n]))
	}
	exchange("production-setup")
	occupied, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.3"), Port: 2152})
	require.NoError(t, err)
	SetupGtpInterface(ue, target)
	require.Same(t, device, pdu.GetTunInterface())
	require.Same(t, route, pdu.GetTunRoute())
	require.Same(t, rule, pdu.GetTunRule())
	require.Equal(t, source.GnbIp, pdu.GetGnbIp())
	exchange("source-retained-after-real-target-bind-failure")
	occupied.Close()
	SetupGtpInterface(ue, target)
	require.Same(t, device, pdu.GetTunInterface())
	require.Same(t, route, pdu.GetTunRoute())
	require.Same(t, rule, pdu.GetTunRule())
	require.Equal(t, target.GnbIp, pdu.GetGnbIp())
	exchange("production-handover-same-socket")
	pdu.ReleaseTunnel()
	require.Nil(t, pdu.GetTunInterface())
	_, err = netlink.LinkByName(device.Attrs().Name)
	require.Error(t, err)
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes, "session route and reservation must retire")
	rules, err := netlink.RuleList(netlink.FAMILY_V4)
	require.NoError(t, err)
	for _, r := range rules {
		require.NotEqual(t, table, r.Table, "source policy must retire")
	}
	for _, local := range []string{"10.88.0.1", "10.88.0.3"} {
		socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(local), Port: 2152})
		require.NoError(t, err)
		socket.Close()
	}
	SetupGtpInterface(ue, source)
	require.NotNil(t, pdu.GetTunInterface())
	exchange("production-reinstall")
	pdu.ReleaseTunnel()
	join()
}

func TestNativeEBPFServiceInitialOpenFailureRetiresStaging(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	join := testpeer.Start(t, "TestEBPFServicePeerProcess", "1")
	previous, previousTables := ebpfRegistry, sessionRoutingTables
	ebpfRegistry = ebpfgtp.NewRegistry()
	sessionRoutingTables = newRoutingTableAllocator(netlink.RouteAdd, netlink.RouteDel, routingTableInUse)
	t.Cleanup(func() { ebpfRegistry, sessionRoutingTables = previous, previousTables })
	occupied, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.1"), Port: 2152})
	require.NoError(t, err)
	defer occupied.Close()
	ue, pdu := sharedSetupUE(t, 1)
	ue.TunnelBackend = config.TunnelBackendEBPF
	pdu.SetIp([12]uint8{10, 60, 0, 1})
	source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
	// Real Registry.Open reaches the initial bind failure after staging the
	// owned TUN and policy. Its concrete nil must never enter cleanup's interface.
	err = setupEBPFTunnel(ue, pdu, source.GNBPduSessions[0], source.GnbIp)
	require.ErrorIs(t, err, syscall.EADDRINUSE, "preserve the original constructor error")
	require.Nil(t, pdu.GetTunInterface())
	require.NotContains(t, sessionRoutingTables.sessions, pdu)
	_, err = netlink.LinkByName("val" + ue.GetMsin())
	require.Error(t, err)
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: firstRoutingTable}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes)
	rules, err := netlink.RuleList(netlink.FAMILY_V4)
	require.NoError(t, err)
	for _, rule := range rules {
		require.NotEqual(t, firstRoutingTable, rule.Table)
	}
	require.NoError(t, occupied.Close())
	// Freed staging resources permit a complete fresh setup and actual traffic.
	SetupGtpInterface(ue, source)
	require.NotNil(t, pdu.GetTunInterface())
	defer pdu.ReleaseTunnel()
	require.Equal(t, firstRoutingTable, pdu.GetTunRoute().Table)
	app, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.60.0.1")}, &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 9000})
	require.NoError(t, err)
	defer app.Close()
	require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
	_, err = app.Write([]byte("after-initial-open-failure"))
	require.NoError(t, err)
	packet := make([]byte, 100)
	n, err := app.Read(packet)
	require.NoError(t, err)
	require.Equal(t, "after-initial-open-failure", string(packet[:n]))
	require.NoError(t, app.Close())
	pdu.ReleaseTunnel()
	join()
	port, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.1"), Port: 2152})
	require.NoError(t, err)
	require.NoError(t, port.Close())
}

type ebpfFailureSession struct {
	err    error
	closed int
}

func (s *ebpfFailureSession) Close() error                { s.closed++; return s.err }
func (s *ebpfFailureSession) Update(ebpfgtp.Config) error { return nil }
func (s *ebpfFailureSession) ProgramFD() int              { return 1 }

type ebpfFailurePort struct {
	err    error
	closed int
}

func (p *ebpfFailurePort) Read([]byte) (int, error)    { return 0, nil }
func (p *ebpfFailurePort) Write(b []byte) (int, error) { return len(b), nil }
func (p *ebpfFailurePort) Close() error                { p.closed++; return p.err }
func TestEBPFCleanupFailureRetainsEndpointAndSourcePolicy(t *testing.T) {
	for _, failure := range []string{"BPF", "TUN"} {
		t.Run(failure, func(t *testing.T) {
			session := &ebpfFailureSession{}
			port := &ebpfFailurePort{}
			if failure == "BPF" {
				session.err = errors.New("map deactivation failed")
			} else {
				port.err = errors.New("endpoint close failed")
			}
			_, pdu := sharedSetupUE(t, 1)
			allocator := newRoutingTableAllocator(func(*netlink.Route) error { return nil }, func(*netlink.Route) error { return nil }, func(uint32) (bool, error) { return false, nil })
			table, _, err := allocator.reserve(pdu)
			require.NoError(t, err)
			tunnel := &ebpfTunnel{session: session, port: port, table: table, rule: netlink.NewRule(), route: &netlink.Route{}}
			prevRule, prevRoute := ruleDel, routeDel
			deleted := 0
			ruleDel = func(*netlink.Rule) error { deleted++; return nil }
			routeDel = func(*netlink.Route) error { deleted++; return nil }
			t.Cleanup(func() { ruleDel = prevRule; routeDel = prevRoute; quarantinedEBPFEndpoints.Delete(tunnel) })
			tunnel.release()
			tunnel.release()
			require.Equal(t, 1, session.closed)
			require.Zero(t, deleted, "source policy must remain while the endpoint may still hold an address")
			_, retained := quarantinedEBPFEndpoints.Load(tunnel)
			require.True(t, retained, "keep a strong reference to the owned TUN descriptor")
			require.Empty(t, allocator.free, "quarantined routing table must not be reused")
			if failure == "BPF" {
				require.Zero(t, port.closed)
			} else {
				require.Equal(t, 1, port.closed)
			}
		})
	}
}
