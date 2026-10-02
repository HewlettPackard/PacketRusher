// SPDX-License-Identifier: Apache-2.0

package context

import (
	"testing"
	"time"
)

func TestDownlinkBatchDoesNotFollowReplacedConnection(t *testing.T) {
	ue := &GNBUe{}
	oldTX := make(chan UEMessage)
	newTX := make(chan UEMessage, 1)
	ue.SetGnbTx(oldTX)
	ue.ProcessDownlink(func() {
		ue.DeliverToUE(UEMessage{Nas: []byte("old connection")})
		ue.CloseUEChannel()
		ue.SetGnbTx(newTX)
	})
	if _, open := <-oldTX; open {
		t.Fatal("old connection remained open")
	}
	if len(newTX) != 0 {
		t.Fatal("a deferred message crossed into the replacement connection")
	}
	ue.CloseUEChannel()
}

func TestAssociationLossCancelsHandlerBatchBeforeFlush(t *testing.T) {
	ue := &GNBUe{}
	tx := make(chan UEMessage)
	ue.SetGnbTx(tx)
	staged, finish := make(chan struct{}), make(chan struct{})
	done := make(chan struct{})
	go func() {
		ue.ProcessDownlink(func() {
			ue.DeliverToUE(UEMessage{Nas: []byte("pending")})
			close(staged)
			<-finish
		})
		close(done)
	}()
	<-staged
	ue.FailUEChannel()
	close(finish)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("failed association retained a staged downlink delivery")
	}
	if _, open := <-tx; open {
		t.Fatal("failed association left the delivery channel open")
	}
}

func TestReplacingDeliveryRetiresBlockedGeneration(t *testing.T) {
	for _, disconnect := range []bool{false, true} {
		name := "replacement"
		if disconnect {
			name = "disconnect"
		}
		t.Run(name, func(t *testing.T) {
			ue := &GNBUe{}
			oldTX := make(chan UEMessage, 1)
			oldTX <- UEMessage{Nas: []byte{99}}
			ue.SetGnbTx(oldTX)
			previous := ue.delivery
			staged, done := make(chan struct{}), make(chan struct{})
			go func() {
				ue.ProcessDownlink(func() {
					ue.DeliverToUE(UEMessage{Nas: []byte{1}})
					close(staged)
				})
				close(done)
			}()
			<-staged
			var replacement chan UEMessage
			if !disconnect {
				replacement = make(chan UEMessage, 1)
			}
			ue.SetGnbTx(replacement)
			select {
			case <-previous.done:
			default:
				t.Fatal("replacement did not cancel previous delivery")
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("replacement left previous delivery blocked")
			}
			<-oldTX // Only the message queued before replacement remains.
			if _, open := <-oldTX; open {
				t.Fatal("retired connection remains open")
			}
			if replacement != nil {
				if len(replacement) != 0 {
					t.Fatal("old batch crossed into replacement")
				}
				if !ue.DeliverToUE(UEMessage{Nas: []byte{2}}) {
					t.Fatal("replacement could not receive")
				}
				if message := <-replacement; message.Nas[0] != 2 {
					t.Fatal("unexpected replacement message")
				}
			}
			ue.FailUEChannel()
		})
	}
}

func TestAssigningCurrentDeliveryPreservesItsGeneration(t *testing.T) {
	ue := &GNBUe{}
	tx := make(chan UEMessage, 1)
	ue.SetGnbTx(tx)
	current := ue.delivery
	ue.ProcessDownlink(func() {
		ue.DeliverToUE(UEMessage{Nas: []byte{1}})
		ue.SetGnbTx(tx)
	})
	if ue.delivery != current {
		t.Fatal("same channel assignment replaced its generation")
	}
	select {
	case <-current.done:
		t.Fatal("same channel assignment retired a live connection")
	default:
	}
	if message := <-tx; message.Nas[0] != 1 {
		t.Fatal("same-channel batch was lost")
	}
	ue.CloseUEChannel()
}
