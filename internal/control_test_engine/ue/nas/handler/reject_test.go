/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUE() *context.UEContext {
	capability := nasType.NewUESecurityCapability(nasMessage.RegistrationRequestUESecurityCapabilityType)
	capability.SetLen(2)
	capability.Buffer = []uint8{0x80, 0x80}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	return ue
}

// countRetries replaces the retry send for the length of a test.
func countRetries(t *testing.T) *[]*context.UEPDUSession {
	var sent []*context.UEPDUSession
	prev := requestPduSession
	requestPduSession = func(_ *context.UEContext, s *context.UEPDUSession) { sent = append(sent, s) }
	t.Cleanup(func() { requestPduSession = prev })
	return &sent
}

func reject(id uint8) *nasMessage.PDUSessionEstablishmentReject {
	r := nasMessage.NewPDUSessionEstablishmentReject(0)
	r.SetPDUSessionID(id)
	r.SetCauseValue(nasMessage.Cause5GSMInsufficientResources)
	return r
}

func noHandOff(t *testing.T, ue *context.UEContext, why string) {
	t.Helper()
	select {
	case <-ue.Deferred():
		t.Fatal(why)
	case <-time.After(200 * time.Millisecond):
	}
}

// dlReject wraps a 5GSM message in a plain DL NAS Transport for PDU session id, as the
// AMF sends a PDU Session Establishment Reject.
func dlReject(id uint8, gsm []byte) *nas.Message {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeDLNASTransport)
	d := nasMessage.NewDLNASTransport(0)
	d.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	d.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	d.SetMessageType(nas.MsgTypeDLNASTransport)
	d.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(nasMessage.PayloadContainerTypeN1SMInfo)
	d.PayloadContainer.SetLen(uint16(len(gsm)))
	d.PayloadContainer.SetPayloadContainerContents(gsm)
	d.PduSessionID2Value = new(nasType.PduSessionID2Value)
	d.PduSessionID2Value.SetIei(nasMessage.DLNASTransportPduSessionID2ValueType)
	d.PduSessionID2Value.SetPduSessionID2Value(id)
	m.GmmMessage.DLNASTransport = d
	return m
}

// rejectBytes is a PDU Session Establishment Reject on the wire: EPD 5GSM (0x2e), the PDU
// session ID, PTI 1, message type 0xc3 and the 5GSM cause, then any optional IEs.
func rejectBytes(id, cause uint8, ies ...byte) []byte {
	return append([]byte{0x2e, id, 0x01, nas.MsgTypePDUSessionEstablishmentReject, cause}, ies...)
}

// A reject arriving in a DL NAS Transport reaches the retry.
func TestDlNasTransportRejectSchedulesTheRetry(t *testing.T) {
	countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)

	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26)))

	assert.Equal(t, 1, session.T3580Retries)
}

// The retry is counted when scheduled, handed to the UE's goroutine after its 1 s
// backoff, and sent only when that goroutine runs it.
func TestHandleEstablishmentRejectRetriesOnTheUEGoroutine(t *testing.T) {
	sent := countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)

	handleEstablishmentReject(ue, reject(session.Id))
	assert.Equal(t, 1, session.T3580Retries, "counted when scheduled")

	select {
	case f := <-ue.Deferred():
		assert.Empty(t, *sent, "nothing is sent until the UE's goroutine runs it")
		f()
	case <-time.After(3 * time.Second):
		t.Fatal("the retry should be handed to the UE's goroutine after its 1 s backoff")
	}
	assert.Equal(t, []*context.UEPDUSession{session}, *sent)
}

func TestHandleEstablishmentRejectStopsAfterFiveRetries(t *testing.T) {
	countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.T3580Retries = maxRejectRetries

	handleEstablishmentReject(ue, reject(session.Id))

	assert.Equal(t, maxRejectRetries, session.T3580Retries)
	noHandOff(t, ue, "no sixth retry")
}

// A reject carrying the Back-off timer value IE (tag 0x37, 1 hour) records the network's
// back-off on the UE, from the wire.
func TestDlNasTransportRejectRecordsTheNetworkBackoff(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)

	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26, 0x37, 0x01, 0x21)))

	assert.Equal(t, 1, session.T3580Retries)
	remaining, _ := ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute)
}

// With the network's timer the wait is the network's, and it holds the UE's other
// requests too.
func TestHandleEstablishmentRejectFollowsTheNetworkTimer(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	sent := countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)

	handleEstablishmentReject(ue, withBackoff(reject(session.Id), nasMessage.GPRSTimer3UnitMultiplesOf1Hour, 1))

	remaining, deactivated := ue.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute, "recorded on the UE")
	assert.False(t, deactivated)
	assert.Equal(t, 1, session.T3580Retries)
	noHandOff(t, ue, "the retry waits for the network's hour")
	assert.Empty(t, *sent)
}

func TestHandleEstablishmentRejectDeactivatedBlocksTheUE(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	sent := countRetries(t)
	ue := newTestUE()
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)

	handleEstablishmentReject(ue, withBackoff(reject(session.Id), gprsTimer3UnitDeactivated, 0))

	_, deactivated := ue.EstablishmentBackoff()
	assert.True(t, deactivated, "the UE's further requests are blocked")
	_, err = ue.GetPduSession(session.Id)
	assert.Error(t, err, "the session's slot is freed")
	noHandOff(t, ue, "nothing should be scheduled after a deactivated timer")
	assert.Empty(t, *sent)
}

// A later reject's timer does not lift a deactivated back-off: that is not one of the
// events that lift it (TS 24.501 6.4.1.4.2 b)).
func TestHandleEstablishmentRejectKeepsTheUEBlockedAfterALaterTimer(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	countRetries(t)
	ue := newTestUE()
	first, err := ue.CreatePDUSession()
	require.NoError(t, err)
	second, err := ue.CreatePDUSession()
	require.NoError(t, err)

	handleEstablishmentReject(ue, withBackoff(reject(first.Id), gprsTimer3UnitDeactivated, 0))
	handleEstablishmentReject(ue, withBackoff(reject(second.Id), nasMessage.GPRSTimer3UnitMultiplesOf1Minute, 1))

	_, deactivated := ue.EstablishmentBackoff()
	assert.True(t, deactivated)
}

// The network's back-off applies to the UE whatever happens to the rejected session.
func TestHandleEstablishmentRejectRecordsTheBackoffForAnySession(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	countRetries(t)

	unknown := newTestUE()
	handleEstablishmentReject(unknown, withBackoff(reject(3), nasMessage.GPRSTimer3UnitMultiplesOf1Hour, 1))
	remaining, _ := unknown.EstablishmentBackoff()
	assert.Greater(t, remaining, 59*time.Minute, "for an unknown session")

	exhausted := newTestUE()
	session, err := exhausted.CreatePDUSession()
	require.NoError(t, err)
	session.T3580Retries = maxRejectRetries
	handleEstablishmentReject(exhausted, withBackoff(reject(session.Id), gprsTimer3UnitDeactivated, 0))
	_, deactivated := exhausted.EstablishmentBackoff()
	assert.True(t, deactivated, "for a session that has used its retries")
}
