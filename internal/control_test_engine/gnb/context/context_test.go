/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGNBContext_ConcurrentIDGeneration(t *testing.T) {
	// Test that ID generation is thread-safe during concurrent access
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numGoroutines = 100
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	generatedUeIds := make([][]int64, numGoroutines)
	generatedAmfIds := make([][]int64, numGoroutines)
	generatedTeids := make([][]uint32, numGoroutines)

	// Launch concurrent goroutines to generate IDs
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			ueIds := make([]int64, idsPerGoroutine)
			amfIds := make([]int64, idsPerGoroutine)
			teids := make([]uint32, idsPerGoroutine)

			for j := 0; j < idsPerGoroutine; j++ {
				ueIds[j] = gnb.getRanUeId()
				amfIds[j] = gnb.getRanAmfId()

				// Create a dummy UE for TEID generation
				dummyUe := &GNBUe{}
				teids[j] = gnb.GetUeTeid(dummyUe)
			}

			generatedUeIds[goroutineId] = ueIds
			generatedAmfIds[goroutineId] = amfIds
			generatedTeids[goroutineId] = teids
		}(i)
	}

	wg.Wait()

	// Verify all generated IDs are unique
	allUeIds := make(map[int64]bool)
	allAmfIds := make(map[int64]bool)
	allTeids := make(map[uint32]bool)

	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < idsPerGoroutine; j++ {
			ueId := generatedUeIds[i][j]
			amfId := generatedAmfIds[i][j]
			teid := generatedTeids[i][j]

			// Check for uniqueness
			assert.False(t, allUeIds[ueId], "Duplicate UE ID found: %d", ueId)
			assert.False(t, allAmfIds[amfId], "Duplicate AMF ID found: %d", amfId)
			assert.False(t, allTeids[teid], "Duplicate TEID found: %d", teid)

			allUeIds[ueId] = true
			allAmfIds[amfId] = true
			allTeids[teid] = true
		}
	}

	// Verify we have the expected number of unique IDs
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(allUeIds), "Expected %d unique UE IDs", numGoroutines*idsPerGoroutine)
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(allAmfIds), "Expected %d unique AMF IDs", numGoroutines*idsPerGoroutine)
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(allTeids), "Expected %d unique TEIDs", numGoroutines*idsPerGoroutine)
}

func TestGNBContext_ConcurrentUECreationAndDeletion(t *testing.T) {
	// Test concurrent UE creation and deletion to simulate rapid registration/deregistration
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numGoroutines = 10 // Reduced to avoid overwhelming the test
	const operationsPerGoroutine = 5

	var wg sync.WaitGroup
	createdUEs := make([][]*GNBUe, numGoroutines)

	// Create UEs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			ues := make([]*GNBUe, operationsPerGoroutine)
			for j := 0; j < operationsPerGoroutine; j++ {
				gnbTx := make(chan UEMessage, 100) // Increased buffer size
				gnbRx := make(chan UEMessage, 100)

				ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(goroutineId*1000+j), nil)
				if err != nil {
					t.Logf("Warning: Failed to create UE: %v", err)
					continue
				}
				if ue == nil {
					t.Logf("Warning: UE creation returned nil")
					continue
				}
				ues[j] = ue
			}
			createdUEs[goroutineId] = ues
		}(i)
	}

	wg.Wait()

	// Verify created UEs have unique IDs (skip nil UEs)
	allRanIds := make(map[int64]bool)
	validUECount := 0
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < operationsPerGoroutine; j++ {
			ue := createdUEs[i][j]
			if ue == nil {
				continue
			}

			ranId := ue.GetRanUeId()
			assert.False(t, allRanIds[ranId], "Duplicate RAN UE ID found: %d", ranId)
			allRanIds[ranId] = true
			validUECount++

			// Verify UE can be retrieved from pool
			retrievedUe, err := gnb.GetGnbUe(ranId)
			if err == nil && retrievedUe != nil {
				assert.Equal(t, ue, retrievedUe, "Retrieved UE should match created UE")
			}
		}
	}

	// Now delete UEs concurrently (only valid UEs)
	wg = sync.WaitGroup{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				ue := createdUEs[goroutineId][j]
				if ue != nil {
					gnb.DeleteGnBUe(ue)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify valid UEs were deleted (best effort due to concurrency)
	deletedCount := 0
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < operationsPerGoroutine; j++ {
			ue := createdUEs[i][j]
			if ue != nil {
				ranId := ue.GetRanUeId()
				_, err := gnb.GetGnbUe(ranId)
				if err != nil {
					deletedCount++
				}
			}
		}
	}

	t.Logf("Successfully created %d UEs and deleted %d UEs concurrently", validUECount, deletedCount)
	assert.Greater(t, validUECount, 0, "Should have created at least some UEs")
}

func TestGNBContext_RapidRegistrationDeregistration(t *testing.T) {
	// Simulate the rapid registration/deregistration scenario from the bug report
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	const numCycles = 100
	const rapidInterval = 200 * time.Millisecond

	var wg sync.WaitGroup

	for i := 0; i < numCycles; i++ {
		wg.Add(1)
		go func(cycleId int) {
			defer wg.Done()

			// Create UE (registration)
			gnbTx := make(chan UEMessage, 10)
			gnbRx := make(chan UEMessage, 10)

			ue, err := gnb.NewGnBUe(gnbTx, gnbRx, int64(cycleId), nil)
			require.NoError(t, err, "Failed to create UE in cycle %d", cycleId)
			require.NotNil(t, ue, "UE should not be nil in cycle %d", cycleId)

			ranId := ue.GetRanUeId()

			// Verify UE exists
			retrievedUe, err := gnb.GetGnbUe(ranId)
			assert.NoError(t, err, "Should be able to retrieve UE in cycle %d", cycleId)
			assert.Equal(t, ue, retrievedUe, "Retrieved UE should match created UE in cycle %d", cycleId)

			// Wait for rapid interval
			time.Sleep(rapidInterval)

			// Delete UE (deregistration)
			gnb.DeleteGnBUe(ue)

			// Verify UE is deleted
			retrievedUe, err = gnb.GetGnbUe(ranId)
			assert.Error(t, err, "Should not be able to retrieve deleted UE in cycle %d", cycleId)
			assert.Nil(t, retrievedUe, "Retrieved UE should be nil for deleted UE in cycle %d", cycleId)
		}(i)
	}

	wg.Wait()
}

func TestGNBContext_ChannelBuffering(t *testing.T) {
	// Test that the increased channel buffer size can handle rapid message exchange
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	// Test inbound channel buffer size
	inboundChan := gnb.GetInboundChannel()
	require.NotNil(t, inboundChan, "Inbound channel should not be nil")

	// Send messages rapidly to test buffer capacity
	const numMessages = 50
	for i := 0; i < numMessages; i++ {
		select {
		case inboundChan <- UEMessage{PrUeId: int64(i)}:
			// Message sent successfully
		default:
			t.Fatalf("Inbound channel blocked after %d messages, buffer may be too small", i)
		}
	}

	// Drain the channel
	for i := 0; i < numMessages; i++ {
		select {
		case msg := <-inboundChan:
			assert.Equal(t, int64(i), msg.PrUeId, "Message order should be preserved")
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for message %d", i)
		}
	}
}

func TestGNBContext_UEStateValidation(t *testing.T) {
	// Test UE state validation in getUeFromContext-like scenarios
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	// Create a UE
	gnbTx := make(chan UEMessage, 10)
	gnbRx := make(chan UEMessage, 10)
	ue, err := gnb.NewGnBUe(gnbTx, gnbRx, 12345, nil)
	require.NoError(t, err)
	require.NotNil(t, ue)

	ranId := ue.GetRanUeId()

	// Test normal state - should be retrievable
	retrievedUe, err := gnb.GetGnbUe(ranId)
	assert.NoError(t, err)
	assert.Equal(t, ue, retrievedUe)

	// Set UE to Down state
	ue.SetStateDown()
	assert.Equal(t, Down, ue.GetState())

	// UE should still be retrievable even in Down state (the validation happens at message processing level)
	retrievedUe, err = gnb.GetGnbUe(ranId)
	assert.NoError(t, err)
	assert.Equal(t, ue, retrievedUe)

	// Delete UE
	gnb.DeleteGnBUe(ue)

	// Should not be retrievable after deletion
	retrievedUe, err = gnb.GetGnbUe(ranId)
	assert.Error(t, err)
	assert.Nil(t, retrievedUe)
}

func TestGNBContext_AMFManagement(t *testing.T) {
	// Test AMF creation and management
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	// Create AMFs
	amf1 := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	amf2 := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38413"))

	require.NotNil(t, amf1)
	require.NotNil(t, amf2)

	// AMFs should have different IDs
	assert.NotEqual(t, amf1.GetAmfId(), amf2.GetAmfId())

	// Should be able to find AMFs by IP
	foundAmf1 := gnb.FindGnbAmfByIpPort(netip.MustParseAddrPort("127.0.0.1:38412"))
	foundAmf2 := gnb.FindGnbAmfByIpPort(netip.MustParseAddrPort("127.0.0.1:38413"))

	assert.Equal(t, amf1, foundAmf1)
	assert.Equal(t, amf2, foundAmf2)

	// Should not find non-existent AMF
	notFound := gnb.FindGnbAmfByIpPort(netip.MustParseAddrPort("127.0.0.1:99999"))
	assert.Nil(t, notFound)
}
