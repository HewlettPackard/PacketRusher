// SPDX-License-Identifier: Apache-2.0
package context

import (
	"runtime"
	"testing"
	"time"

	"my5G-RANTester/internal/analytics"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
)

func TestSessionRequestAssociationLossUnblocksSend(t *testing.T) {
	for _, retrying := range []bool{false, true} {
		name := "initial"
		if retrying {
			name = "retry"
		}
		t.Run(name, func(t *testing.T) {
			results := analytics.NewRecorder()
			ue := &UEContext{Results: results, scenarioChan: make(chan scenario.ScenarioMessage)}
			rx := make(chan gnbcontext.UEMessage, 1)
			rx <- gnbcontext.UEMessage{} // No receiver, forcing a blocking send.
			ue.SetGnbRx(rx)
			lost := make(chan struct{})
			ue.SetGnbConnectionLost(lost)
			session, token := sessionRequestFixture(t, ue, retrying)
			completed := make(chan error, 1)
			go func() {
				encode := func() ([]byte, error) { return []byte{1}, nil }
				if retrying {
					completed <- ue.StartPduSessionRetry(token, encode)
				} else {
					completed <- ue.StartPduSessionRequest(session, encode)
				}
			}()
			// Wait until the attempt is accounted for, not merely until encoding
			// starts. Closing the association must then settle this pending send.
			deadline := time.Now().Add(time.Second)
			for sessionProcedure(t, results).Pending != 1 {
				if time.Now().After(deadline) {
					t.Fatal("session request did not begin accounting")
				}
				runtime.Gosched()
			}
			close(lost)
			select {
			case err := <-completed:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				// Prevent a broken implementation from leaking its blocked sender.
				<-rx
				<-completed
				t.Fatal("association loss left the PDU session send blocked")
			}
			if len(rx) != 1 {
				t.Fatal("failed association received a session request")
			}
			wantStarted, wantFailure := uint64(1), uint64(0)
			if retrying {
				wantStarted, wantFailure = 2, 1
			}
			assertSessionLossResults(t, results, wantStarted, wantFailure, 1)
			terminated := make(chan struct{})
			go func() { ue.Terminate(); close(terminated) }()
			select {
			case <-terminated:
			case <-time.After(time.Second):
				t.Fatal("UE termination could not acquire the session send's lock")
			}
			results.Close()
			assertSessionLossResults(t, results, wantStarted, wantFailure, 1)
		})
	}
}

func TestSessionRequestAlreadyLostAssociationDoesNotStart(t *testing.T) {
	for _, retrying := range []bool{false, true} {
		name := "initial"
		if retrying {
			name = "retry"
		}
		t.Run(name, func(t *testing.T) {
			results := analytics.NewRecorder()
			ue := &UEContext{Results: results, scenarioChan: make(chan scenario.ScenarioMessage)}
			rx := make(chan gnbcontext.UEMessage, 1)
			ue.SetGnbRx(rx)
			lost := make(chan struct{})
			ue.SetGnbConnectionLost(lost)
			session, token := sessionRequestFixture(t, ue, retrying)
			close(lost)
			encoded := 0
			encode := func() ([]byte, error) { encoded++; return []byte{1}, nil }
			var err error
			if retrying {
				err = ue.StartPduSessionRetry(token, encode)
			} else {
				err = ue.StartPduSessionRequest(session, encode)
			}
			if err != nil {
				t.Fatal(err)
			}
			if encoded != 0 || session.T3580Retries != 0 || len(rx) != 0 {
				t.Fatalf("failed association encoded=%d retries=%d sent=%d", encoded, session.T3580Retries, len(rx))
			}
			wantStarted, wantFailure := uint64(0), uint64(0)
			if retrying {
				wantStarted, wantFailure = 1, 1
				// Consuming a rejected queued token must invalidate it permanently.
				ue.SetGnbConnectionLost(nil)
				if err := ue.StartPduSessionRetry(token, encode); err != nil {
					t.Fatal(err)
				}
				if encoded != 0 || len(rx) != 0 {
					t.Fatal("cancelled token started after the connection recovered")
				}
			}
			assertSessionLossResults(t, results, wantStarted, wantFailure, 0)
			ue.Terminate()
		})
	}
}

func sessionRequestFixture(t *testing.T, ue *UEContext, retrying bool) (*UEPDUSession, PduSessionRetry) {
	t.Helper()
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	if !retrying {
		return session, PduSessionRetry{}
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	session.EstablishmentFailed()
	ue.Lock()
	ue.schedulePduSessionRetryLocked(session, 0)
	ue.Unlock()
	return session, awaitSessionRetry(t, ue.PduSessionRetries())
}

func sessionProcedure(t *testing.T, results *analytics.Recorder) analytics.ProcedureResult {
	t.Helper()
	for _, procedure := range results.Snapshot().Procedures {
		if procedure.Procedure == analytics.SessionEstablishment {
			return procedure
		}
	}
	t.Fatal("session-establishment results are missing")
	return analytics.ProcedureResult{}
}

func assertSessionLossResults(t *testing.T, results *analytics.Recorder, started, failure, cancelled uint64) {
	t.Helper()
	p := sessionProcedure(t, results)
	if p.Started != started || p.Failure != failure || p.Cancelled != cancelled || p.Pending != 0 || p.Success != 0 {
		t.Fatalf("session results=%+v; want started=%d failure=%d cancelled=%d pending=0 success=0", p, started, failure, cancelled)
	}
}
