/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package sender

import (
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"testing"
	"time"
)

// A UE sending to a gNB that no longer reads, as its association with the AMF failed,
// must not stay blocked: it could then never terminate.
func TestAssociationLossReleasesBlockedSend(t *testing.T) {
	ue := &context.UEContext{}
	ue.SetGnbRx(make(chan gnbContext.UEMessage))
	lost := make(chan struct{})
	ue.SetGnbConnectionLost(lost)

	sent := make(chan struct{})
	go func() { SendToGnb(ue, []byte{0x7e}); close(sent) }()
	close(lost)

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("association loss left the send blocked")
	}
}
