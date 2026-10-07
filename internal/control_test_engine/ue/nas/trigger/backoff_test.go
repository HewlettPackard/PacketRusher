/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package trigger

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"

	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUE(t *testing.T) (*context.UEContext, chan gnbcontext.UEMessage) {
	t.Helper()
	capability := &ie.UESecCapability{Length: 2, EA05G: true, IA05G: true}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	t.Cleanup(ue.Terminate)
	uplink := make(chan gnbcontext.UEMessage, 4)
	ue.SetGnbRx(uplink)
	return ue, uplink
}

func handedOff(t *testing.T, ue *context.UEContext) func() {
	t.Helper()
	select {
	case f := <-ue.Deferred():
		return f
	case <-time.After(2 * time.Second):
		t.Fatal("the held request should be handed to the UE's goroutine when the back-off ends")
		return nil
	}
}

func TestPduSessionRequestIsSentWithoutBackoff(t *testing.T) {
	ue, uplink := newTestUE(t)

	InitPduSessionRequest(ue)

	assert.Len(t, uplink, 1)
}

// A first request held by a running timer goes out when the timer ends, without using
// one of the session's retries.
func TestHeldRequestIsSentWhenTheBackoffEnds(t *testing.T) {
	ue, uplink := newTestUE(t)
	ue.SetEstablishmentBackoff(context.BackoffT3396, 50*time.Millisecond, false)

	InitPduSessionRequest(ue)
	assert.Empty(t, uplink, "held while the back-off runs")

	handedOff(t, ue)()
	assert.Len(t, uplink, 1)
	session, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Zero(t, session.T3580Retries, "a held first request is not a retry")
}

// A held request whose session is deleted, and whose ID a new session takes, must neither
// send nor touch the new session, whether that happens before the timer ends or after the
// request has been handed to the UE's goroutine.
func TestHeldRequestIsDiscardedAfterItsSessionIsReplaced(t *testing.T) {
	for _, afterWake := range []bool{false, true} {
		ue, uplink := newTestUE(t)
		ue.SetEstablishmentBackoff(context.BackoffT3396, 50*time.Millisecond, false)
		InitPduSessionRequest(ue)
		first, err := ue.GetPduSession(1)
		require.NoError(t, err)

		var held func()
		if afterWake {
			held = handedOff(t, ue)
		}
		require.NoError(t, ue.DeletePduSession(first.Id))
		replacement, err := ue.CreatePDUSession()
		require.NoError(t, err)
		require.Equal(t, first.Id, replacement.Id, "fixture must reuse the session ID")
		if !afterWake {
			held = handedOff(t, ue)
		}

		held()

		assert.Empty(t, uplink, "after wake: %t", afterWake)
		current, err := ue.GetPduSession(replacement.Id)
		require.NoError(t, err)
		assert.Same(t, replacement, current, "the new session keeps its slot")
		assert.Equal(t, context.SM5G_PDU_SESSION_INACTIVE, replacement.GetStateSM(), "and its state")
	}
}

// A request a deactivated timer forbids is dropped, and its session deleted.
func TestRequestForbiddenByADeactivatedTimerIsDropped(t *testing.T) {
	ue, uplink := newTestUE(t)
	ue.SetEstablishmentBackoff(context.BackoffDNNAndSNSSAI, 0, true)

	InitPduSessionRequest(ue)

	assert.Empty(t, uplink)
	_, err := ue.GetPduSession(1)
	assert.Error(t, err, "the forbidden session is deleted")
	select {
	case <-ue.Deferred():
		t.Fatal("nothing should be scheduled for a forbidden request")
	case <-time.After(100 * time.Millisecond):
	}
}

// A held request is checked again when its wait ends, so a back-off extended in the
// meantime keeps it held.
func TestHeldRequestStaysHeldWhenTheBackoffIsExtended(t *testing.T) {
	ue, uplink := newTestUE(t)
	ue.SetEstablishmentBackoff(context.BackoffT3396, 50*time.Millisecond, false)
	InitPduSessionRequest(ue)
	ue.SetEstablishmentBackoff(context.BackoffT3396, 300*time.Millisecond, false)

	handedOff(t, ue)()
	assert.Empty(t, uplink, "still held: the back-off was extended")

	handedOff(t, ue)()
	assert.Len(t, uplink, 1)
}

// A held request that wakes after the UE lost its gNB connection is not sent, and its
// session is left as it was.
func TestHeldRequestIsNotSentWithoutAGnbConnection(t *testing.T) {
	ue, uplink := newTestUE(t)
	ue.SetEstablishmentBackoff(context.BackoffT3396, 50*time.Millisecond, false)
	InitPduSessionRequest(ue)
	held := handedOff(t, ue)
	ue.SetGnbRx(nil)

	held()

	assert.Empty(t, uplink)
	session, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Equal(t, context.SM5G_PDU_SESSION_INACTIVE, session.GetStateSM(), "not requested")
}
