// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

// Model the IPv4 kernel route key: preferred source and output device are not
// additional keys. Different default-route metrics can coexist in one table.
type testRouteKey struct{ table, metric int }
type routingTestKernel struct {
	mu                  sync.Mutex
	routes              map[testRouteKey]netlink.Route
	external            map[uint32]bool
	failAdd, failDelete bool
}

func newRoutingTestKernel() *routingTestKernel {
	return &routingTestKernel{routes: make(map[testRouteKey]netlink.Route), external: make(map[uint32]bool)}
}

func (k *routingTestKernel) add(route *netlink.Route) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.failAdd {
		return errors.New("route add failed")
	}
	key := testRouteKey{route.Table, route.Priority}
	if _, ok := k.routes[key]; ok {
		return syscall.EEXIST
	}
	k.routes[key] = *route
	return nil
}

func (k *routingTestKernel) replace(route *netlink.Route) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.routes[testRouteKey{route.Table, route.Priority}] = *route
	return nil
}

func (k *routingTestKernel) remove(route *netlink.Route) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.failDelete {
		return errors.New("route delete failed")
	}
	key := testRouteKey{route.Table, route.Priority}
	if _, ok := k.routes[key]; !ok {
		return syscall.ESRCH
	}
	delete(k.routes, key)
	return nil
}

func (k *routingTestKernel) inUse(table uint32) (bool, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.external[table] {
		return true, nil
	}
	for key := range k.routes {
		if key.table == int(table) {
			return true, nil
		}
	}
	return false, nil
}

func (k *routingTestKernel) route(table int) netlink.Route {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.routes[testRouteKey{table, 1}]
}

func mockSessionRoutingTables(t *testing.T) *routingTestKernel {
	t.Helper()
	old := sessionRoutingTables
	kernel := newRoutingTestKernel()
	sessionRoutingTables = newRoutingTableAllocator(kernel.add, kernel.remove, kernel.inUse)
	t.Cleanup(func() { sessionRoutingTables = old })
	return kernel
}

func TestRoutingReservationsCoordinateIndependentAllocators(t *testing.T) {
	kernel := newRoutingTestKernel()
	kernel.external[firstRoutingTable] = true
	// A stale snapshot makes both allocators attempt the same first available
	// table. RouteAdd must still exclude one claimant, as the kernel does.
	staleSnapshot := func(table uint32) (bool, error) { return table == firstRoutingTable, nil }
	allocators := []*routingTableAllocator{
		newRoutingTableAllocator(kernel.add, kernel.remove, staleSnapshot),
		newRoutingTableAllocator(kernel.add, kernel.remove, staleSnapshot),
	}
	const count = 64
	reservations := make([]*routingTableReservation, count)
	errorsFound := make([]error, count)
	var workers sync.WaitGroup
	for i := range reservations {
		workers.Add(1)
		go func() {
			defer workers.Done()
			reservations[i], _, errorsFound[i] = allocators[i%2].reserve(&context.UEPDUSession{})
		}()
	}
	workers.Wait()
	seen := make(map[uint32]bool)
	for i, held := range reservations {
		require.NoError(t, errorsFound[i])
		require.Greater(t, held.table, uint32(firstRoutingTable))
		require.False(t, seen[held.table], "two processes claimed table %d", held.table)
		seen[held.table] = true
	}
	for _, held := range reservations {
		held.release()
	}
	require.Empty(t, kernel.routes)
}

func TestRoutingReservationOwnershipAndFailures(t *testing.T) {
	kernel := newRoutingTestKernel()
	allocator := newRoutingTableAllocator(kernel.add, kernel.remove, kernel.inUse)
	session := &context.UEPDUSession{}
	first, owned, err := allocator.reserve(session)
	require.NoError(t, err)
	require.True(t, owned)
	retained, owned, err := allocator.reserve(session)
	require.NoError(t, err)
	require.False(t, owned)
	require.Same(t, first, retained)
	first.release()
	reused, _, err := allocator.reserve(session)
	require.NoError(t, err)
	require.Equal(t, first.table, reused.table)
	first.release() // A stale cleanup cannot release the new lifetime's claim.
	require.Same(t, reused, allocator.sessions[session])
	require.Len(t, kernel.routes, 1)
	kernel.failDelete = true
	reused.release()
	require.NotContains(t, allocator.sessions, session)
	kernel.failDelete = false
	fresh, _, err := allocator.reserve(session)
	require.NoError(t, err)
	require.NotEqual(t, reused.table, fresh.table, "failed claim cleanup must quarantine its table")
	fresh.release()
	kernel.failAdd = true
	_, _, err = allocator.reserve(&context.UEPDUSession{})
	require.Error(t, err)
	require.Empty(t, allocator.sessions)
	allocator.next, allocator.free = lastRoutingTable+1, nil
	_, _, err = allocator.reserve(session)
	require.ErrorContains(t, err, "no routing table identifiers")
}

// Run only inside an isolated network namespace. Setting the opt-in requires
// all kernel operations to succeed; missing privileges or VRF support fail.
func TestKernelRoutingTableReservations(t *testing.T) {
	if os.Getenv("PACKETRUSHER_KERNEL_TEST") != "1" {
		t.Skip("requires an isolated privileged network namespace")
	}
	require.Equal(t, "1", os.Getenv("PACKETRUSHER_ISOLATED_NETNS"), "refusing kernel writes without explicit isolated-netns guard")
	foreignRoute := &netlink.Route{Dst: &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, Table: firstRoutingTable, Type: syscall.RTN_BLACKHOLE}
	require.NoError(t, netlink.RouteAdd(foreignRoute))
	t.Cleanup(func() { _ = netlink.RouteDel(foreignRoute) })
	foreignRule := netlink.NewRule()
	foreignRule.Table, foreignRule.Priority = firstRoutingTable+1, 10
	foreignRule.Src = &net.IPNet{IP: net.IPv4(198, 51, 100, 1), Mask: net.CIDRMask(32, 32)}
	require.NoError(t, netlink.RuleAdd(foreignRule))
	t.Cleanup(func() { _ = netlink.RuleDel(foreignRule) })
	foreignVRF := &netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: "foreign-vrf"}, Table: firstRoutingTable + 2}
	require.NoError(t, netlink.LinkAdd(foreignVRF))
	t.Cleanup(func() { _ = netlink.LinkDel(foreignVRF) })
	for table := uint32(firstRoutingTable); table <= firstRoutingTable+2; table++ {
		occupied, err := routingTableInUse(table)
		require.NoError(t, err)
		require.True(t, occupied)
	}
	allocator := newRoutingTableAllocator(netlink.RouteAdd, netlink.RouteDel, routingTableInUse)
	session := &context.UEPDUSession{}
	held, _, err := allocator.reserve(session)
	require.NoError(t, err)
	require.Equal(t, uint32(firstRoutingTable+3), held.table)
	t.Cleanup(held.release)
	// The second allocator's snapshot is deliberately stale: the exclusive
	// kernel claim must protect the first allocator before its policy exists.
	competitor := newRoutingTableAllocator(netlink.RouteAdd, netlink.RouteDel, func(uint32) (bool, error) { return false, nil })
	competitor.next = held.table
	other, _, err := competitor.reserve(&context.UEPDUSession{})
	require.NoError(t, err)
	require.NotEqual(t, held.table, other.table)
	t.Cleanup(other.release)
	backend := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "target-backend"}}
	require.NoError(t, netlink.LinkAdd(backend))
	t.Cleanup(func() { _ = netlink.LinkDel(backend) })
	require.NoError(t, netlink.LinkSetUp(backend))
	address := &netlink.Addr{IPNet: &net.IPNet{IP: net.IPv4(10, 42, 0, 1), Mask: net.CIDRMask(32, 32)}}
	require.NoError(t, netlink.AddrAdd(backend, address))
	route := &netlink.Route{Dst: foreignRoute.Dst, Table: int(held.table), Priority: 1, Protocol: 4, Scope: netlink.SCOPE_LINK, LinkIndex: backend.Attrs().Index, Src: address.IP}
	require.NoError(t, netlink.RouteReplace(route))
	t.Cleanup(func() { _ = netlink.RouteDel(route) })
	policy := netlink.NewRule()
	policy.Src, policy.Table, policy.Priority = address.IPNet, int(held.table), 100
	require.NoError(t, netlink.RuleAdd(policy))
	t.Cleanup(func() { _ = netlink.RuleDel(policy) })
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: int(held.table)}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Len(t, routes, 2, "active route must coexist with its independent lifetime claim")
	lookup, err := netlink.RouteGetWithOptions(net.IPv4(203, 0, 113, 1), &netlink.RouteGetOptions{SrcAddr: address.IP})
	require.NoError(t, err)
	require.Equal(t, backend.Attrs().Index, lookup[0].LinkIndex, "claim must not capture active session traffic")
	require.NoError(t, netlink.RouteDel(route))
	require.NoError(t, netlink.RuleDel(policy))
	held.release()
	reused, _, err := allocator.reserve(session)
	require.NoError(t, err)
	require.Equal(t, held.table, reused.table)
	t.Cleanup(reused.release)
	held.release()
	occupied, err := routingTableInUse(reused.table)
	require.NoError(t, err)
	require.True(t, occupied, "old cleanup removed the new claim")
}
