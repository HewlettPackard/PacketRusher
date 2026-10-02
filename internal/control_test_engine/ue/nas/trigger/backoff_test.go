/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package trigger

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/stretchr/testify/assert"
)

func newTestUE() *context.UEContext {
	capability := nasType.NewUESecurityCapability(nasMessage.RegistrationRequestUESecurityCapabilityType)
	capability.SetLen(2)
	capability.Buffer = []uint8{0x80, 0x80}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	return ue
}

// countSends replaces the NAS send for the length of a test.
func countSends(t *testing.T) *int {
	sends := 0
	prev := sendEstablishmentRequest
	sendEstablishmentRequest = func(*context.UEContext, *context.UEPDUSession) { sends++ }
	t.Cleanup(func() { sendEstablishmentRequest = prev })
	return &sends
}

func sessions(ue *context.UEContext) int {
	n := 0
	for _, s := range ue.PduSession {
		if s != nil {
			n++
		}
	}
	return n
}

func handedOff(t *testing.T, ue *context.UEContext) func() {
	t.Helper()
	select {
	case f := <-ue.Deferred():
		return f
	case <-time.After(2 * time.Second):
		t.Fatal("the held request should be handed to the UE's goroutine when the back-off ends")
		return nil
	}
}

// InitPduSessionRequest goes through the UE's back-off.
func TestInitPduSessionRequestHonoursTheUEBackoff(t *testing.T) {
	sends := countSends(t)
	ue := newTestUE()

	InitPduSessionRequest(ue)
	assert.Equal(t, 1, *sends, "no back-off: sent straight away")

	ue.SetEstablishmentBackoff(0, true)
	InitPduSessionRequest(ue)
	assert.Equal(t, 1, *sends, "deactivated: not sent")
	assert.Equal(t, 1, sessions(ue), "the dropped session's slot is freed")

	// A deactivated back-off stays so; a timer needs a UE of its own.
	held := newTestUE()
	held.SetEstablishmentBackoff(50*time.Millisecond, false)
	InitPduSessionRequest(held)
	assert.Equal(t, 1, *sends, "held while the back-off runs")
	f := handedOff(t, held)
	assert.Equal(t, 1, *sends, "nothing is sent until the UE's goroutine runs it")
	f()
	assert.Equal(t, 2, *sends)
}

// A held request is checked again when its wait ends, so a back-off extended in the
// meantime keeps it held.
func TestInitPduSessionRequestStaysHeldWhenTheBackoffIsExtended(t *testing.T) {
	sends := countSends(t)
	ue := newTestUE()

	ue.SetEstablishmentBackoff(50*time.Millisecond, false)
	InitPduSessionRequest(ue)
	ue.SetEstablishmentBackoff(300*time.Millisecond, false)

	handedOff(t, ue)()
	assert.Zero(t, *sends, "still held: the back-off was extended")
	handedOff(t, ue)()
	assert.Equal(t, 1, *sends)
}
