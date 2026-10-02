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

// handleEstablishmentReject records a failed attempt and schedules a retry on the
// UE goroutine. The typed token preserves session lifetime and cancellation; the
// retry count advances only when that goroutine executes an eligible attempt.
func handleEstablishmentReject(ue *context.UEContext, reject *nasMessage.PDUSessionEstablishmentReject) {
	pduSessionId := reject.GetPDUSessionID()
	pduSession, err := ue.GetPduSession(pduSessionId)

	if err != nil {
		log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
		return
	}
	pduSession.EstablishmentFailed()
	trigger.RetryPduSessionRequest(ue, pduSession)
}
