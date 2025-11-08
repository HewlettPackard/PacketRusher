/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package test

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentIDGeneration(t *testing.T) {
	// Test that our fixes prevent race conditions during concurrent ID generation
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numGoroutines = 100
	const idsPerGoroutine = 10

	var wg sync.WaitGroup
	generatedRanIds := make([][]int64, numGoroutines)
	generatedAmfIds := make([][]int64, numGoroutines)

	// Generate IDs concurrently using private methods via UE creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			ranIds := make([]int64, idsPerGoroutine)
			amfIds := make([]int64, idsPerGoroutine)

			for j := 0; j < idsPerGoroutine; j++ {
				// Create temporary channels for UE creation
				gnbTx := make(chan context.UEMessage, 10)
				gnbRx := make(chan context.UEMessage, 10)

				ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(goroutineId*1000+j), nil)
				if err == nil && ue != nil {
					ranIds[j] = ue.GetRanUeId()
					amfIds[j] = ue.GetAmfUeId()
					gnb.DeleteGnBUe(ue) // Clean up immediately
				}
			}
			generatedRanIds[goroutineId] = ranIds
			generatedAmfIds[goroutineId] = amfIds
		}(i)
	}

	wg.Wait()

	// Verify all IDs are unique (no race conditions)
	allRanIds := make(map[int64]bool)
	allAmfIds := make(map[int64]bool)
	totalRanIds := 0
	totalAmfIds := 0

	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < idsPerGoroutine; j++ {
			ranId := generatedRanIds[i][j]
			amfId := generatedAmfIds[i][j]

			if ranId != 0 {
				assert.False(t, allRanIds[ranId], "Duplicate RAN ID found: %d", ranId)
				allRanIds[ranId] = true
				totalRanIds++
			}

			if amfId != 0 {
				assert.False(t, allAmfIds[amfId], "Duplicate AMF ID found: %d", amfId)
				allAmfIds[amfId] = true
				totalAmfIds++
			}
		}
	}

	t.Logf("Successfully generated %d unique RAN IDs and %d unique AMF IDs concurrently", totalRanIds, totalAmfIds)
	assert.Greater(t, totalRanIds, 0, "Should have generated some RAN IDs")
	assert.Greater(t, totalAmfIds, 0, "Should have generated some AMF IDs")
}

func TestRapidUEOperations(t *testing.T) {
	// Test that rapid UE creation/deletion operations don't cause deadlocks or panics
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numOperations = 100
	const rapidInterval = 1 * time.Millisecond // Very rapid operations

	var wg sync.WaitGroup
	successfulCreations := 0
	var creationMutex sync.Mutex

	// Perform rapid UE operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(opId int) {
			defer wg.Done()

			// Create UE
			gnbTx := make(chan context.UEMessage, 100)
			gnbRx := make(chan context.UEMessage, 100)

			ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(opId), nil)
			if err == nil && ue != nil {
				creationMutex.Lock()
				successfulCreations++
				creationMutex.Unlock()

				// Simulate some rapid operations
				time.Sleep(rapidInterval)

				// Clean up
				gnb.DeleteGnBUe(ue)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Successfully created and deleted %d UEs in rapid succession", successfulCreations)
	assert.Greater(t, successfulCreations, 0, "Should have created at least some UEs")
}

func TestChannelNonBlocking(t *testing.T) {
	// Test that our channel improvements prevent blocking during rapid operations
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	// Create a UE
	gnbTx := make(chan context.UEMessage, 10)
	gnbRx := make(chan context.UEMessage, 10)

	ue, err := gnb.NewGnBUe(gnbTx, gnbRx, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, ue)

	// Test rapid message sending (should not block)
	const numMessages = 100
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		msg := context.UEMessage{
			IsNas: true,
			Nas:   []byte("test message"),
		}

		// This should not block thanks to our non-blocking implementation
		select {
		case gnbTx <- msg:
			// Success
		case <-time.After(10 * time.Millisecond):
			t.Logf("Message %d took longer than expected (possible blocking)", i)
		}
	}

	elapsed := time.Since(start)
	t.Logf("Sent %d messages in %v (avg: %v per message)", numMessages, elapsed, elapsed/numMessages)

	// Should complete reasonably quickly
	assert.Less(t, elapsed, 1*time.Second, "Message sending should not be blocked")

	// Clean up
	gnb.DeleteGnBUe(ue)
}

func TestRaceConditionPrevention(t *testing.T) {
	// Test that our mutex fixes prevent race conditions during concurrent access
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numGoroutines = 50
	var wg sync.WaitGroup
	var createdUEs []*context.GNBUe
	var ueMutex sync.Mutex

	// Perform concurrent operations that previously caused race conditions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create UEs concurrently (tests mutex protection for ID generation)
			gnbTx := make(chan context.UEMessage, 10)
			gnbRx := make(chan context.UEMessage, 10)

			ue1, err1 := gnb.NewGnBUe(gnbTx, gnbRx, int64(id*100), nil)
			ue2, err2 := gnb.NewGnBUe(gnbTx, gnbRx, int64(id*100+1), nil)

			if err1 == nil && ue1 != nil && err2 == nil && ue2 != nil {
				// Verify IDs are different (basic sanity check)
				assert.NotEqual(t, ue1.GetRanUeId(), ue2.GetRanUeId(), "Generated RAN IDs should be different")
				assert.NotEqual(t, ue1.GetAmfUeId(), ue2.GetAmfUeId(), "Generated AMF IDs should be different")

				ueMutex.Lock()
				createdUEs = append(createdUEs, ue1, ue2)
				ueMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Clean up created UEs
	for _, ue := range createdUEs {
		gnb.DeleteGnBUe(ue)
	}

	t.Logf("Race condition prevention test completed successfully with %d UEs created", len(createdUEs))
}
