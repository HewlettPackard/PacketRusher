// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

// Routing tables identify a PDU session, not its reusable UPF TEID. Keep their
// identifiers away from Linux's reserved tables and within the signed int range
// accepted by netlink on both 32-bit and 64-bit hosts.
const (
	firstRoutingTable         = 1000
	lastRoutingTable          = 1<<31 - 1
	routingTableClaimPriority = 1<<31 - 1
)

var sessionRoutingTables = newRoutingTableAllocator(netlink.RouteAdd, netlink.RouteDel, routingTableInUse)

type routingTableAllocator struct {
	mu       sync.Mutex
	next     uint32
	free     []uint32
	sessions map[*context.UEPDUSession]*routingTableReservation
	claim    func(*netlink.Route) error
	unclaim  func(*netlink.Route) error
	inUse    func(uint32) (bool, error)
}

type routingTableReservation struct {
	table     uint32
	session   *context.UEPDUSession
	allocator *routingTableAllocator
	claim     *netlink.Route
}

func newRoutingTableAllocator(claim, unclaim func(*netlink.Route) error, inUse func(uint32) (bool, error)) *routingTableAllocator {
	return &routingTableAllocator{next: firstRoutingTable, sessions: make(map[*context.UEPDUSession]*routingTableReservation), claim: claim, unclaim: unclaim, inUse: inUse}
}

// reserve returns an existing session reservation during handover. A new
// reservation belongs to staging until commit; failed staging must return only
// its own reservation, after removing any policy/VRF and route it installed.
func (a *routingTableAllocator) reserve(session *context.UEPDUSession) (*routingTableReservation, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if held := a.sessions[session]; held != nil {
		return held, false, nil
	}
	for {
		var table uint32
		if last := len(a.free) - 1; last >= 0 {
			table, a.free = a.free[last], a.free[:last]
		} else {
			if a.next > lastRoutingTable {
				return nil, false, errors.New("no routing table identifiers left")
			}
			table = a.next
			a.next++
		}
		occupied, err := a.inUse(table)
		if err != nil || occupied {
			if err != nil {
				return nil, false, fmt.Errorf("inspect routing table %d: %w", table, err)
			}
			continue // Do not replace unrelated routes, policies or VRFs.
		}
		claim := &netlink.Route{
			Dst:   &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			Table: int(table), Priority: routingTableClaimPriority,
			Protocol: 4, Type: syscall.RTN_BLACKHOLE, Scope: netlink.SCOPE_UNIVERSE,
		}
		if err := a.claim(claim); errors.Is(err, syscall.EEXIST) {
			continue // A concurrent allocator claimed it after our inspection.
		} else if err != nil {
			return nil, false, fmt.Errorf("reserve routing table %d: %w", table, err)
		}
		held := &routingTableReservation{table: table, session: session, allocator: a, claim: claim}
		a.sessions[session] = held
		return held, true, nil
	}
}

// release is called only after owned routing objects are gone. The identity
// check makes an old cleanup harmless if the same session object is reused.
func (r *routingTableReservation) release() {
	if r == nil {
		return
	}
	a := r.allocator
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sessions[r.session] != r {
		return
	}
	if err := a.unclaim(r.claim); !routingObjectRemoved(err) {
		delete(a.sessions, r.session)
		log.Warn("[UE][GTP] Keeping routing table ", r.table, " reservation after claim cleanup failed: ", err)
		return
	}
	delete(a.sessions, r.session)
	a.free = append(a.free, r.table)
}

// Incomplete final cleanup abandons the old binding but leaves its kernel claim
// reserved. Reusing the session object must allocate a fresh table rather than
// inherit orphaned routing policy from its previous lifetime.
func (r *routingTableReservation) quarantine() {
	if r == nil {
		return
	}
	a := r.allocator
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sessions[r.session] == r {
		delete(a.sessions, r.session)
	}
}

// A blackhole default at a distinct, lowest-preference metric remains in the
// actual network namespace throughout staging, handover and final cleanup.
// RouteAdd uses NLM_F_EXCL, so competing PacketRusher processes (including host-
// network containers with separate filesystems) cannot claim the same table.
// Existing routes, policies and VRFs are excluded first. Administrators must not
// deliberately replace a live session's objects; a process crash leaves its
// claim behind, preventing another process from overwriting orphaned routing.
func routingTableInUse(table uint32) (bool, error) {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: int(table)}, netlink.RT_FILTER_TABLE)
	if err != nil || len(routes) != 0 {
		return len(routes) != 0, err
	}
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return false, err
	}
	for _, rule := range rules {
		if rule.Table == int(table) {
			return true, nil
		}
	}
	links, err := netlink.LinkList()
	if err != nil {
		return false, err
	}
	for _, link := range links {
		if vrf, ok := link.(*netlink.Vrf); ok && vrf.Table == table {
			return true, nil
		}
	}
	return false, nil
}

func routingObjectRemoved(err error) bool {
	return err == nil || errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.ESRCH) || errors.Is(err, syscall.ENODEV)
}
