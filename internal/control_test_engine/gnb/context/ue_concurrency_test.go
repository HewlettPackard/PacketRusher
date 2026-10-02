/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"sync"
	"testing"
)

// Each NGAP message is dispatched on its own goroutine, and the handlers set the UE's
// AMF UE NGAP ID while other goroutines read it to build messages. The field was a
// plain int64, which the race detector reported intermittently in ./test. Run with
// -race, this fails on a plain field.
func TestGnbUeAmfUeIdIsSafeForConcurrentUse(t *testing.T) {
	ue := &GNBUe{}

	var wg sync.WaitGroup
	for i := int64(0); i < 8; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); ue.SetAmfUeId(i) }()
		go func() { defer wg.Done(); _ = ue.GetAmfUeId() }()
	}
	wg.Wait()
}

// The UE's state is set by the NGAP handlers (Ready on a setup response, Down on a
// release) while the gNB's other goroutines read it. Run with -race, this fails on a
// plain field.
func TestGnbUeStateIsSafeForConcurrentUse(t *testing.T) {
	ue := &GNBUe{}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); ue.SetStateReady() }()
		go func() { defer wg.Done(); _ = ue.GetState() }()
	}
	wg.Wait()
}
