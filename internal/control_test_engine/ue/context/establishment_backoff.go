/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import "time"

// BackoffTimer is one of the timers a PDU Session Establishment Reject can start to hold
// back further establishment requests (TS 24.501 6.4.1.4.2, 6.4.1.4.3). Each runs, stops
// and is deactivated on its own: a reject sets only the timer its cause belongs to.
type BackoffTimer int

const (
	// BackoffT3396 is started by cause 26, per DNN (6.4.1.4.2).
	BackoffT3396 BackoffTimer = iota
	// BackoffT3584 is started by cause 67, per [S-NSSAI, DNN] (6.4.1.4.2).
	BackoffT3584
	// BackoffT3585 is started by cause 69, per S-NSSAI (6.4.1.4.2).
	BackoffT3585
	// BackoffDNN is the back-off timer cause 27 starts, per [PLMN, DNN] (6.4.1.4.3).
	BackoffDNN
	// BackoffDNNAndSNSSAI is the back-off timer the other causes of 6.4.1.4.3 start, per
	// [PLMN, DNN, S-NSSAI].
	BackoffDNNAndSNSSAI

	backoffTimers
)

type backoffState struct {
	until       time.Time
	deactivated bool
}

// SetEstablishmentBackoff applies the Back-off timer value IE of a reject to one timer: it
// runs for wait, is stopped by a zero wait, or is deactivated. Once deactivated, a timer
// stays so: switch-off or USIM removal would lift it, and for the congestion timers also a
// PDU Session Modification or Authentication Command, or a Release Command without the
// IE; none of those is modelled. Like the rest of the PDU session state, the timers are
// only touched on the UE's goroutine.
func (ue *UEContext) SetEstablishmentBackoff(timer BackoffTimer, wait time.Duration, deactivated bool) {
	state := &ue.backoff[timer]
	if state.deactivated {
		return
	}
	state.deactivated = deactivated
	state.until = time.Time{}
	if !deactivated && wait > 0 {
		state.until = time.Now().Add(wait)
	}
}

// EstablishmentBackoff reports how long a further PDU session establishment request must
// still wait, and whether a timer forbids it outright. Every timer applies to every request
// of this UE, which has one DNN and one S-NSSAI, so a request waits for the longest and any
// deactivated timer forbids it.
func (ue *UEContext) EstablishmentBackoff() (remaining time.Duration, deactivated bool) {
	for _, state := range ue.backoff {
		deactivated = deactivated || state.deactivated
		remaining = max(remaining, time.Until(state.until))
	}
	return remaining, deactivated
}
