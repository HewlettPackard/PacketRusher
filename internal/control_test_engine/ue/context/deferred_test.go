/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"runtime"
	"testing"
	"time"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUE() *UEContext {
	capability := nasType.NewUESecurityCapability(nasMessage.RegistrationRequestUESecurityCapabilityType)
	capability.SetLen(2)
	capability.Buffer = []uint8{0x80, 0x80}
	ue := &UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	return ue
}

func TestRunOnUEAfterHandsTheJobToTheUEGoroutine(t *testing.T) {
	ue := newTestUE()
	ran := false
	ue.RunOnUEAfter(10*time.Millisecond, func() { ran = true })

	select {
	case f := <-ue.Deferred():
		assert.False(t, ran, "the job runs only when the UE's goroutine runs it")
		f()
		assert.True(t, ran)
	case <-time.After(2 * time.Second):
		t.Fatal("the job should be handed to the UE's goroutine once the wait is over")
	}
}

// waitFor polls cond for up to two seconds on the test's own goroutine. It does not use
// testify's Eventually, which runs cond on goroutines of its own and so skews a count of
// them.
func waitFor(cond func() bool) bool {
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if cond() {
			return true
		}
	}
	return false
}

// A job still waiting when the UE terminates must not run, and its goroutine ends there
// rather than sleeping out its wait.
func TestTerminateCancelsWaitingJobs(t *testing.T) {
	ue := newTestUE()
	before := runtime.NumGoroutine()
	ran := make(chan struct{}, 2)
	ue.RunOnUEAfter(time.Hour, func() { ran <- struct{}{} })
	require.True(t, waitFor(func() bool { return runtime.NumGoroutine() > before }),
		"the waiting job's goroutine should be running")

	ue.Terminate()

	assert.True(t, waitFor(func() bool { return runtime.NumGoroutine() <= before }),
		"the waiting job's goroutine should end at Terminate, not after its hour")
	handed := make(chan bool, 1)
	go func() { handed <- ue.RunOnUE(func() { ran <- struct{}{} }) }()
	select {
	case ok := <-handed:
		assert.False(t, ok, "nothing is handed to a terminated UE")
	case <-time.After(2 * time.Second):
		t.Fatal("RunOnUE should give up once the UE has terminated")
	}
	assert.Empty(t, ran, "no job ran after Terminate")
}

func TestEstablishmentBackoff(t *testing.T) {
	ue := newTestUE()

	remaining, deactivated := ue.EstablishmentBackoff()
	assert.Zero(t, remaining)
	assert.False(t, deactivated)

	ue.SetEstablishmentBackoff(time.Hour, false)
	remaining, deactivated = ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute)
	assert.False(t, deactivated)

	// TS 24.501 6.4.1.4.2 c): a zero value stops the timer.
	ue.SetEstablishmentBackoff(0, false)
	remaining, _ = ue.EstablishmentBackoff()
	assert.Zero(t, remaining)

	ue.SetEstablishmentBackoff(0, true)
	_, deactivated = ue.EstablishmentBackoff()
	assert.True(t, deactivated)

	// 6.4.1.4.2 b): a later value, even zero, does not lift a deactivated back-off.
	ue.SetEstablishmentBackoff(0, false)
	_, deactivated = ue.EstablishmentBackoff()
	assert.True(t, deactivated)
	ue.SetEstablishmentBackoff(time.Minute, false)
	_, deactivated = ue.EstablishmentBackoff()
	assert.True(t, deactivated)
}
