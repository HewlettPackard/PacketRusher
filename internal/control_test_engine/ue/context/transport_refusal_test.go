// SPDX-License-Identifier: Apache-2.0
package context

import (
	"testing"

	"my5G-RANTester/internal/analytics"
)

func TestStaleSessionTransportRefusalCannotFailReplacement(t *testing.T) {
	ue := &UEContext{Results: analytics.NewRecorder()}
	first, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	first.SetStateSM_PDU_SESSION_PENDING()
	if err := ue.DeletePduSession(first.Id); err != nil {
		t.Fatal(err)
	}
	replacement, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	if first.Id != replacement.Id {
		t.Fatal("fixture must reuse the deleted ID")
	}
	replacement.SetStateSM_PDU_SESSION_PENDING()
	first.EstablishmentTransportFailed(1)
	assertSessionResults(t, ue.Results, 2, 0, 1, 1)
	replacement.EstablishmentTransportFailed(2)
	assertSessionResults(t, ue.Results, 2, 0, 1, 1)
	replacement.EstablishmentTransportFailed(1)
	replacement.EstablishmentTransportFailed(1)
	assertSessionResults(t, ue.Results, 2, 0, 1, 0)
	for _, p := range ue.Results.Snapshot().Procedures {
		if p.Procedure == analytics.SessionEstablishment && p.Failure != 1 {
			t.Fatalf("transport refusal failures=%d; want 1", p.Failure)
		}
	}
}
