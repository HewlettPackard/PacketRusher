/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package context

import (
	"testing"
	"time"

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
			ue := &UEContext{scenarioChan: make(chan scenario.ScenarioMessage)}
			rx := make(chan gnbcontext.UEMessage, 1)
			rx <- gnbcontext.UEMessage{} // No receiver, forcing a blocking send.
			ue.SetGnbRx(rx)
			lost := make(chan struct{})
			ue.SetGnbConnectionLost(lost)
			session, token := sessionRequestFixture(t, ue, retrying)
			completed, encoded := make(chan error, 1), make(chan struct{})
			go func() {
				encode := func() ([]byte, error) { close(encoded); return []byte{1}, nil }
				if retrying {
					completed <- ue.StartPduSessionRetry(token, encode)
				} else {
					completed <- ue.StartPduSessionRequest(session, encode)
				}
			}()
			<-encoded
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
			terminated := make(chan struct{})
			go func() { ue.Terminate(); close(terminated) }()
			select {
			case <-terminated:
			case <-time.After(time.Second):
				t.Fatal("UE termination could not acquire the session send's lock")
			}
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
			ue := &UEContext{scenarioChan: make(chan scenario.ScenarioMessage)}
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
			if retrying {
				// Consuming a rejected queued token must invalidate it permanently.
				ue.SetGnbConnectionLost(nil)
				if err := ue.StartPduSessionRetry(token, encode); err != nil {
					t.Fatal(err)
				}
				if encoded != 0 || len(rx) != 0 {
					t.Fatal("cancelled token started after the connection recovered")
				}
			}
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
