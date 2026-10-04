// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func TestDedicatedReleaseWaitsForSocketWorker(t *testing.T) {
	previous := deleteDedicatedLink
	deleted := make(chan struct{})
	deleteDedicatedLink = func(string) error { close(deleted); return nil }
	t.Cleanup(func() { deleteDedicatedLink = previous })
	tunnel := &dedicatedTunnel{name: "gtp0123", stop: make(chan bool), done: make(chan struct{})}
	finished := make(chan struct{})
	go func() { tunnel.release(false); close(finished) }()
	<-tunnel.stop
	select {
	case <-deleted:
		t.Fatal("link deleted before worker finished")
	default:
	}
	close(tunnel.done)
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("release did not finish")
	}
}

func TestDedicatedHandoverKeepsUEEndpointAndPolicy(t *testing.T) {
	t.Run("policy", func(t *testing.T) { testDedicatedHandover(t, config.TunnelTun) })
	t.Run("VRF", func(t *testing.T) { testDedicatedHandover(t, config.TunnelVrf) })
}

func testDedicatedHandover(t *testing.T, mode config.TunnelMode) {
	kernel := mockSessionRoutingTables(t)
	oldAdd, oldDelete, oldFind := addDedicatedLink, deleteDedicatedLink, findTunnelLink
	oldPDR, oldFAR, oldQER := addTunnelPDR, addTunnelFAR, addTunnelQER
	oldAddr, oldRule, oldVRF := addTunnelAddress, addTunnelRule, addTunnelVRF
	oldMaster, oldUp, oldRoute, oldEndpoint, oldLinkDelete := setTunnelMaster, setTunnelUp, replaceTunnelRoute, makeUEEndpoint, deleteTunnelLink
	oldRuleDel, oldRouteDel := ruleDel, routeDel
	oldModPDR, oldModFAR, oldModQER := modifyTunnelPDR, modifyTunnelFAR, modifyTunnelQER
	oldMTU, oldEndpointMTU := setTunnelMTU, setUEEndpointMTU
	t.Cleanup(func() {
		addDedicatedLink, deleteDedicatedLink, findTunnelLink = oldAdd, oldDelete, oldFind
		addTunnelPDR, addTunnelFAR, addTunnelQER = oldPDR, oldFAR, oldQER
		addTunnelAddress, addTunnelRule, addTunnelVRF = oldAddr, oldRule, oldVRF
		setTunnelMaster, setTunnelUp, replaceTunnelRoute, makeUEEndpoint, deleteTunnelLink = oldMaster, oldUp, oldRoute, oldEndpoint, oldLinkDelete
		ruleDel, routeDel = oldRuleDel, oldRouteDel
		modifyTunnelPDR, modifyTunnelFAR, modifyTunnelQER = oldModPDR, oldModFAR, oldModQER
		setTunnelMTU, setUEEndpointMTU = oldMTU, oldEndpointMTU
	})
	setTunnelMTU = func(link netlink.Link, ip netip.Addr, _ int) error {
		link.Attrs().MTU = 1556 - 100*int(ip.As4()[3])
		return nil
	}
	setUEEndpointMTU = func(netlink.Link, int) error { return nil }
	var mu sync.Mutex
	links := map[string]netlink.Link{}
	var events []string
	nextIndex := 10
	addDedicatedLink = func(name string, _, _ int, _, _ string, stop chan bool) error {
		mu.Lock()
		nextIndex++
		links[name] = &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name, Index: nextIndex}}
		events = append(events, "create:"+name)
		mu.Unlock()
		<-stop
		mu.Lock()
		events = append(events, "stop:"+name)
		mu.Unlock()
		return nil
	}
	findTunnelLink = func(name string) (netlink.Link, error) {
		mu.Lock()
		defer mu.Unlock()
		if link := links[name]; link != nil {
			return link, nil
		}
		return nil, errors.New("not found")
	}
	deleteDedicatedLink = func(name string) error {
		mu.Lock()
		defer mu.Unlock()
		delete(links, name)
		events = append(events, "delete:"+name)
		return nil
	}
	deleteTunnelLink = func(link netlink.Link) error {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, "delete:"+link.Attrs().Name)
		return nil
	}
	addTunnelPDR = func([]string) error { return nil }
	addTunnelFAR = func([]string) error { return nil }
	addTunnelQER = func([]string) error { return nil }
	addTunnelAddress = func(netlink.Link, *netlink.Addr) error { return nil }
	ruleAdds := 0
	addTunnelRule = func(*netlink.Rule) error { ruleAdds++; return nil }
	ruleDel = func(*netlink.Rule) error { mu.Lock(); events = append(events, "delete-rule"); mu.Unlock(); return nil }
	routeDel = func(*netlink.Route) error {
		mu.Lock()
		events = append(events, "delete-route")
		mu.Unlock()
		return nil
	}
	vrfCreates := 0
	addTunnelVRF = func(netlink.Link) error { vrfCreates++; return nil }
	updates := 0
	modifyTunnelPDR = func([]string) error { updates++; return nil }
	modifyTunnelFAR = func([]string) error { updates++; return nil }
	modifyTunnelQER = func([]string) error { updates++; return nil }
	setTunnelMaster = func(netlink.Link, netlink.Link) error { return nil }
	setTunnelUp = func(netlink.Link) error { return nil }
	makeUEEndpoint = func(name string) (netlink.Link, error) {
		return &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name, Index: 99}}, nil
	}
	var routeError error
	replaceTunnelRoute = func(route *netlink.Route) error {
		if routeError != nil {
			return routeError
		}
		mu.Lock()
		events = append(events, "route")
		mu.Unlock()
		return nil
	}
	ue := &context.UEContext{TunnelMode: mode, TunnelBackend: config.TunnelBackendKernel}
	ue.UeSecurity.Msin = "7005551000"
	session := &context.UEPDUSession{Id: 1}
	session.SetIp([12]uint8{10, 1, 0, 1})
	ue.PduSession[0] = session
	build := func(ip string, teid uint32) gnbContext.UEMessage {
		ctx := &gnbContext.GNBUe{}
		ctx.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
		pdu, err := ctx.CreatePduSession(1, "192.0.2.100", "01", "000000", 0, 1, 8, 9, teid, teid+1)
		require.NoError(t, err)
		var sessions [16]*gnbContext.GnbPDUSession
		sessions[0] = pdu
		return gnbContext.UEMessage{GNBPduSessions: sessions, GnbIp: netip.MustParseAddr(ip)}
	}
	SetupGtpInterface(ue, build("192.0.2.1", 10))
	endpoint, policy, table, vrf := session.GetUEInterface(), session.GetTunRule(), session.GetTunRoute().Table, session.GetVrfDevice()
	require.Equal(t, 1456, endpoint.Attrs().MTU)
	require.Equal(t, endpoint.Attrs().MTU, session.GetTunInterface().Attrs().MTU)
	reservation := sessionRoutingTables.sessions[session]
	if mode == config.TunnelTun {
		require.Equal(t, table, policy.Table)
	} else {
		require.Equal(t, uint32(table), vrf.Table)
	}
	SetupGtpInterface(ue, build("192.0.2.2", 20))
	require.Same(t, reservation, sessionRoutingTables.sessions[session])
	require.Same(t, endpoint, session.GetUEInterface())
	require.Equal(t, 1356, endpoint.Attrs().MTU)
	require.Equal(t, endpoint.Attrs().MTU, session.GetTunInterface().Attrs().MTU)
	if mode == config.TunnelTun {
		require.Same(t, policy, session.GetTunRule())
	} else {
		require.Same(t, vrf, session.GetVrfDevice())
	}
	require.Equal(t, table, session.GetTunRoute().Table)
	if mode == config.TunnelTun {
		require.Equal(t, 1, ruleAdds)
	} else {
		require.Equal(t, 1, vrfCreates)
	}
	mu.Lock()
	handoverEvents := append([]string(nil), events...)
	mu.Unlock()
	require.Equal(t, []string{"create:gtp07005551000", "route", "create:gtp17005551000", "route", "stop:gtp07005551000", "delete:gtp07005551000"}, handoverEvents)
	// Repeated setup on the same N3 endpoint updates rules without rebinding.
	sameN3Backend := session.GetTunInterface()
	SetupGtpInterface(ue, build("192.0.2.2", 21))
	require.Same(t, sameN3Backend, session.GetTunInterface())
	require.Equal(t, 4, updates)
	// A target that fails staging must leave the completed source usable.
	source := session.GetTunInterface()
	addTunnelFAR = func([]string) error { return errors.New("target FAR failed") }
	SetupGtpInterface(ue, build("192.0.2.3", 30))
	require.Same(t, source, session.GetTunInterface())
	require.Same(t, endpoint, session.GetUEInterface())
	require.Equal(t, netip.MustParseAddr("192.0.2.2"), session.GetGnbIp())
	mu.Lock()
	failureEvents := append([]string(nil), events...)
	mu.Unlock()
	require.NotContains(t, failureEvents, "delete:val7005551000")
	require.NotContains(t, failureEvents, "delete-rule")
	// A failure after reusing the reservation must not return the source table.
	addTunnelFAR = func([]string) error { return nil }
	routeError = errors.New("target route failed")
	SetupGtpInterface(ue, build("192.0.2.3", 30))
	require.Same(t, source, session.GetTunInterface())
	require.Same(t, reservation, sessionRoutingTables.sessions[session])
	require.Equal(t, table, session.GetTunRoute().Table)
	require.Len(t, kernel.routes, 1, "target failure removed or duplicated the source claim")
	require.Equal(t, 1356, endpoint.Attrs().MTU, "failed route commit must restore source endpoint MTU")
	require.Equal(t, 1356, source.Attrs().MTU, "source backend must remain unchanged")
	// An endpoint update failure must not replace the source route or tunnel.
	routeError = nil
	setUEEndpointMTU = func(_ netlink.Link, mtu int) error {
		if mtu != 1356 {
			return errors.New("target endpoint MTU failed")
		}
		return nil
	}
	SetupGtpInterface(ue, build("192.0.2.3", 31))
	require.Same(t, source, session.GetTunInterface())
	require.Equal(t, 1356, endpoint.Attrs().MTU)
	require.Equal(t, netip.MustParseAddr("192.0.2.2"), session.GetGnbIp())
	require.Len(t, kernel.routes, 1)
	setUEEndpointMTU = func(netlink.Link, int) error { return nil }
	routeError = errors.New("target route failed")
	// Failed initial commit does own its reservation and must return it, along
	// with its newly created policy/VRF and endpoint.
	freshUE := &context.UEContext{TunnelMode: mode, TunnelBackend: config.TunnelBackendKernel}
	freshUE.UeSecurity.Msin = "7005551001"
	fresh := &context.UEPDUSession{Id: 1}
	fresh.SetIp([12]uint8{10, 1, 0, 2})
	freshUE.PduSession[0] = fresh
	t.Cleanup(fresh.ReleaseTunnel)
	SetupGtpInterface(freshUE, build("192.0.2.4", 10))
	require.Nil(t, fresh.GetTunInterface())
	require.NotContains(t, sessionRoutingTables.sessions, fresh)
	require.Len(t, kernel.routes, 1, "failed initial commit leaked its claim")
	routeError = nil
	SetupGtpInterface(freshUE, build("192.0.2.4", 10))
	require.NotEqual(t, table, fresh.GetTunRoute().Table, "recycled TEID collided with the retained table")
	require.Equal(t, 1156, fresh.GetUEInterface().Attrs().MTU)
	require.Equal(t, fresh.GetUEInterface().Attrs().MTU, fresh.GetTunInterface().Attrs().MTU)
	fresh.ReleaseTunnel()
	// If restoring a retained endpoint fails, the completed source is no longer
	// consistent. Exercise the production caller's retirement, not just the helper.
	routeError = errors.New("target route failed")
	restores := 0
	setUEEndpointMTU = func(actual netlink.Link, mtu int) error {
		require.Same(t, endpoint, actual)
		if mtu == 1356 {
			restores++
			return errors.New("source endpoint MTU restoration failed")
		}
		return nil
	}
	SetupGtpInterface(ue, build("192.0.2.3", 32))
	require.Equal(t, 1, restores)
	require.Nil(t, session.GetTunInterface(), "failed source MTU rollback must release the tunnel")
	require.Nil(t, session.GetUEInterface())
	require.Nil(t, session.GetTunRule())
	require.Nil(t, session.GetTunRoute())
	require.Nil(t, session.GetVrfDevice())
	session.ReleaseTunnel()
	require.Empty(t, sessionRoutingTables.sessions)
	require.Empty(t, kernel.routes)
	mu.Lock()
	finalEvents := append([]string(nil), events...)
	mu.Unlock()
	require.Contains(t, finalEvents, "delete:val7005551000")
	require.Contains(t, finalEvents, "stop:gtp17005551000")
	endpointDeletes := 0
	for _, event := range finalEvents {
		if event == "delete:val7005551000" {
			endpointDeletes++
		}
	}
	require.Equal(t, 1, endpointDeletes, "source endpoint cleanup must be idempotent")
	require.Nil(t, session.GetTunInterface())
}

func TestSameN3UpdateRestoresRulesAfterPartialFailure(t *testing.T) {
	oldPDR, oldFAR, oldQER := modifyTunnelPDR, modifyTunnelFAR, modifyTunnelQER
	t.Cleanup(func() { modifyTunnelPDR, modifyTunnelFAR, modifyTunnelQER = oldPDR, oldFAR, oldQER })
	build := func(teid uint32, qfi int64) *gnbContext.GnbPDUSession {
		ue := &gnbContext.GNBUe{}
		ue.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
		pdu, err := ue.CreatePduSession(1, "192.0.2.100", "01", "000000", 0, qfi, 8, 9, teid, teid+1)
		require.NoError(t, err)
		return pdu
	}
	ip := netip.MustParseAddr("192.0.2.1")
	previous := dedicatedRuleSet("gtp0123", build(10, 1), "10.1.0.1", ip)
	for failureAt := 1; failureAt <= 4; failureAt++ {
		t.Run(fmt.Sprintf("operation-%d", failureAt), func(t *testing.T) {
			state := map[string][]string{"FAR2": previous.farUp, "PDR1": previous.pdrDown, "QER1": previous.qer, "PDR2": previous.pdrUp}
			wanted := maps.Clone(state)
			calls := 0
			apply := func(kind string, args []string) error {
				calls++
				state[kind+args[1]] = append([]string(nil), args...)
				if calls == failureAt {
					return errors.New("kernel acknowledged partial failure")
				}
				return nil
			}
			modifyTunnelFAR = func(args []string) error { return apply("FAR", args) }
			modifyTunnelPDR = func(args []string) error { return apply("PDR", args) }
			modifyTunnelQER = func(args []string) error { return apply("QER", args) }
			tunnel := &dedicatedTunnel{name: "gtp0123", activeRules: previous}
			err := tunnel.refresh(build(20, 2), "10.1.0.1", ip)
			require.Error(t, err)
			require.False(t, errors.Is(err, errTunnelRollback))
			require.Equal(t, wanted, state, "partially applied rules were not restored")
			require.Equal(t, previous, tunnel.activeRules)
		})
	}
	modifyTunnelFAR = func([]string) error { return errors.New("netlink unavailable") }
	modifyTunnelPDR = func([]string) error { return nil }
	modifyTunnelQER = func([]string) error { return nil }
	tunnel := &dedicatedTunnel{name: "gtp0123", activeRules: previous}
	require.ErrorIs(t, tunnel.refresh(build(20, 2), "10.1.0.1", ip), errTunnelRollback)
	calls := 0
	modifyTunnelFAR = func([]string) error { calls++; return nil }
	require.Error(t, tunnel.refresh(build(20, 0), "10.1.0.1", ip))
	require.Zero(t, calls, "unsupported QFI transition changed rules")
}
