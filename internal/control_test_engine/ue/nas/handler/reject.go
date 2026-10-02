/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/trigger"

	"github.com/free5gc/nas/nasMessage"
	log "github.com/sirupsen/logrus"
)

// maxRejectRetries is how many times a rejected PDU session is requested again.
const maxRejectRetries = 5

// requestPduSession sends the retry of a rejected session, through the gate on the UE's
// back-off. A variable so that tests can observe the retries without a NAS security
// context.
var requestPduSession = trigger.RequestPduSessionWhenAllowed

// handleEstablishmentReject requests a rejected PDU session again after a wait, at most
// five times. The wait is an exponential backoff (T3580), or with PR_HONOUR_BACKOFF=1 the
// network's Back-off timer value (retryAfterReject). The retry is counted here, on the
// UE's goroutine, and sent from the UE's goroutine too: the UE's NAS count is read and
// updated without a lock, so a request sent from another goroutine could take the same
// count as a message the UE sends meanwhile. A retry still waiting when the UE terminates
// is not sent.
//
// A back-off the network sent applies to every further establishment request of the UE,
// so it is recorded first, even when the session is unknown or has used its retries.
// After a deactivated timer the session's slot is freed, since it will not be requested
// again.
func handleEstablishmentReject(ue *context.UEContext, reject *nasMessage.PDUSessionEstablishmentReject) {
	pduSessionId := reject.GetPDUSessionID()
	pduSession, err := ue.GetPduSession(pduSessionId)
	retries := 0
	if err == nil {
		retries = pduSession.T3580Retries
	}
	decision := retryAfterReject(reject, retries)
	if decision.fromNetwork {
		ue.SetEstablishmentBackoff(decision.wait, !decision.retry)
	}

	switch {
	case err != nil:
		log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
	case pduSession.T3580Retries >= maxRejectRetries:
		log.Error("[UE][NAS] We re-tried five times to create PDU Session ", pduSessionId, ", Aborting.")
	case !decision.retry:
		log.Error("[UE][NAS] Network deactivated the back-off timer for PDU Session ", pduSessionId, "; not retrying")
		_ = ue.DeletePduSession(pduSessionId)
	default:
		if decision.fromNetwork {
			log.Info("[UE][NAS] Network asked us to retry PDU Session ", pduSessionId, " in ", decision.wait)
		}
		pduSession.T3580Retries++
		ue.RunOnUEAfter(decision.wait, func() { requestPduSession(ue, pduSession) })
	}
}
