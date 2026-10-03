/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"sync"
	"testing"
)

func TestHandoverTargetConcurrentPublication(t *testing.T) {
	ue := new(GNBUe)
	first, second := new(GNBContext), new(GNBContext)
	if got := ue.GetHandoverGnodeB(); got != nil {
		t.Fatalf("initial handover target = %p, want nil", got)
	}

	// The CLI trigger publishes the target while the downlink worker reads it
	// when processing Handover Command or UE Context Release.
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 10000; i++ {
			ue.SetHandoverGnodeB(first)
			ue.SetHandoverGnodeB(second)
			ue.SetHandoverGnodeB(nil)
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 30000; i++ {
			if got := ue.GetHandoverGnodeB(); got != nil && got != first && got != second {
				t.Errorf("unexpected handover target = %p", got)
				return
			}
		}
	}()
	close(start)
	workers.Wait()

	ue.SetHandoverGnodeB(second)
	if got := ue.GetHandoverGnodeB(); got != second {
		t.Fatalf("published target = %p, want %p", got, second)
	}
	ue.SetHandoverGnodeB(nil)
	if got := ue.GetHandoverGnodeB(); got != nil {
		t.Fatalf("cleared target = %p, want nil", got)
	}
}
