/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"testing"
	"time"

	"github.com/free5gc/nas/ie"
	"github.com/free5gc/nas/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rejectWithBackoff(id, cause uint8, unit ie.GPRSTimer3UnitType, value uint8) *message.PDUSessEstRej {
	r := reject(id)
	r.Cause5GSM.Value = cause
	r.BackoffTimerValue = &ie.GPRSTimer3{Unit: unit, Value: value}
	return r
}

func TestDecodeBackoffTimer(t *testing.T) {
	for _, c := range []struct {
		unit ie.GPRSTimer3UnitType
		want time.Duration
	}{
		{ie.TimerIncIn_2Seconds, 6 * time.Second},
		{ie.TimerIncIn_30Seconds, 90 * time.Second},
		{ie.TimerIncIn_1Minute, 3 * time.Minute},
		{ie.TimerIncIn_10Minutes, 30 * time.Minute},
		{ie.TimerIncIn_1Hour, 3 * time.Hour},
		{ie.TimerIncIn_10Hours, 30 * time.Hour},
	} {
		wait, deactivated, ok := decodeBackoffTimer(ie.GPRSTimer3{Unit: c.unit, Value: 3})
		assert.True(t, ok, "unit %d", c.unit)
		assert.False(t, deactivated, "unit %d", c.unit)
		assert.Equal(t, c.want, wait, "unit %d", c.unit)
	}

	_, deactivated, ok := decodeBackoffTimer(ie.GPRSTimer3{Unit: ie.TimerDeactivated})
	assert.True(t, ok)
	assert.True(t, deactivated)

	// TS 24.008 table 10.5.163a NOTE 1: the 320-hour unit is not a back-off unit.
	_, _, ok = decodeBackoffTimer(ie.GPRSTimer3{Unit: ie.TimerIncIn_320Hours, Value: 3})
	assert.False(t, ok, "unit 110 in a Back-off timer value IE is not read as 320 hours")
}

// Each cause starts the timer of its own family (TS 24.501 6.4.1.4.2, 6.4.1.4.3), and the
// causes of 6.4.1.4.3 that ignore the IE start none.
func TestBackoffTimerFor(t *testing.T) {
	for cause, want := range map[uint8]context.BackoffTimer{
		26: context.BackoffT3396,
		67: context.BackoffT3584,
		69: context.BackoffT3585,
		27: context.BackoffDNN,
		31: context.BackoffDNNAndSNSSAI,
		33: context.BackoffDNNAndSNSSAI,
	} {
		timer, ok := backoffTimerFor(cause)
		assert.True(t, ok, "cause %d", cause)
		assert.Equal(t, want, timer, "cause %d", cause)
	}
	for _, cause := range []uint8{28, 39, 46, 50, 51, 54, 57, 58, 61, 68, 86} {
		_, ok := backoffTimerFor(cause)
		assert.False(t, ok, "cause %d ignores the IE", cause)
	}
}

func TestRecordNetworkBackoffIsOptIn(t *testing.T) {
	r := rejectWithBackoff(1, 26, ie.TimerIncIn_1Hour, 1)

	for _, value := range []string{"", "true"} {
		t.Setenv("PR_HONOUR_BACKOFF", value)
		ue := newTestUE(t)
		recordNetworkBackoff(ue, r)
		remaining, _ := ue.EstablishmentBackoff()
		assert.Zero(t, remaining, "PR_HONOUR_BACKOFF=%q keeps the local schedule", value)
	}

	t.Setenv("PR_HONOUR_BACKOFF", "1")
	ue := newTestUE(t)
	recordNetworkBackoff(ue, r)
	remaining, _ := ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute)
}

// A cause 67 zero must not stop T3396: with sessions 1 and 2 outstanding, cause 26 starts
// an hour of T3396, and cause 67 with a zero timer then stops T3584 only. The UE stays held.
func TestRejectStopsOnlyTheTimerOfItsCause(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	ue := newTestUE(t)
	first, err := ue.CreatePDUSession()
	require.NoError(t, err)
	second, err := ue.CreatePDUSession()
	require.NoError(t, err)
	first.SetStateSM_PDU_SESSION_PENDING()
	second.SetStateSM_PDU_SESSION_PENDING()

	handleEstablishmentReject(ue, rejectWithBackoff(first.Id, 26, ie.TimerIncIn_1Hour, 1))
	handleEstablishmentReject(ue, rejectWithBackoff(second.Id, 67, ie.TimerIncIn_2Seconds, 0))

	remaining, deactivated := ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute, "T3396 still runs")
	assert.False(t, deactivated)
}

func TestRejectWithADeactivatedTimerDeletesTheSession(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	handleEstablishmentReject(ue, rejectWithBackoff(session.Id, 26, ie.TimerDeactivated, 0))

	_, deactivated := ue.EstablishmentBackoff()
	assert.True(t, deactivated)
	_, err = ue.GetPduSession(session.Id)
	assert.Error(t, err, "the session is deleted")
	assert.Zero(t, session.T3580Retries, "no retry is scheduled after a deactivated timer")
}

// From the wire: a reject with cause 26 and a Back-off timer value IE (tag 0x37, one
// hour) starts T3396.
func TestDlNasTransportRejectRecordsTheNetworkBackoff(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26, 0x37, 0x01, 0x21)))

	remaining, _ := ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute)
}

// From the wire: the same IE with the 320-hour unit (110) is not a back-off, so nothing is
// recorded and the local schedule applies.
func TestDlNasTransportRejectIgnoresThe320HourUnit(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26, 0x37, 0x01, 0xc1)))

	remaining, deactivated := ue.EstablishmentBackoff()
	assert.Zero(t, remaining)
	assert.False(t, deactivated)
	assert.Equal(t, 1, session.T3580Retries, "the local retry is scheduled")
}

// A retry woken while a timer runs is held again, and sent once the timer ends. It was
// counted once, when it was scheduled.
func TestRetryHeldByTheBackoffIsCountedOnce(t *testing.T) {
	ue := newTestUE(t)
	uplink := make(chan gnbcontext.UEMessage, 2)
	ue.SetGnbRx(uplink)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	handleEstablishmentReject(ue, reject(session.Id))
	retry := awaitHandOff(t, ue)
	ue.SetEstablishmentBackoff(context.BackoffT3584, 50*time.Millisecond, false)
	retry()
	assert.Empty(t, uplink, "held while the back-off runs")

	awaitHandOff(t, ue)()
	assert.Len(t, uplink, 1)
	assert.Equal(t, 1, session.T3580Retries)
}

func awaitHandOff(t *testing.T, ue *context.UEContext) func() {
	t.Helper()
	select {
	case f := <-ue.Deferred():
		return f
	case <-time.After(3 * time.Second):
		t.Fatal("work should be handed to the UE's goroutine")
		return nil
	}
}

// The network's back-off applies to the UE whatever happens to the rejected session, so
// it is recorded for an unknown session and for one that has used its retries.
func TestRejectRecordsTheBackoffForAnySession(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")

	unknown := newTestUE(t)
	handleEstablishmentReject(unknown, rejectWithBackoff(3, 26, ie.TimerIncIn_1Hour, 1))
	remaining, _ := unknown.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute, "for an unknown session")

	exhausted := newTestUE(t)
	session, err := exhausted.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	session.T3580Retries = maxRejectRetries
	handleEstablishmentReject(exhausted, rejectWithBackoff(session.Id, 26, ie.TimerIncIn_1Hour, 1))
	remaining, _ = exhausted.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute, "for a session that has used its retries")
}

// A held retry is skipped if its session was accepted while it waited, as a retry is.
func TestHeldRetryIsSkippedOnceTheSessionIsAccepted(t *testing.T) {
	ue := newTestUE(t)
	uplink := make(chan gnbcontext.UEMessage, 2)
	ue.SetGnbRx(uplink)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	handleEstablishmentReject(ue, reject(session.Id))
	retry := awaitHandOff(t, ue)
	ue.SetEstablishmentBackoff(context.BackoffT3584, 50*time.Millisecond, false)
	retry()
	session.SetStateSM_PDU_SESSION_ACTIVE()

	awaitHandOff(t, ue)()
	assert.Empty(t, uplink, "an accepted session is not requested again")
	assert.Equal(t, context.SM5G_PDU_SESSION_ACTIVE, session.GetStateSM())
}
