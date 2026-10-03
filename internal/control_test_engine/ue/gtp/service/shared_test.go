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
	"github.com/stretchr/testify/require"
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

// With PR_VERIFY_RULES=1 and no client to read the rule back with, a failed create must
// still be reported. The check used to answer "installed" whenever it could not check,
// and the helper returned nil for a PDR that was never created. Here the create goes
// through the per-call fallback and fails: with ENODEV for the missing interface when
// gtp5g is loaded, or earlier, looking up the gtp5g netlink family, when it is not.
func TestAddPDRVerifiedKeepsCreateErrorWhenItCannotCheck(t *testing.T) {
	t.Setenv("PR_VERIFY_RULES", "1")

	dev := &gtp.Device{}
	_, checked := dev.PDRInstalled(2)
	require.False(t, checked, "a device without a client cannot check")

	err := addPDRVerified(dev, 2, []string{"nosuchgtp0", "2", "--pcd", "2", "--ue-ipv4", "10.0.0.1", "--far-id", "2"}, "uplink")
	assert.Error(t, err, "the failed create should be reported")
}

func TestSharedHandoverTransfersPolicyAndPreservesTargetAddress(t *testing.T) {
	oldRule, oldRoute := ruleDel, routeDel
	ruleDel = func(*netlink.Rule) error { t.Fatal("retirement deleted transferred policy"); return nil }
	routeDel = func(*netlink.Route) error { t.Fatal("retirement deleted replaced route"); return nil }
	t.Cleanup(func() { ruleDel, routeDel = oldRule, oldRoute })
	for _, sameDevice := range []bool{true, false} {
		dev := &fakeDevice{}
		held := &sharedTunnel{dev: dev, ueIP: "10.1.0.1", rule: &netlink.Rule{}, route: &netlink.Route{}}
		source := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Index: 10}}
		targetIndex := 11
		if sameDevice {
			targetIndex = 10
		}
		held.retireOn(source, &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Index: targetIndex}})
		held.release(nil)
		if sameDevice {
			require.Equal(t, []string{"rules"}, dev.calls)
		} else {
			require.Equal(t, []string{"address", "rules"}, dev.calls)
		}
	}
}
