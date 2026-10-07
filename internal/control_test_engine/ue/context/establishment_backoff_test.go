/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"testing"
	"time"
)

// A zero value stops only the timer of its own cause: a T3584 stop must not lift a
// running T3396 (TS 24.501 6.4.1.4.2).
func TestBackoffTimersAreKeptApart(t *testing.T) {
	ue := &UEContext{}
	ue.SetEstablishmentBackoff(BackoffT3396, time.Hour, false)
	ue.SetEstablishmentBackoff(BackoffT3584, 0, false)

	if remaining, _ := ue.EstablishmentBackoff(); remaining < 59*time.Minute {
		t.Fatalf("T3396 should still hold requests, remaining %v", remaining)
	}

	for timer := BackoffTimer(0); timer < backoffTimers; timer++ {
		ue := &UEContext{}
		ue.SetEstablishmentBackoff(timer, time.Hour, false)
		if remaining, _ := ue.EstablishmentBackoff(); remaining < 59*time.Minute {
			t.Fatalf("timer %d should hold requests while it runs", timer)
		}
		ue.SetEstablishmentBackoff(timer, 0, false)
		if remaining, _ := ue.EstablishmentBackoff(); remaining != 0 {
			t.Fatalf("a zero value should stop timer %d, remaining %v", timer, remaining)
		}
	}
}

// A deactivated timer forbids every further request and stays so, whatever a later reject
// sets for that timer or another one.
func TestDeactivatedBackoffTimerStaysSo(t *testing.T) {
	ue := &UEContext{}
	ue.SetEstablishmentBackoff(BackoffT3585, 0, true)
	ue.SetEstablishmentBackoff(BackoffT3585, time.Minute, false)
	ue.SetEstablishmentBackoff(BackoffT3585, 0, false)
	ue.SetEstablishmentBackoff(BackoffT3396, 0, false)

	if _, deactivated := ue.EstablishmentBackoff(); !deactivated {
		t.Fatal("a deactivated timer should still forbid requests")
	}
}
