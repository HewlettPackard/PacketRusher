/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"sync"
	"testing"
	"time"
)

func TestNormalContextReleaseDoesNotSignalAssociationLoss(t *testing.T) {
	ue := &GNBUe{}
	lost := make(chan struct{})
	tx := make(chan UEMessage)
	ue.SetConnectionLost(lost)
	ue.SetGnbTx(tx)
	ue.CloseUEChannel()
	select {
	case <-lost:
		t.Fatal("normal context release failed the association")
	default:
	}
	if _, open := <-tx; open {
		t.Fatal("context release left TX open")
	}
}

func TestAssociationLossCancelsFullDeliveryExactlyOnce(t *testing.T) {
	ue := &GNBUe{}
	lost := make(chan struct{})
	tx := make(chan UEMessage, 1)
	ue.SetConnectionLost(lost)
	ue.SetGnbTx(tx)
	tx <- UEMessage{}
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		if ue.DeliverToUE(UEMessage{}) {
			t.Error("delivery to failed UE succeeded")
		}
	}()
	var failures sync.WaitGroup
	for range 4 {
		failures.Add(1)
		go func() { defer failures.Done(); ue.FailUEChannel() }()
	}
	failures.Wait()
	select {
	case <-lost:
	default:
		t.Fatal("full TX prevented association-loss notification")
	}
	select {
	case <-senderDone:
	case <-time.After(time.Second):
		t.Fatal("failed association left a blocked delivery")
	}
	for range tx {
	}
}

func TestBindingFaultSignalAfterAssociationFailedNotifiesUE(t *testing.T) {
	ue := &GNBUe{}
	ue.FailUEChannel()
	lost := make(chan struct{})
	ue.SetConnectionLost(lost)
	ue.SetConnectionLost(lost)
	select {
	case <-lost:
	default:
		t.Fatal("late UE attach missed the target association failure")
	}
}
