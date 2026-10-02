// SPDX-License-Identifier: Apache-2.0
package context

import (
	"testing"
	"time"

	"my5G-RANTester/internal/analytics"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
)

// Queued retries can outlive deletion or UE termination. Processing such a token
// after its ID was reused must not encode NAS or start a phantom attempt.
func TestQueuedSessionRetryCannotStartAfterCancellation(t *testing.T) {
	for _, termination := range []bool{false, true} {
		name := "session deletion"
		if termination {
			name = "UE termination"
		}
		t.Run(name, func(t *testing.T) {
			results := analytics.NewRecorder()
			ue := &UEContext{Results: results, scenarioChan: make(chan scenario.ScenarioMessage)}
			first, err := ue.CreatePDUSession()
			if err != nil {
				t.Fatal(err)
			}
			first.SetStateSM_PDU_SESSION_PENDING()
			first.EstablishmentFailed()
			ue.Lock()
			ue.schedulePduSessionRetryLocked(first, 0)
			ue.Unlock()
			retry := awaitSessionRetry(t, ue.PduSessionRetries())
			if termination {
				ue.Terminate()
				if _, err := ue.CreatePDUSession(); err == nil {
					t.Fatal("terminated UE accepted a new session")
				}
			} else if err := ue.DeletePduSession(first.Id); err != nil {
				t.Fatal(err)
			}
			replacementUE := ue
			if termination {
				// Registration loops create a new UE with the same reporting ID.
				replacementUE = &UEContext{Results: results}
			}
			replacement, err := replacementUE.CreatePDUSession()
			if err != nil {
				t.Fatal(err)
			}
			if replacement.Id != first.Id {
				t.Fatal("fixture must reuse the reporting session ID")
			}
			replacement.SetStateSM_PDU_SESSION_PENDING()
			replacement.SetStateSM_PDU_SESSION_ACTIVE()
			encoded := 0
			if err := ue.StartPduSessionRetry(retry, func() ([]byte, error) {
				encoded++
				return nil, nil
			}); err != nil {
				t.Fatal(err)
			}
			first.SetStateSM_PDU_SESSION_PENDING() // stale state updates are ignored too
			first.SetStateSM_PDU_SESSION_ACTIVE()
			if encoded != 0 {
				t.Fatal("cancelled retry encoded a request")
			}
			assertSessionResults(t, results, 2, 1, 0, 0)
		})
	}
}

func TestDeletingSessionCancelsScheduledRetry(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder()}
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	session.EstablishmentFailed()
	ue.Lock()
	ue.schedulePduSessionRetryLocked(session, 20*time.Millisecond)
	ue.Unlock()
	if ue.SchedulePduSessionRetry(session) {
		t.Fatal("duplicate rejection scheduled a second pending retry")
	}
	if err := ue.DeletePduSession(session.Id); err != nil {
		t.Fatal(err)
	}
	select {
	case retry := <-ue.PduSessionRetries():
		// If expiration raced deletion, the already queued token must be inert.
		if err := ue.StartPduSessionRetry(retry, func() ([]byte, error) {
			t.Error("deleted session encoded a retry")
			return nil, nil
		}); err != nil {
			t.Fatal(err)
		}
	case <-time.After(40 * time.Millisecond):
	}
	if ue.SchedulePduSessionRetry(session) {
		t.Fatal("deleted session accepted another retry")
	}
	assertSessionResults(t, ue.Results, 1, 0, 0, 0)
}

func TestSessionRetryRunsOnlyWhenEventLoopProcessesIt(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder(), gnbRx: make(chan gnbcontext.UEMessage, 5)}
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	session.EstablishmentFailed()
	for i := 0; i < 5; i++ {
		ue.Lock()
		ue.schedulePduSessionRetryLocked(session, 0)
		ue.Unlock()
		retry := awaitSessionRetry(t, ue.PduSessionRetries())
		// Timer expiration itself does not begin an attempt or consume a retry.
		if session.T3580Retries != i {
			t.Fatal("timer consumed retry before UE processing")
		}
		assertSessionResults(t, ue.Results, uint64(i+1), 0, 0, 0)
		encoded := 0
		encode := func() ([]byte, error) { encoded++; return nil, nil }
		if err := ue.StartPduSessionRetry(retry, encode); err != nil {
			t.Fatal(err)
		}
		if err := ue.StartPduSessionRetry(retry, encode); err != nil {
			t.Fatal(err)
		}
		if encoded != 1 {
			t.Fatal("queued token encoded more than one request")
		}
		assertSessionResults(t, ue.Results, uint64(i+2), 0, 0, 1)
		session.EstablishmentFailed()
	}
	if ue.SchedulePduSessionRetry(session) {
		t.Fatal("session exceeded the five-retry limit")
	}
}

func TestSessionAcceptInvalidatesQueuedRetry(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder()}
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	ue.Lock()
	ue.schedulePduSessionRetryLocked(session, 0)
	ue.Unlock()
	retry := awaitSessionRetry(t, ue.PduSessionRetries())
	session.SetStateSM_PDU_SESSION_ACTIVE()
	encoded := false
	if err := ue.StartPduSessionRetry(retry, func() ([]byte, error) {
		encoded = true
		return nil, nil
	}); err != nil {
		t.Fatal(err)
	}
	if encoded {
		t.Fatal("accepted session retried establishment")
	}
	assertSessionResults(t, ue.Results, 1, 1, 0, 0)
}

func TestTerminationCancelsRetryStartingAtSameTime(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder(), scenarioChan: make(chan scenario.ScenarioMessage), gnbRx: make(chan gnbcontext.UEMessage, 1)}
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	session.EstablishmentFailed()
	ue.Lock()
	ue.schedulePduSessionRetryLocked(session, 0)
	ue.Unlock()
	retry := awaitSessionRetry(t, ue.PduSessionRetries())
	encoding, resume, started, terminated := make(chan struct{}), make(chan struct{}), make(chan error, 1), make(chan struct{})
	go func() {
		started <- ue.StartPduSessionRetry(retry, func() ([]byte, error) {
			close(encoding)
			<-resume
			return nil, nil
		})
	}()
	select {
	case <-encoding:
	case <-time.After(time.Second):
		t.Fatal("retry did not start encoding")
	}
	go func() {
		ue.Terminate()
		close(terminated)
	}()
	close(resume)
	select {
	case err := <-started:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("retry did not finish")
	}
	select {
	case <-terminated:
	case <-time.After(time.Second):
		t.Fatal("UE did not terminate")
	}
	// Starting and cancellation are atomic: termination cancels the new attempt
	// rather than allowing a late Begin to remain pending after CancelUE.
	assertSessionResults(t, ue.Results, 2, 0, 1, 0)
}

func TestSessionRetryWithClosedConnectionDoesNotStartAttempt(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder(), gnbRx: make(chan gnbcontext.UEMessage, 1)}
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	session.EstablishmentFailed()
	ue.Lock()
	ue.schedulePduSessionRetryLocked(session, 0)
	ue.Unlock()
	retry := awaitSessionRetry(t, ue.PduSessionRetries())
	// Idle/context release can close the gNB channel while leaving the UE alive.
	ue.SetGnbRx(nil)
	encoded := 0
	encode := func() ([]byte, error) {
		encoded++
		return nil, nil
	}
	if err := ue.StartPduSessionRetry(retry, encode); err != nil {
		t.Fatal(err)
	}
	if err := ue.StartPduSessionRequest(session, encode); err != nil {
		t.Fatal(err)
	}
	if encoded != 0 || session.T3580Retries != 0 {
		t.Fatalf("disconnected request encoded=%d retries=%d", encoded, session.T3580Retries)
	}
	assertSessionResults(t, ue.Results, 1, 0, 0, 0)
	// The same UE and session can request again once the connection resumes.
	ue.SetGnbRx(make(chan gnbcontext.UEMessage, 1))
	if err := ue.StartPduSessionRequest(session, encode); err != nil {
		t.Fatal(err)
	}
	if encoded != 1 {
		t.Fatal("resumed connection did not encode its request")
	}
	assertSessionResults(t, ue.Results, 2, 0, 0, 1)
}

func awaitSessionRetry(t *testing.T, queue <-chan PduSessionRetry) PduSessionRetry {
	t.Helper()
	select {
	case retry := <-queue:
		return retry
	case <-time.After(time.Second):
		t.Fatal("retry timer did not queue work")
		return PduSessionRetry{}
	}
}
