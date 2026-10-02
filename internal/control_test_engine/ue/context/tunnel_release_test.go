/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"my5G-RANTester/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

// countDeletes replaces the rule and route removals with counters for one test.
func countDeletes(t *testing.T) (rules, routes *int) {
	rules, routes = new(int), new(int)
	oldRule, oldRoute := ruleDel, routeDel
	ruleDel = func(*netlink.Rule) error { *rules++; return nil }
	routeDel = func(*netlink.Route) error { *routes++; return nil }
	t.Cleanup(func() { ruleDel, routeDel = oldRule, oldRoute })

	return rules, routes
}

// A shared tunnel is given back once: a second release would return its rule slot to
// the device twice, and two later UEs would be handed the same identifiers.
func TestReleaseTunnelRunsOnce(t *testing.T) {
	pduSession := &UEPDUSession{}
	calls := 0
	pduSession.SetTunnelRelease(func() { calls++ })

	pduSession.ReleaseTunnel()
	pduSession.ReleaseTunnel()

	assert.Equal(t, 1, calls)
}

// A session the network releases is removed from the UE before Terminate runs, so
// Terminate never sees it: its tunnel has to be given back here.
func TestDeletePduSessionReleasesSharedTunnel(t *testing.T) {
	ue := &UEContext{TunnelMode: config.TunnelShared}
	pduSession := &UEPDUSession{Wait: make(chan bool)}
	ue.PduSession[0] = pduSession
	released := false
	pduSession.SetTunnelRelease(func() { released = true })

	require.NoError(t, ue.DeletePduSession(1))

	assert.True(t, released)
	assert.Nil(t, ue.PduSession[0])
}

// Terminate releases a shared tunnel before it reads the session's rule and route,
// since the release removes them itself. Reading them first removes them twice.
func TestTerminateReleasesSharedTunnelBeforeReadingIt(t *testing.T) {
	rules, routes := countDeletes(t)

	ue := newTestUE()
	ue.TunnelMode = config.TunnelShared
	pduSession := &UEPDUSession{}
	pduSession.SetTunRule(&netlink.Rule{})
	pduSession.SetTunRoute(&netlink.Route{})
	released := 0
	pduSession.SetTunnelRelease(func() {
		// As the shared tunnel's release does: it removes both, then clears them.
		released++
		_ = ruleDel(pduSession.GetTunRule())
		_ = routeDel(pduSession.GetTunRoute())
		pduSession.SetTunRule(nil)
		pduSession.SetTunRoute(nil)
	})
	ue.PduSession[0] = pduSession

	ue.Terminate()

	assert.Equal(t, 1, released)
	assert.Equal(t, 1, *rules, "the rule is removed once, by the release")
	assert.Equal(t, 1, *routes, "the route is removed once, by the release")
}

// A UE that owns its device keeps removing its rule and route itself.
func TestTerminateRemovesDedicatedRuleAndRoute(t *testing.T) {
	rules, routes := countDeletes(t)

	ue := newTestUE()
	ue.TunnelMode = config.TunnelTun
	pduSession := &UEPDUSession{}
	pduSession.SetTunRule(&netlink.Rule{})
	pduSession.SetTunRoute(&netlink.Route{})
	ue.PduSession[0] = pduSession

	ue.Terminate()

	assert.Equal(t, 1, *rules)
	assert.Equal(t, 1, *routes)
}
