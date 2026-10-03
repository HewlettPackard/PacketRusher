/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
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

func reject(id uint8) *nasMessage.PDUSessionEstablishmentReject {
	r := nasMessage.NewPDUSessionEstablishmentReject(0)
	r.SetPDUSessionID(id)
	r.SetCauseValue(nasMessage.Cause5GSMInsufficientResources)
	return r
}

func noHandOff(t *testing.T, ue *context.UEContext, why string) {
	t.Helper()
	select {
	case <-ue.PduSessionRetries():
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

// A reject arriving in a DL NAS Transport reaches the cancellable retry queue.
func TestDlNasTransportRejectSchedulesTheRetry(t *testing.T) {
	ue := newTestUE()
	t.Cleanup(ue.Terminate)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()

	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26)))

	assert.Zero(t, session.T3580Retries, "a scheduled attempt is not counted until executed")
	select {
	case retry := <-ue.PduSessionRetries():
		assert.Same(t, session, retry.Session())
	case <-time.After(3 * time.Second):
		t.Fatal("the DL NAS reject should schedule its session's retry")
	}
}

// The timer hands a typed token to the UE's goroutine after its 1 s backoff.
// NAS is encoded, sent and counted only when that goroutine executes the token.
func TestHandleEstablishmentRejectRetriesOnTheUEGoroutine(t *testing.T) {
	ue := newTestUE()
	t.Cleanup(ue.Terminate)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	uplink := make(chan gnbcontext.UEMessage, 1)
	ue.SetGnbRx(uplink)

	handleEstablishmentReject(ue, reject(session.Id))
	assert.Zero(t, session.T3580Retries, "not counted when scheduled")

	select {
	case retry := <-ue.PduSessionRetries():
		assert.Empty(t, uplink, "the timer must not send NAS")
		assert.Zero(t, session.T3580Retries, "not counted by the timer")
		require.NoError(t, ue.StartPduSessionRetry(retry, func() ([]byte, error) {
			return []byte{0x7e}, nil
		}))
	case <-time.After(3 * time.Second):
		t.Fatal("the retry should be handed to the UE's goroutine after its 1 s backoff")
	}
	assert.Equal(t, 1, session.T3580Retries, "counted when the UE executes the attempt")
	assert.Equal(t, gnbcontext.UEMessage{IsNas: true, Nas: []byte{0x7e}}, <-uplink)
}

func TestHandleEstablishmentRejectStopsAfterFiveRetries(t *testing.T) {
	ue := newTestUE()
	t.Cleanup(ue.Terminate)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	session.T3580Retries = maxRejectRetries

	handleEstablishmentReject(ue, reject(session.Id))

	assert.Equal(t, maxRejectRetries, session.T3580Retries)
	noHandOff(t, ue, "no sixth retry")
}
