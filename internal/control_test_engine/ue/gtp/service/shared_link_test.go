/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package service

import (
	"errors"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/free5gc/go-gtp5gnl"
	"github.com/khirono/go-nl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedLinkNameIsPerGnbAndFitsIfnamsiz(t *testing.T) {
	a := sharedLinkName(netip.MustParseAddr("10.0.0.1"))
	b := sharedLinkName(netip.MustParseAddr("10.0.0.2"))

	assert.Equal(t, "valgnb0a000001", a)
	assert.NotEqual(t, a, b, "each gNB needs its own device")
	assert.LessOrEqual(t, len(a), 15, "Linux interface names are capped at 15 characters")
	assert.Equal(t, a, sharedLinkName(netip.MustParseAddr("::ffff:10.0.0.1")), "an IPv4-mapped address names the same device")
}

// Every UE on a device needs PDR and FAR identifiers no other UE on it holds, while all of
// them share the one QER.
func TestTakeAllocatesDisjointRuleIDs(t *testing.T) {
	link := &sharedLink{}
	const ues = 1000

	var wg sync.WaitGroup
	got := make(chan ruleIDs, ues)
	for i := 0; i < ues; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got <- link.take()
		}()
	}
	wg.Wait()
	close(got)

	pdrs := map[uint32]bool{}
	fars := map[uint32]bool{}
	for ids := range got {
		for _, id := range []uint32{ids.pdrDown, ids.pdrUp} {
			require.False(t, pdrs[id], "PDR ID %d handed out twice", id)
			pdrs[id] = true
		}
		for _, id := range []uint32{ids.farDown, ids.farUp} {
			require.False(t, fars[id], "FAR ID %d handed out twice", id)
			fars[id] = true
		}
		assert.Equal(t, uint32(sharedQERID), ids.qer)
	}
	assert.Len(t, pdrs, 2*ues)
	assert.Len(t, fars, 2*ues)
}

// The first UE on a shared device gets the identifiers a dedicated device always used, so
// the dedicated mode's rules are unchanged.
func TestFirstSharedUeMatchesDedicatedIDs(t *testing.T) {
	assert.Equal(t, dedicatedRuleIDs(), (&sharedLink{}).take())
}

// gtp5g refuses a QER that already exists, so concurrent UEs must create it exactly once.
func TestEnsureQERCreatesOnce(t *testing.T) {
	link := &sharedLink{}
	var calls atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, link.EnsureQER(func() error { calls.Add(1); return nil }))
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), calls.Load())
}

// One failed attempt must not disable QoS marking for every later UE on the device.
func TestEnsureQERRetriesAfterFailure(t *testing.T) {
	link := &sharedLink{}
	var calls int
	fail := true
	create := func() error {
		calls++
		if fail {
			return errors.New("refused")
		}
		return nil
	}

	assert.False(t, link.EnsureQER(create))
	fail = false
	assert.True(t, link.EnsureQER(create), "the next UE should retry the create")
	assert.True(t, link.EnsureQER(create))
	assert.Equal(t, 2, calls, "once created, the QER is not created again")
}

// A UE that leaves gives its identifiers back, so a --loop run reuses them instead of
// walking the 16-bit PDR identifier space.
func TestTakeReusesGivenBackSlot(t *testing.T) {
	link := &sharedLink{}
	first := link.take()
	second := link.take()

	link.giveBack(first)
	assert.Equal(t, first, link.take(), "a returned slot is taken before a new one")
	assert.Equal(t, idsForSlot(2), link.take(), "with none returned, a new slot opens")
	assert.NotEqual(t, first.pdrUp, second.pdrUp)
}

// Without a long-lived client a device behaves exactly as before: the per-call wrapper
// installs the rule, with the arguments unchanged.
func TestAddRuleFallsBackWithoutClient(t *testing.T) {
	link := &sharedLink{name: "valgnb0a000001"}
	args := []string{"valgnb0a000001", "3", "--action", "2"}

	var fellBack []string
	err := link.addRule(args,
		func([]string) ([]nl.Attr, error) { t.Fatal("should not parse without a client"); return nil, nil },
		func(*gtp5gnl.Client, *gtp5gnl.Link, gtp5gnl.OID, []nl.Attr) error {
			t.Fatal("should not create through a client it does not have")
			return nil
		},
		func(a []string) error { fellBack = a; return nil },
	)

	require.NoError(t, err)
	assert.Equal(t, args, fellBack)
}

func TestSetupConcurrency(t *testing.T) {
	for value, want := range map[string]int{"": 32, "8": 8, "0": 32, "-1": 32, "many": 32} {
		t.Setenv("PR_SETUP_SLOTS", value)
		assert.Equal(t, want, setupConcurrency(), "PR_SETUP_SLOTS=%q", value)
	}
}
