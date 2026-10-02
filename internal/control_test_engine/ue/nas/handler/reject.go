/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
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

// requestPduSession sends the retry of a rejected session. A variable so that tests can
// observe the retries without a NAS security context.
var requestPduSession = trigger.InitPduSessionRequestInner

// handleEstablishmentReject requests a rejected PDU session again after an exponential
// backoff (T3580), at most five times. The retry is counted here, on the UE's goroutine,
// and sent from the UE's goroutine too: the UE's NAS count is read and updated without a
// lock, so a request sent from another goroutine could take the same count as a message
// the UE sends meanwhile. A retry still waiting when the UE terminates is not sent.
func handleEstablishmentReject(ue *context.UEContext, reject *message.PDUSessEstRej) {
	pduSessionId := reject.PDUSessId
	pduSession, err := ue.GetPduSession(pduSessionId)

	switch {
	case err != nil:
		log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
	case pduSession.T3580Retries >= maxRejectRetries:
		log.Error("[UE][NAS] We re-tried five times to create PDU Session ", pduSessionId, ", Aborting.")
	default:
		wait := time.Duration(math.Pow(5, float64(pduSession.T3580Retries))) * time.Second
		pduSession.T3580Retries++
		ue.RunOnUEAfter(wait, func() { requestPduSession(ue, pduSession) })
	}
}
