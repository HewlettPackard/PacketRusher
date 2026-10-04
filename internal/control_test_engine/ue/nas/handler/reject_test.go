/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 * © Copyright 2026 Valentin D'Emmanuele
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

// A rejected session is requested again by the UE's goroutine after its 1 s backoff,
// unless it was released meanwhile.
func TestRejectedSessionIsRetriedUnlessReleased(t *testing.T) {
	for _, released := range []bool{false, true} {
		ue := newTestUE(t)
		uplink := make(chan gnbcontext.UEMessage, 1)
		ue.SetGnbRx(uplink)
		session, err := ue.CreatePDUSession()
		require.NoError(t, err)
		session.SetStateSM_PDU_SESSION_PENDING()

		HandlerDlNasTransportPduaccept(ue, dlReject(session.Id, rejectBytes(session.Id, 26)))
		assert.Equal(t, 1, session.T3580Retries)
		if released {
			require.NoError(t, ue.DeletePduSession(session.Id))
		}

		select {
		case retry := <-ue.Deferred():
			assert.Empty(t, uplink, "nothing is sent until the UE's goroutine runs the retry")
			retry()
		case <-time.After(3 * time.Second):
			t.Fatal("the retry should be handed to the UE's goroutine after its 1 s backoff")
		}
		assert.Equal(t, released, len(uplink) == 0, "released: %t", released)
	}
}

func TestRejectedSessionIsNotRetriedAfterFiveRetries(t *testing.T) {
	ue := newTestUE(t)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	session.T3580Retries = maxRejectRetries

	handleEstablishmentReject(ue, reject(session.Id))

	assert.Equal(t, maxRejectRetries, session.T3580Retries, "no sixth retry is scheduled")
}
