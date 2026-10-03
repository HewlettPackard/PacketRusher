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

	"github.com/free5gc/nas/ie"
	"github.com/free5gc/nas/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUE(t *testing.T) *context.UEContext {
	t.Helper()
	capability := &ie.UESecCapability{Length: 2, EA05G: true, IA05G: true}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	t.Cleanup(ue.Terminate)
	return ue
}

func reject(id uint8) *message.PDUSessEstRej {
	return &message.PDUSessEstRej{PDUSessId: id, PTI: 1, Cause5GSM: &ie.Cause5GSM{Value: 26}}
}

func awaitRetry(t *testing.T, ue *context.UEContext) context.PduSessionRetry {
	t.Helper()
	select {
	case retry := <-ue.PduSessionRetries():
		return retry
	case <-time.After(3 * time.Second):
		t.Fatal("reject did not queue a retry for the UE event loop")
		return context.PduSessionRetry{}
	}
}

// dlReject wraps a 5GSM message in a plain DL NAS Transport for PDU session id, as the
// AMF sends a PDU Session Establishment Reject.
func dlReject(id uint8, gsm []byte) *message.DLNASTransport {
	return &message.DLNASTransport{
		PayloadCntrType: &ie.PayloadCntrType{Value: ie.PayloadCntrType_N1SMInfo},
		PayloadCntr:     &ie.PayloadCntr{Pct: ie.PayloadCntrType_N1SMInfo, Contents: gsm},
		PDUSessID:       &ie.PDUSessId2{Value: id},
	}
}

// rejectBytes is a PDU Session Establishment Reject on the wire: EPD 5GSM (0x2e), the PDU
// session ID, PTI 1, message type 0xc3 and the 5GSM cause, then any optional IEs.
func rejectBytes(id, cause uint8, ies ...byte) []byte {
	return append([]byte{0x2e, id, 0x01, byte(message.MsgTypePDUSessEstRej), cause}, ies...)
}

// A real 5GSM reject carried by DL NAS Transport reaches the typed retry queue.
func TestDlNasTransportRejectSchedulesTheRetry(t *testing.T) {
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26)))
	retry := awaitRetry(t, ue)
	require.Same(t, session, retry.Session())
	assert.Zero(t, session.T3580Retries, "queueing must not count an unexecuted request")
}

// Timer expiry hands over a token, not an encoder. Sending and counting happen
// only when the UE event loop validates and executes that token.
func TestHandleEstablishmentRejectRetriesOnTheUEGoroutine(t *testing.T) {
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	ue.SetGnbRx(make(chan gnbcontext.UEMessage, 1))
	handleEstablishmentReject(ue, reject(session.Id))
	retry := awaitRetry(t, ue)
	assert.Zero(t, session.T3580Retries)
	assert.Empty(t, ue.GetGnbRx(), "the timer must not send NAS")
	encoded := false
	require.NoError(t, ue.StartPduSessionRetry(retry, func() ([]byte, error) {
		encoded = true
		return []byte{0x42}, nil
	}))
	assert.True(t, encoded)
	assert.Equal(t, 1, session.T3580Retries)
	select {
	case msg := <-ue.GetGnbRx():
		assert.Equal(t, []byte{0x42}, msg.Nas)
	default:
		t.Fatal("the event loop did not send its retry")
	}
}

func TestHandleEstablishmentRejectStopsAfterFiveRetries(t *testing.T) {
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.SetStateSM_PDU_SESSION_PENDING()
	session.T3580Retries = 5
	handleEstablishmentReject(ue, reject(session.Id))
	assert.Equal(t, 5, session.T3580Retries)
	select {
	case <-ue.PduSessionRetries():
		t.Fatal("no sixth retry may be queued")
	case <-time.After(200 * time.Millisecond):
	}
}
