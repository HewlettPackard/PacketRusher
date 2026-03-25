/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sender

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestUE() *context.GNBUe {
	ue := &context.GNBUe{}
	ue.SetRanUeId(12345)
	return ue
}

func TestSendToUe_ValidChannel(t *testing.T) {
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 10)
	ue.SetGnbTx(gnbTx)

	testMessage := []byte("test NAS message")

	// Send message
	SendToUe(ue, testMessage)

	// Verify message was received
	select {
	case receivedMsg := <-gnbTx:
		assert.True(t, receivedMsg.IsNas, "Message should be marked as NAS")
		assert.Equal(t, testMessage, receivedMsg.Nas, "Message content should match")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestSendToUe_NilChannel(t *testing.T) {
	ue := createTestUE()
	ue.SetGnbTx(nil) // Simulate closed channel

	testMessage := []byte("test NAS message")

	// Send message - should not panic
	SendToUe(ue, testMessage)

	// No specific assertion needed - test passes if no panic occurs
}

func TestSendToUe_FullChannel(t *testing.T) {
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 1) // Small buffer
	ue.SetGnbTx(gnbTx)

	// Fill the channel
	gnbTx <- context.UEMessage{}

	testMessage := []byte("test NAS message")

	// The send must block until a consumer drains the channel.
	go func() {
		time.Sleep(10 * time.Millisecond)
		<-gnbTx // drain the pre-filled message
	}()

	SendToUe(ue, testMessage)

	// The message must have been delivered (not dropped).
	assert.Equal(t, 1, len(gnbTx), "Message must be delivered after channel is drained")
	msg := <-gnbTx
	assert.True(t, msg.IsNas)
	assert.Equal(t, testMessage, msg.Nas)
}

func TestSendMessageToUe_ValidChannel(t *testing.T) {
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 10)
	ue.SetGnbTx(gnbTx)

	testMessage := context.UEMessage{
		IsNas:  false,
		PrUeId: 12345,
	}

	// Send message
	SendMessageToUe(ue, testMessage)

	// Verify message was received
	select {
	case receivedMsg := <-gnbTx:
		assert.Equal(t, testMessage.IsNas, receivedMsg.IsNas, "IsNas should match")
		assert.Equal(t, testMessage.PrUeId, receivedMsg.PrUeId, "PrUeId should match")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestSendMessageToUe_NilChannel(t *testing.T) {
	ue := createTestUE()
	ue.SetGnbTx(nil) // Simulate closed channel

	testMessage := context.UEMessage{
		IsNas:  false,
		PrUeId: 12345,
	}

	// Send message - should not panic
	SendMessageToUe(ue, testMessage)

	// No specific assertion needed - test passes if no panic occurs
}

func TestSendMessageToUe_FullChannel(t *testing.T) {
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 1) // Small buffer
	ue.SetGnbTx(gnbTx)

	// Fill the channel
	gnbTx <- context.UEMessage{}

	testMessage := context.UEMessage{
		IsNas:  false,
		PrUeId: 12345,
	}

	// The send must block until a consumer drains the channel.
	go func() {
		time.Sleep(10 * time.Millisecond)
		<-gnbTx // drain the pre-filled message
	}()

	SendMessageToUe(ue, testMessage)

	// The message must have been delivered (not dropped).
	assert.Equal(t, 1, len(gnbTx), "Message must be delivered after channel is drained")
	msg := <-gnbTx
	assert.Equal(t, testMessage.PrUeId, msg.PrUeId)
}

func TestConcurrentSendOperations(t *testing.T) {
	// Test concurrent send operations to ensure thread safety
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 100) // Buffer size of 100
	ue.SetGnbTx(gnbTx)

	const numGoroutines = 50
	const messagesPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	// Start a goroutine to consume messages and prevent channel from filling
	consumedCount := 0
	consumerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-gnbTx:
				consumedCount++
			case <-time.After(100 * time.Millisecond):
				consumerDone <- true
				return
			}
		}
	}()

	// Start concurrent senders
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineId int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				testMessage := []byte("test message")
				SendToUe(ue, testMessage)

				testUEMessage := context.UEMessage{
					PrUeId: int64(goroutineId*1000 + j),
				}
				SendMessageToUe(ue, testUEMessage)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent send operations")
		}
	}

	// Wait for consumer to finish
	<-consumerDone

	// Sends are now blocking: all messages must be delivered to the consumer.
	expectedMessages := numGoroutines * messagesPerGoroutine * 2 // 2 types of messages per iteration
	assert.Equal(t, expectedMessages, consumedCount,
		"All sent messages must be received with blocking sends")
}

func TestRapidSendOperations(t *testing.T) {
	// Test rapid send operations to simulate the 200ms scenario
	ue := createTestUE()
	gnbTx := make(chan context.UEMessage, 1000) // Large buffer to catch all messages
	ue.SetGnbTx(gnbTx)

	const numMessages = 100
	const rapidInterval = 2 * time.Millisecond // Very rapid sending

	start := time.Now()

	for i := 0; i < numMessages; i++ {
		testMessage := []byte("rapid test message")
		SendToUe(ue, testMessage)

		time.Sleep(rapidInterval)
	}

	duration := time.Since(start)

	// Verify all messages were sent quickly
	assert.Less(t, duration, 1*time.Second, "Rapid send operations should complete quickly")

	// Verify we received all messages
	assert.Equal(t, numMessages, len(gnbTx), "Should receive all rapidly sent messages")
}

// Note: a test for "channel closed during a concurrent send" is intentionally
// omitted: concurrently calling close(ch) while another goroutine is blocked on
// ch <- msg is flagged as a data race by the race detector even though the Go
// runtime handles it gracefully (the blocked send panics and our recover catches
// it). The coverage for that code path is exercised by the integration tests
// run with -race in test/concurrent_fixes_test.go.
