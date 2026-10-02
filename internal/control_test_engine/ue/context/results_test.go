package context

import (
	"testing"

	"my5G-RANTester/internal/analytics"
)

func TestDeletingPendingSessionCancelsAttemptBeforeIDReuse(t *testing.T) {
	results := analytics.NewRecorder()
	ue := &UEContext{Results: results}
	first, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	first.SetStateSM_PDU_SESSION_PENDING()
	if err := ue.DeletePduSession(first.Id); err != nil {
		t.Fatal(err)
	}
	assertSessionResults(t, results, 1, 0, 1, 0)

	second, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	if second.Id != first.Id {
		t.Fatal("fixture must reuse the deleted session ID")
	}
	second.SetStateSM_PDU_SESSION_PENDING()
	second.SetStateSM_PDU_SESSION_PENDING() // retransmission is still one attempt
	assertSessionResults(t, results, 2, 0, 1, 1)
	second.SetStateSM_PDU_SESSION_ACTIVE()
	second.SetStateSM_PDU_SESSION_ACTIVE() // duplicate accept is still one success
	if err := ue.DeletePduSession(second.Id); err != nil {
		t.Fatal(err)
	}
	assertSessionResults(t, results, 2, 1, 1, 0)
}

func assertSessionResults(t *testing.T, results *analytics.Recorder, started, success, cancelled, pending uint64) {
	t.Helper()
	for _, procedure := range results.Snapshot().Procedures {
		if procedure.Procedure != analytics.SessionEstablishment {
			continue
		}
		if procedure.Started != started || procedure.Success != success || procedure.Cancelled != cancelled || procedure.Pending != pending {
			t.Fatalf("session results=%+v; want started=%d success=%d cancelled=%d pending=%d", procedure, started, success, cancelled, pending)
		}
		if procedure.Started != procedure.Success+procedure.Failure+procedure.Cancelled+procedure.Pending {
			t.Fatal("attempt counts are inconsistent")
		}
		return
	}
	t.Fatal("session-establishment results are missing")
}
