/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package tools

import (
	"my5G-RANTester/internal/control_test_engine/procedures"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startPark runs m.Park on scenario and returns a channel carrying its result, after
// waiting until the member reports itself parked.
func startPark(t *testing.T, m *ChurnMember, scenario chan procedures.UeTesterMessage) chan bool {
	t.Helper()
	done := make(chan bool, 1)
	go func() { done <- m.Park(scenario) }()
	require.Eventually(t, func() bool { return m.State() == ChurnParked }, 2*time.Second, time.Millisecond)
	return done
}

func result(t *testing.T, done chan bool) bool {
	t.Helper()
	select {
	case v := <-done:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal("Park did not return")
		return false
	}
}

func TestChurnMemberParkReturnsOnRearm(t *testing.T) {
	m := NewChurnMember()
	done := startPark(t, m, make(chan procedures.UeTesterMessage))

	assert.True(t, m.Rearm(), "a parked member takes the rearm")
	assert.True(t, result(t, done))
	assert.Equal(t, ChurnStarting, m.State(), "a rearmed member is starting again")
}

func TestChurnMemberParkEndsOnTerminateKillOrClose(t *testing.T) {
	for _, typ := range []procedures.UeTesterMessageType{procedures.Terminate, procedures.Kill} {
		m := NewChurnMember()
		scenario := make(chan procedures.UeTesterMessage)
		done := startPark(t, m, scenario)
		scenario <- procedures.UeTesterMessage{Type: typ}
		assert.False(t, result(t, done), "message type %d ends the member", typ)
	}

	// A closed scenario channel yields Registration, the zero value; it must not be
	// taken for a rearm.
	m := NewChurnMember()
	scenario := make(chan procedures.UeTesterMessage)
	done := startPark(t, m, scenario)
	close(scenario)
	assert.False(t, result(t, done))
}

func TestChurnMemberParkIgnoresOtherMessages(t *testing.T) {
	m := NewChurnMember()
	scenario := make(chan procedures.UeTesterMessage)
	done := startPark(t, m, scenario)

	// Unbuffered, so each send completes only once the parked member has read it.
	for _, typ := range []procedures.UeTesterMessageType{procedures.Idle, procedures.Registration} {
		select {
		case scenario <- procedures.UeTesterMessage{Type: typ}:
		case <-time.After(2 * time.Second):
			t.Fatalf("message type %d was not read while parked", typ)
		}
	}
	select {
	case v := <-done:
		t.Fatalf("Park returned %v without a rearm", v)
	default:
	}
	assert.Equal(t, ChurnParked, m.State())

	assert.True(t, m.Rearm())
	assert.True(t, result(t, done))
}

// A rearm for a member that has returned must not block the driver.
func TestChurnMemberRearmAfterExit(t *testing.T) {
	m := NewChurnMember()
	m.Exit()

	assert.Equal(t, ChurnGone, m.State())
	rearmed := make(chan bool, 1)
	go func() { rearmed <- m.Rearm() }()
	select {
	case v := <-rearmed:
		assert.False(t, v)
	case <-time.After(2 * time.Second):
		t.Fatal("Rearm blocked on a member that has returned")
	}
}
