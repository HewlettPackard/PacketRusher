/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package test

import (
	gnbctx "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"net/netip"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGNBContext creates a GNBContext suitable for unit testing.
func newTestGNBContext(t *testing.T) *gnbctx.GNBContext {
	t.Helper()
	gnb := &gnbctx.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	require.NotNil(t, amf)
	amf.SetStateActive()
	return gnb
}

// TestConcurrentIDGeneration verifies that concurrent calls to NewGnBUe produce
// unique RAN UE IDs. This exercises the actual atomic ID generators in GNBContext.
// Run with -race to detect data races: go test -race ./test/...
func TestConcurrentIDGeneration(t *testing.T) {
	gnb := newTestGNBContext(t)

	const numGoroutines = 100
	var wg sync.WaitGroup
	ids := make([]int64, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			gnbTx := make(chan gnbctx.UEMessage, 10)
			gnbRx := make(chan gnbctx.UEMessage, 10)
			ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(idx+1000), nil)
			if assert.NoError(t, err) && assert.NotNil(t, ue) {
				ids[idx] = ue.GetRanUeId()
			}
		}(i)
	}
	wg.Wait()

	seen := make(map[int64]bool, numGoroutines)
	for _, id := range ids {
		if id == 0 {
			continue // UE creation failed (counted by assert above)
		}
		assert.False(t, seen[id], "Duplicate RAN UE ID generated concurrently: %d", id)
		seen[id] = true
	}
}

// TestRapidUEOperations simulates the rapid registration/deregistration cycle
// described in issue #187 (concurrent NewGnBUe / GetGnbUe / DeleteGnBUe).
// It must be run with -race to be useful: go test -race ./test/...
func TestRapidUEOperations(t *testing.T) {
	gnb := newTestGNBContext(t)

	const numCycles = 50
	var wg sync.WaitGroup

	for i := 0; i < numCycles; i++ {
		wg.Add(1)
		go func(prUeId int64) {
			defer wg.Done()
			gnbTx := make(chan gnbctx.UEMessage, 10)
			gnbRx := make(chan gnbctx.UEMessage, 10)

			ue, err := gnb.NewGnBUe(gnbTx, gnbRx, prUeId, nil)
			if !assert.NoError(t, err) || !assert.NotNil(t, ue) {
				return
			}

			ranId := ue.GetRanUeId()

			// UE must be retrievable immediately after creation.
			retrieved, err := gnb.GetGnbUe(ranId)
			assert.NoError(t, err)
			assert.Equal(t, ue, retrieved)

			// Delete (deregistration).
			gnb.DeleteGnBUe(ue)

			// UE must no longer be retrievable.
			_, err = gnb.GetGnbUe(ranId)
			assert.Error(t, err, "deleted UE should not be found")
		}(int64(i + 2000))
	}
	wg.Wait()
}

// TestSendToUeBlocksUntilConsumed verifies that SendToUe blocks when the channel
// is full (preserving the message) rather than silently dropping it.
func TestSendToUeBlocksUntilConsumed(t *testing.T) {
	ue := &gnbctx.GNBUe{}
	gnbTx := make(chan gnbctx.UEMessage, 1)
	ue.SetRanUeId(99)
	ue.SetGnbTx(gnbTx)

	// Pre-fill the channel so SendToUe will initially block.
	gnbTx <- gnbctx.UEMessage{}

	sent := make(chan struct{})
	go func() {
		sender.SendToUe(ue, []byte("hello"))
		close(sent)
	}()

	// Drain the pre-filled message so SendToUe can proceed.
	<-gnbTx

	// SendToUe must complete and the message must arrive.
	<-sent
	require.Equal(t, 1, len(gnbTx))
	msg := <-gnbTx
	assert.True(t, msg.IsNas)
	assert.Equal(t, []byte("hello"), msg.Nas)
}

// TestSendToUeDropsWhenChannelNil verifies that SendToUe silently drops the
// message (no panic) when the channel has been closed by DeleteGnBUe.
func TestSendToUeDropsWhenChannelNil(t *testing.T) {
	ue := &gnbctx.GNBUe{}
	ue.SetRanUeId(99)
	ue.SetGnbTx(nil) // simulate post-DeleteGnBUe state

	// Must not panic.
	sender.SendToUe(ue, []byte("dropped"))
}

// TestRaceConditionPrevention creates and retrieves UEs from multiple goroutines
// simultaneously. When run with -race this will catch any remaining data races
// in the ID generators or UE pool: go test -race ./test/...
func TestRaceConditionPrevention(t *testing.T) {
	gnb := newTestGNBContext(t)

	const numWorkers = 50
	var wg sync.WaitGroup
	createdUEs := make([]*gnbctx.GNBUe, numWorkers)

	// Create UEs concurrently.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			gnbTx := make(chan gnbctx.UEMessage, 10)
			gnbRx := make(chan gnbctx.UEMessage, 10)
			ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(idx+3000), nil)
			if assert.NoError(t, err) {
				createdUEs[idx] = ue
			}
		}(i)
	}
	wg.Wait()

	// Retrieve UEs concurrently.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ue := createdUEs[idx]
			if ue == nil {
				return
			}
			retrieved, err := gnb.GetGnbUe(ue.GetRanUeId())
			assert.NoError(t, err)
			assert.Equal(t, ue, retrieved)
		}(i)
	}
	wg.Wait()

	// Delete UEs concurrently.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if createdUEs[idx] != nil {
				gnb.DeleteGnBUe(createdUEs[idx])
			}
		}(i)
	}
	wg.Wait()
}
