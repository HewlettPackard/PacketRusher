/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package gtp

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

func TestDeviceNameIsPerGnbAndFitsIfnamsiz(t *testing.T) {
	a := deviceName(netip.MustParseAddr("10.0.0.1"))
	b := deviceName(netip.MustParseAddr("10.0.0.2"))

	assert.Equal(t, "valgnb0a000001", a)
	assert.NotEqual(t, a, b, "each gNB needs its own device")
	assert.LessOrEqual(t, len(a), 15, "Linux interface names are capped at 15 characters")
	assert.Equal(t, a, deviceName(netip.MustParseAddr("::ffff:10.0.0.1")), "an IPv4-mapped address names the same device")
}

// Every UE on a device needs PDR and FAR identifiers no other UE on it holds.
func TestTakeAllocatesDisjointRuleIDs(t *testing.T) {
	dev := &Device{}
	const ues = 1000

	var wg sync.WaitGroup
	got := make(chan RuleIDs, ues)
	for i := 0; i < ues; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got <- dev.Take()
		}()
	}
	wg.Wait()
	close(got)

	pdrs := map[uint32]bool{}
	fars := map[uint32]bool{}
	for ids := range got {
		for _, id := range []uint32{ids.PDRDown, ids.PDRUp} {
			require.False(t, pdrs[id], "PDR ID %d handed out twice", id)
			pdrs[id] = true
		}
		for _, id := range []uint32{ids.FARDown, ids.FARUp} {
			require.False(t, fars[id], "FAR ID %d handed out twice", id)
			fars[id] = true
		}
	}
	assert.Len(t, pdrs, 2*ues)
	assert.Len(t, fars, 2*ues)
}

// The first UE on a shared device gets the identifiers a dedicated device always used, so
// the dedicated mode's rules are unchanged.
func TestFirstSharedUeMatchesDedicatedIDs(t *testing.T) {
	assert.Equal(t, DedicatedRuleIDs(), (&Device{}).Take())
}

// gtp5g refuses a QER that already exists, so concurrent UEs with one QFI must create it
// exactly once, and all of them must reference it.
func TestEnsureQERCreatesOncePerQFI(t *testing.T) {
	dev := &Device{}
	var calls atomic.Int32

	var wg sync.WaitGroup
	ids := make(chan uint32, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, ok := dev.EnsureQER(9, func(uint32) error { calls.Add(1); return nil })
			assert.True(t, ok)
			ids <- id
		}()
	}
	wg.Wait()
	close(ids)

	assert.Equal(t, int32(1), calls.Load())
	for id := range ids {
		assert.Equal(t, uint32(1), id)
	}
}

// A UE whose flow has another QFI must not be pointed at the QER of the first UE's QFI:
// the uplink would be marked with the wrong QFI.
func TestEnsureQERSeparatesQFIs(t *testing.T) {
	dev := &Device{}
	created := map[uint32]bool{}
	create := func(id uint32) error {
		require.False(t, created[id], "QER %d created twice", id)
		created[id] = true
		return nil
	}

	first, ok := dev.EnsureQER(9, create)
	require.True(t, ok)
	second, ok := dev.EnsureQER(5, create)
	require.True(t, ok)
	again, ok := dev.EnsureQER(9, create)
	require.True(t, ok)

	assert.NotEqual(t, first, second, "different QFIs need different QERs")
	assert.Equal(t, first, again, "one QFI keeps its QER")
	assert.Len(t, created, 2)
}

// One failed attempt must not disable QoS marking for every later UE with that QFI, and
// must not use up an identifier.
func TestEnsureQERRetriesAfterFailure(t *testing.T) {
	dev := &Device{}
	var tried []uint32
	fail := true
	create := func(id uint32) error {
		tried = append(tried, id)
		if fail {
			return errors.New("refused")
		}
		return nil
	}

	_, ok := dev.EnsureQER(9, create)
	assert.False(t, ok)
	fail = false
	id, ok := dev.EnsureQER(9, create)
	assert.True(t, ok, "the next UE should retry the create")
	_, ok = dev.EnsureQER(9, create)
	assert.True(t, ok)
	assert.Equal(t, []uint32{1, 1}, tried, "the retry reuses the identifier; once created, it is not created again")
	assert.Equal(t, uint32(1), id)
}

// A UE that leaves gives its identifiers back, so a --loop run reuses them instead of
// walking the 16-bit PDR identifier space.
func TestTakeReusesGivenBackSlot(t *testing.T) {
	dev := &Device{}
	first := dev.Take()
	second := dev.Take()

	dev.giveBack(first)
	assert.Equal(t, first, dev.Take(), "a returned slot is taken before a new one")
	assert.Equal(t, idsForSlot(2), dev.Take(), "with none returned, a new slot opens")
	assert.NotEqual(t, first.PDRUp, second.PDRUp)
}

// Without a long-lived client a device behaves exactly as before: the per-call wrapper
// installs the rule, with the arguments unchanged.
func TestAddRuleFallsBackWithoutClient(t *testing.T) {
	dev := &Device{name: "valgnb0a000001"}
	args := []string{"valgnb0a000001", "3", "--action", "2"}

	var fellBack []string
	err := dev.addRule(args,
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
