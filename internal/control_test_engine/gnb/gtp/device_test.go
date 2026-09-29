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
	"syscall"
	"testing"
	"time"

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
			ids, err := dev.Take()
			assert.NoError(t, err)
			got <- ids
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
	ids, err := (&Device{}).Take()
	require.NoError(t, err)
	assert.Equal(t, DedicatedRuleIDs(), ids)
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

// One failed attempt must not disable QoS marking for every later UE with that QFI; the
// retry uses the same identifier.
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

// A create the kernel applied without acknowledging leaves a QER with the identifier.
// Another QFI must not be given it, or its create collides with a QER marking the
// first QFI.
func TestEnsureQERNeverReusesAnIdentifierAcrossQFIs(t *testing.T) {
	dev := &Device{}

	_, ok := dev.EnsureQER(9, func(uint32) error { return errors.New("lost ack") })
	require.False(t, ok)

	var other uint32
	id, ok := dev.EnsureQER(5, func(id uint32) error { other = id; return nil })
	require.True(t, ok)
	assert.Equal(t, uint32(2), other, "identifier 1 stays reserved for QFI 9")
	assert.Equal(t, uint32(2), id)
}

// The retry after a lost acknowledgement finds the QER already there. Its identifier is
// reserved for this QFI, so the QER there is this QFI's and is usable.
func TestEnsureQERTakesAlreadyExistsAsCreated(t *testing.T) {
	dev := &Device{}

	_, ok := dev.EnsureQER(9, func(uint32) error { return errors.New("lost ack") })
	require.False(t, ok)

	id, ok := dev.EnsureQER(9, func(uint32) error { return syscall.EEXIST })
	assert.True(t, ok)
	assert.Equal(t, uint32(1), id)
}

// A QER that cannot be created is not retried by every UE for the rest of the run.
func TestEnsureQERGivesUpAfterMaxAttempts(t *testing.T) {
	dev := &Device{}
	calls := 0
	failing := func(uint32) error { calls++; return errors.New("refused") }

	for i := 0; i < maxQERAttempts+5; i++ {
		_, ok := dev.EnsureQER(9, failing)
		assert.False(t, ok)
	}

	assert.Equal(t, maxQERAttempts, calls)
}

// The PDR identifier is 16 bits and each UE takes two, so a device holds a bounded
// number of UEs. Past that, a UE gets an error, not identifiers that wrap onto another
// UE's rules.
func TestTakeRefusesPastTheLastSlot(t *testing.T) {
	dev := &Device{nextUE: maxSlots - 1}

	last, err := dev.Take()
	require.NoError(t, err)
	assert.Equal(t, uint32(65534), last.PDRUp, "the last slot's uplink PDR is the largest 16-bit identifier it can use")

	_, err = dev.Take()
	assert.ErrorIs(t, err, ErrDeviceFull)

	dev.giveBack(last)
	again, err := dev.Take()
	require.NoError(t, err, "a slot given back can be taken again")
	assert.Equal(t, last, again)
}

// The gNB waits for its UEs to give their rules back before it removes the device, but
// not for ones that never will.
func TestWaitIdle(t *testing.T) {
	dev := &Device{}
	first, err := dev.Take()
	require.NoError(t, err)
	second, err := dev.Take()
	require.NoError(t, err)

	assert.Equal(t, 2, dev.WaitIdle(20*time.Millisecond, time.Second), "nothing given back: stop after the stall")

	go func() {
		time.Sleep(30 * time.Millisecond)
		dev.giveBack(first)
		time.Sleep(30 * time.Millisecond)
		dev.giveBack(second)
	}()
	assert.Zero(t, dev.WaitIdle(200*time.Millisecond, 5*time.Second), "releases still coming: wait for them")

	_, err = dev.Take()
	require.NoError(t, err)
	begun := time.Now()
	assert.Equal(t, 1, dev.WaitIdle(time.Hour, 50*time.Millisecond))
	assert.Less(t, time.Since(begun), time.Second, "the limit bounds the wait")
}

// A UE that leaves gives its identifiers back, so a --loop run reuses them instead of
// walking the 16-bit PDR identifier space.
func TestTakeReusesGivenBackSlot(t *testing.T) {
	dev := &Device{}
	first, _ := dev.Take()
	second, _ := dev.Take()

	dev.giveBack(first)
	again, _ := dev.Take()
	assert.Equal(t, first, again, "a returned slot is taken before a new one")
	next, _ := dev.Take()
	assert.Equal(t, idsForSlot(2), next, "with none returned, a new slot opens")
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
