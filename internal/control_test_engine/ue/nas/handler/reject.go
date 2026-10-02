/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/trigger"

	"github.com/free5gc/nas/message"
	log "github.com/sirupsen/logrus"
)

// Keep rejection handling on the UE's event loop. The timer queues a session
// token; only the event loop may validate its lifetime and encode the retry.
func handleEstablishmentReject(ue *context.UEContext, reject *message.PDUSessEstRej) {
	session, err := ue.GetPduSession(reject.PDUSessId)
	if err != nil {
		log.Error(err)
		return
	}
	session.EstablishmentFailed()
	trigger.RetryPduSessionRequest(ue, session)
}
