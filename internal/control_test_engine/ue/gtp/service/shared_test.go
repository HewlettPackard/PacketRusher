/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package service

import (
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestSetupConcurrency(t *testing.T) {
	for value, want := range map[string]int{"": 32, "8": 8, "0": 32, "-1": 32, "many": 32} {
		t.Setenv("PR_SETUP_SLOTS", value)
		assert.Equal(t, want, setupConcurrency(), "PR_SETUP_SLOTS=%q", value)
	}
}

type fakeDevice struct {
	calls []string
	ueIP  string
	ids   gtp.RuleIDs
}

func (f *fakeDevice) RemoveAddress(ueIP string) {
	f.calls = append(f.calls, "address")
	f.ueIP = ueIP
}

func (f *fakeDevice) Release(ids gtp.RuleIDs) {
	f.calls = append(f.calls, "rules")
	f.ids = ids
}

// Releasing a UE's shared tunnel removes everything setup recorded -- route, then policy
// rule, then address and GTP-U rules -- and leaves the session holding none of it.
func TestSharedTunnelReleaseRemovesWhatSetupInstalled(t *testing.T) {
	var order []string
	oldRule, oldRoute := ruleDel, routeDel
	ruleDel = func(*netlink.Rule) error { order = append(order, "policy rule"); return nil }
	routeDel = func(*netlink.Route) error { order = append(order, "route"); return nil }
	t.Cleanup(func() { ruleDel, routeDel = oldRule, oldRoute })

	dev := &fakeDevice{}
	ids := gtp.RuleIDs{Slot: 4, PDRDown: 9, PDRUp: 10, FARDown: 9, FARUp: 10}
	rule, route := &netlink.Rule{}, &netlink.Route{}
	held := &sharedTunnel{dev: dev, ueIP: "10.60.0.7", ids: ids, rule: rule, route: route}

	pduSession := &context.UEPDUSession{}
	pduSession.SetTunRule(rule)
	pduSession.SetTunRoute(route)
	pduSession.SetTunInterface(&netlink.Dummy{})

	held.release(pduSession)

	assert.Equal(t, []string{"route", "policy rule"}, order)
	assert.Equal(t, []string{"address", "rules"}, dev.calls)
	assert.Equal(t, "10.60.0.7", dev.ueIP)
	assert.Equal(t, ids, dev.ids)
	assert.Nil(t, pduSession.GetTunRule())
	assert.Nil(t, pduSession.GetTunRoute())
	assert.Nil(t, pduSession.GetTunInterface())
}

// A setup that fails before its policy rule or route exists gives back only what it got.
func TestSharedTunnelReleaseAfterEarlyFailure(t *testing.T) {
	deletes := 0
	oldRule, oldRoute := ruleDel, routeDel
	ruleDel = func(*netlink.Rule) error { deletes++; return nil }
	routeDel = func(*netlink.Route) error { deletes++; return nil }
	t.Cleanup(func() { ruleDel, routeDel = oldRule, oldRoute })

	dev := &fakeDevice{}
	held := &sharedTunnel{dev: dev, ueIP: "10.60.0.7", ids: gtp.RuleIDs{Slot: 1}}

	held.release(&context.UEPDUSession{})

	assert.Zero(t, deletes)
	assert.Equal(t, []string{"address", "rules"}, dev.calls)
}
