// SPDX-License-Identifier: Apache-2.0
package context

import (
	"my5G-RANTester/internal/analytics"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"time"

	log "github.com/sirupsen/logrus"
)

// PduSessionRetry identifies one scheduled retry, including its session lifetime.
// Only the UE event loop executes it; the timer never encodes or sends NAS.
type PduSessionRetry struct {
	session    *UEPDUSession
	generation uint64
}

func (retry PduSessionRetry) Session() *UEPDUSession { return retry.session }

func (ue *UEContext) pduSessionRetriesLocked() chan PduSessionRetry {
	if ue.pduSessionRetries == nil {
		ue.pduSessionRetries = make(chan PduSessionRetry, len(ue.PduSession))
	}
	return ue.pduSessionRetries
}

func (ue *UEContext) PduSessionRetries() <-chan PduSessionRetry {
	ue.Lock()
	defer ue.Unlock()
	return ue.pduSessionRetriesLocked()
}

// hasSessionLocked distinguishes a current session from a deleted session whose
// ID has since been reused. All retry lifetime checks use the UE lock.
func (ue *UEContext) hasSessionLocked(session *UEPDUSession) bool {
	return !ue.terminated && session != nil && session.Id >= 1 &&
		int(session.Id) <= len(ue.PduSession) && ue.PduSession[session.Id-1] == session
}

func (session *UEPDUSession) cancelRetryLocked() {
	session.retryGeneration++
	if session.retryTimer != nil {
		session.retryTimer.Stop()
		session.retryTimer = nil
	}
	if session.retryCancel != nil {
		close(session.retryCancel)
		session.retryCancel = nil
	}
}

// StartPduSessionRequest validates ownership, builds the request, starts its
// measurement and sends it while holding the UE lock. A deleted or terminated
// session cannot begin another attempt, even if its timer callback already ran.
// Call this from the UE event loop; the encoder must not acquire the UE lock.
func (ue *UEContext) StartPduSessionRequest(session *UEPDUSession, encode func() ([]byte, error)) error {
	ue.Lock()
	defer ue.Unlock()
	if !ue.hasSessionLocked(session) {
		return nil
	}
	return ue.startPduSessionRequestLocked(session, encode)
}

func (ue *UEContext) startPduSessionRequestLocked(session *UEPDUSession, encode func() ([]byte, error)) error {
	if !ue.gnbConnectionOpenLocked() {
		log.Warn("[UE] Do not start a PDU session request as the gNB channel is closed")
		return nil
	}
	payload, err := encode()
	if err != nil {
		return err
	}
	session.setPendingLocked()
	select {
	case ue.gnbRx <- gnbcontext.UEMessage{IsNas: true, Nas: payload}:
	case <-ue.gnbConnectionLost:
		session.results.Finish(session.resultsUE, session.Id, analytics.SessionEstablishment, analytics.Cancelled)
		log.Warn("[UE] Cancelled a PDU session request after the gNB association failed")
	}
	return nil
}

// An association can fail while the UE event loop waits on a full gNB channel.
// Its out-of-band signal releases the send before Terminate acquires the UE lock.
func (ue *UEContext) gnbConnectionOpenLocked() bool {
	if ue.gnbRx == nil {
		return false
	}
	select {
	case <-ue.gnbConnectionLost:
		return false
	default:
		return true
	}
}

// SchedulePduSessionRetry schedules at most one retry for the current session,
// with the existing exponential backoff and five-retry limit.
func (ue *UEContext) SchedulePduSessionRetry(session *UEPDUSession) bool {
	ue.Lock()
	defer ue.Unlock()
	if !ue.hasSessionLocked(session) || session.StateSM != SM5G_PDU_SESSION_ACTIVE_PENDING ||
		session.retryCancel != nil || session.T3580Retries >= 5 {
		return false
	}
	delay := time.Second
	for i := 0; i < session.T3580Retries; i++ {
		delay *= 5
	}
	ue.schedulePduSessionRetryLocked(session, delay)
	return true
}

func (ue *UEContext) schedulePduSessionRetryLocked(session *UEPDUSession, delay time.Duration) {
	session.retryGeneration++
	generation := session.retryGeneration
	cancel := make(chan struct{})
	session.retryCancel = cancel
	queue := ue.pduSessionRetriesLocked()
	session.retryTimer = time.AfterFunc(delay, func() {
		select {
		case queue <- PduSessionRetry{session: session, generation: generation}:
		case <-cancel:
		}
	})
}

// StartPduSessionRetry must be called from the UE event loop. This serializes NAS
// encoding and its security counters with incoming NAS processing. Cancellation
// invalidates queued tokens as well as stopping timers that have not expired.
func (ue *UEContext) StartPduSessionRetry(retry PduSessionRetry, encode func() ([]byte, error)) error {
	ue.Lock()
	defer ue.Unlock()
	session := retry.session
	if !ue.hasSessionLocked(session) || session.retryGeneration != retry.generation || session.retryCancel == nil {
		return nil
	}
	session.cancelRetryLocked()
	if !ue.gnbConnectionOpenLocked() {
		log.Warn("[UE] Do not retry a PDU session request as the gNB channel is closed")
		return nil
	}
	session.T3580Retries++
	return ue.startPduSessionRequestLocked(session, encode)
}
