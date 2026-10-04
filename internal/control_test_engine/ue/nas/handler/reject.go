/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"math"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/trigger"
	"time"

	"github.com/free5gc/nas/message"
	log "github.com/sirupsen/logrus"
)

// maxRejectRetries is how many times a rejected PDU session is requested again.
const maxRejectRetries = 5

// handleEstablishmentReject requests a rejected PDU session again after an exponential
// backoff (T3580), at most five times. The retry is sent from the UE's goroutine, as the
// UE's NAS count is read and updated without a lock. It is skipped if the session was
// released, replaced or accepted while it waited, or if the UE is no longer connected to
// a gNB. A retry still waiting when the UE terminates is not sent.
func handleEstablishmentReject(ue *context.UEContext, reject *message.PDUSessEstRej) {
	pduSessionId := reject.PDUSessId
	pduSession, err := ue.GetPduSession(pduSessionId)
	if err != nil {
		log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
		return
	}
	pduSession.EstablishmentFailed()
	if pduSession.T3580Retries >= maxRejectRetries {
		log.Error("[UE][NAS] We re-tried five times to create PDU Session ", pduSessionId, ", Aborting.")
		return
	}
	wait := time.Duration(math.Pow(5, float64(pduSession.T3580Retries))) * time.Second
	pduSession.T3580Retries++
	ue.RunOnUEAfter(wait, func() {
		current, _ := ue.GetPduSession(pduSessionId)
		if current != pduSession || pduSession.GetStateSM() != context.SM5G_PDU_SESSION_ACTIVE_PENDING || ue.GetGnbRx() == nil {
			return
		}
		trigger.InitPduSessionRequestInner(ue, pduSession)
	})
}
