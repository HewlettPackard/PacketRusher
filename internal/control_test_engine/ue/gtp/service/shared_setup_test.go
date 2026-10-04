// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

type setupTestSharedDevice struct {
	gtp.Device // Exercise the production per-device rule-ID allocator.
	name       string
	mu         sync.Mutex
	released   []gtp.RuleIDs
}

func (d *setupTestSharedDevice) Name() string          { return d.name }
func (d *setupTestSharedDevice) AddPDR([]string) error { return nil }
func (d *setupTestSharedDevice) AddFAR([]string) error { return nil }
func (d *setupTestSharedDevice) AddQER([]string) error { return nil }
func (d *setupTestSharedDevice) RemoveAddress(string)  {}
func (d *setupTestSharedDevice) Release(ids gtp.RuleIDs) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.released = append(d.released, ids)
}

type sharedSetupFixture struct {
	kernel                *routingTestKernel
	allocator             *routingTableAllocator
	failPolicy, failRoute bool
}

func newSharedSetupFixture(t *testing.T) *sharedSetupFixture {
	t.Helper()
	fixture := &sharedSetupFixture{kernel: mockSessionRoutingTables(t), allocator: sessionRoutingTables}
	source, target := &setupTestSharedDevice{name: "shared-source"}, &setupTestSharedDevice{name: "shared-target"}
	links := map[string]netlink.Link{
		"shared-source": &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "shared-source", Index: 10}},
		"shared-target": &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "shared-target", Index: 11}},
	}
	oldDevice, oldFind, oldAddr := sharedDeviceFor, findTunnelLink, addTunnelAddress
	oldRule, oldRoute, oldRuleDel, oldRouteDel := addTunnelRule, replaceTunnelRoute, ruleDel, routeDel
	t.Cleanup(func() {
		sharedDeviceFor, findTunnelLink, addTunnelAddress = oldDevice, oldFind, oldAddr
		addTunnelRule, replaceTunnelRoute, ruleDel, routeDel = oldRule, oldRoute, oldRuleDel, oldRouteDel
	})
	sharedDeviceFor = func(msg gnbContext.UEMessage) sharedGTPDevice {
		if msg.GnbIp == netip.MustParseAddr("192.0.2.1") {
			return source
		}
		return target
	}
	findTunnelLink = func(name string) (netlink.Link, error) {
		if link := links[name]; link != nil {
			return link, nil
		}
		return nil, errors.New("unknown link")
	}
	addTunnelAddress = func(netlink.Link, *netlink.Addr) error { return nil }
	addTunnelRule = func(*netlink.Rule) error {
		if fixture.failPolicy {
			return errors.New("policy creation failed")
		}
		return nil
	}
	ruleDel = func(*netlink.Rule) error { return nil }
	routeDel = fixture.kernel.remove
	replaceTunnelRoute = func(route *netlink.Route) error {
		if fixture.failRoute {
			return errors.New("route replacement failed")
		}
		return fixture.kernel.replace(route)
	}
	return fixture
}

func sharedSetupUE(t *testing.T, number int) (*context.UEContext, *context.UEPDUSession) {
	t.Helper()
	ue := &context.UEContext{TunnelMode: config.TunnelShared, TunnelBackend: config.TunnelBackendKernel}
	ue.UeSecurity.Msin = fmt.Sprintf("700555%04d", number)
	session := &context.UEPDUSession{Id: 1}
	session.SetIp([12]uint8{10, 42, byte(number / 256), byte(number % 256)})
	ue.PduSession[0] = session
	t.Cleanup(session.ReleaseTunnel)
	return ue, session
}

func sharedSetupMessage(t *testing.T, ip string, teid uint32) gnbContext.UEMessage {
	t.Helper()
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, "192.0.2.100", "01", "000000", 0, 0, 8, 9, teid, teid+1)
	require.NoError(t, err)
	var sessions [16]*gnbContext.GnbPDUSession
	sessions[0] = pdu
	return gnbContext.UEMessage{GNBPduSessions: sessions, GnbIp: netip.MustParseAddr(ip)}
}

func TestSharedSetupRetainsTableWhenSourceTEIDIsRecycled(t *testing.T) {
	fixture := newSharedSetupFixture(t)
	a, ap := sharedSetupUE(t, 1)
	b, bp := sharedSetupUE(t, 2)
	SetupGtpInterface(a, sharedSetupMessage(t, "192.0.2.1", 10))
	require.NotNil(t, ap.GetTunRoute())
	table, policy := ap.GetTunRoute().Table, ap.GetTunRule()
	reservation := fixture.allocator.sessions[ap]
	SetupGtpInterface(a, sharedSetupMessage(t, "192.0.2.2", 20))
	require.Same(t, policy, ap.GetTunRule())
	require.Same(t, reservation, fixture.allocator.sessions[ap])
	require.Equal(t, table, ap.GetTunRoute().Table)
	require.Equal(t, 11, fixture.kernel.route(table).LinkIndex)
	SetupGtpInterface(b, sharedSetupMessage(t, "192.0.2.1", 10))
	require.NotNil(t, bp.GetTunRoute())
	require.NotEqual(t, table, bp.GetTunRule().Table)
	require.Equal(t, 11, fixture.kernel.route(table).LinkIndex, "recycled TEID overwrote the handed-over session's route")
	require.Equal(t, 10, fixture.kernel.route(bp.GetTunRoute().Table).LinkIndex)
	// Failed target commit leaves A's original reservation and source route.
	fixture.failRoute = true
	SetupGtpInterface(a, sharedSetupMessage(t, "192.0.2.1", 30))
	fixture.failRoute = false
	require.Same(t, reservation, fixture.allocator.sessions[ap])
	require.Same(t, policy, ap.GetTunRule())
	require.Equal(t, 11, fixture.kernel.route(table).LinkIndex)
	require.Equal(t, netip.MustParseAddr("192.0.2.2"), ap.GetGnbIp())
	ap.ReleaseTunnel()
	require.NotContains(t, fixture.allocator.sessions, ap)
	require.Equal(t, 10, fixture.kernel.route(bp.GetTunRoute().Table).LinkIndex)
	// Reusing a session object starts a new lifetime, and delayed old cleanup is
	// powerless against its newly acquired reservation.
	SetupGtpInterface(a, sharedSetupMessage(t, "192.0.2.1", 10))
	require.Equal(t, table, ap.GetTunRoute().Table)
	current := fixture.allocator.sessions[ap]
	require.NotSame(t, reservation, current)
	reservation.release()
	require.Same(t, current, fixture.allocator.sessions[ap])
	require.Equal(t, 10, fixture.kernel.route(table).LinkIndex)
}

func TestSharedSetupFailureReturnsOnlyOwnedTable(t *testing.T) {
	for _, stage := range []string{"policy", "route"} {
		t.Run(stage, func(t *testing.T) {
			fixture := newSharedSetupFixture(t)
			ue, session := sharedSetupUE(t, 1)
			fixture.failPolicy, fixture.failRoute = stage == "policy", stage == "route"
			SetupGtpInterface(ue, sharedSetupMessage(t, "192.0.2.1", 10))
			require.Nil(t, session.GetTunInterface())
			require.Empty(t, fixture.allocator.sessions)
			require.Empty(t, fixture.kernel.routes, "failed initial setup left a table claim")
			fixture.failPolicy, fixture.failRoute = false, false
			SetupGtpInterface(ue, sharedSetupMessage(t, "192.0.2.1", 10))
			require.Equal(t, firstRoutingTable, session.GetTunRoute().Table)
		})
	}
}

func TestSharedSetupConcurrentSessionsHaveUniqueTables(t *testing.T) {
	fixture := newSharedSetupFixture(t)
	const count = 64
	ues := make([]*context.UEContext, count)
	sessions := make([]*context.UEPDUSession, count)
	messages := make([]gnbContext.UEMessage, count)
	for i := range ues {
		ues[i], sessions[i] = sharedSetupUE(t, i+1)
		// The same UL TEID on different UPFs is legal; host table identities
		// cannot rely on the TEID being globally unique.
		ip := "192.0.2.1"
		if i%2 != 0 {
			ip = "192.0.2.2"
		}
		messages[i] = sharedSetupMessage(t, ip, 10)
	}
	var workers sync.WaitGroup
	for i := range ues {
		workers.Add(1)
		go func() { defer workers.Done(); SetupGtpInterface(ues[i], messages[i]) }()
	}
	workers.Wait()
	seen := make(map[int]bool)
	for _, session := range sessions {
		require.NotNil(t, session.GetTunRoute())
		table := session.GetTunRoute().Table
		require.False(t, seen[table], "two sessions selected table %d", table)
		seen[table] = true
		require.Equal(t, session.GetTunRoute().LinkIndex, fixture.kernel.route(table).LinkIndex)
	}
	for _, session := range sessions {
		session.ReleaseTunnel()
	}
	require.Empty(t, fixture.kernel.routes)
	require.Empty(t, fixture.allocator.sessions)
}

func TestSharedFinalCleanupFailureQuarantinesTableOnSessionReuse(t *testing.T) {
	fixture := newSharedSetupFixture(t)
	ue, session := sharedSetupUE(t, 1)
	SetupGtpInterface(ue, sharedSetupMessage(t, "192.0.2.1", 10))
	table := session.GetTunRoute().Table
	fixture.kernel.failDelete = true
	session.ReleaseTunnel()
	fixture.kernel.failDelete = false
	require.NotContains(t, fixture.allocator.sessions, session)
	SetupGtpInterface(ue, sharedSetupMessage(t, "192.0.2.1", 10))
	require.NotEqual(t, table, session.GetTunRoute().Table)
	require.Equal(t, 10, fixture.kernel.route(table).LinkIndex, "orphaned routing must not be overwritten")
}
