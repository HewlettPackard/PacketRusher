// SPDX-License-Identifier: Apache-2.0

package ngap

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/ngap/ngapType"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
)

func TestOrderedDispatcherPreservesUEOrderAndConcurrency(t *testing.T) {
	dispatcher := newOrderedDispatcher(128)
	blocked, started, other := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	var received []int
	dispatcher.enqueue(1, func() {
		close(started)
		<-blocked
		mu.Lock()
		received = append(received, 0)
		mu.Unlock()
	})
	<-started
	for i := 1; i <= 50; i++ {
		value := i
		dispatcher.enqueue(1, func() { mu.Lock(); received = append(received, value); mu.Unlock() })
	}
	dispatcher.enqueue(2, func() { close(other) })
	select {
	case <-other:
	case <-time.After(time.Second):
		t.Fatal("another UE was blocked by the first UE")
	}
	close(blocked)
	dispatcher.wait()
	want := make([]int, 51)
	for i := range want {
		want[i] = i
	}
	if !reflect.DeepEqual(received, want) {
		t.Fatalf("UE order: got %v", received)
	}
	if len(dispatcher.queues) != 0 {
		t.Fatal("idle UE workers were retained")
	}
}

func TestAMFOnlyReleaseSharesNASQueueBeforeContextPublication(t *testing.T) {
	nasPDU := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{Value: ngapType.InitiatingMessageValue{
			DownlinkNASTransport: &ngapType.DownlinkNASTransport{ProtocolIEs: ngapType.ProtocolIEContainerDownlinkNASTransportIEs{
				List: []ngapType.DownlinkNASTransportIEs{{Value: ngapType.DownlinkNASTransportIEsValue{
					RANUENGAPID: &ngapType.RANUENGAPID{Value: 12}, AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 34},
				}}},
			}},
		}},
	}
	release := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{Value: ngapType.InitiatingMessageValue{
			UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
				List: []ngapType.UEContextReleaseCommandIEs{{Value: ngapType.UEContextReleaseCommandIEsValue{
					UENGAPIDs: &ngapType.UENGAPIDs{AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 34}},
				}}},
			}},
		}},
	}
	if messageUEKey(nasPDU) != messageUEKey(release) {
		t.Fatal("AMF-only release does not use the earlier NAS message's queue")
	}
}

func TestZeroAMFIDRoutesToSameQueue(t *testing.T) {
	message := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{Value: ngapType.InitiatingMessageValue{
			DownlinkNASTransport: &ngapType.DownlinkNASTransport{ProtocolIEs: ngapType.ProtocolIEContainerDownlinkNASTransportIEs{
				List: []ngapType.DownlinkNASTransportIEs{{Value: ngapType.DownlinkNASTransportIEsValue{
					RANUENGAPID: &ngapType.RANUENGAPID{Value: 12}, AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 0},
				}}},
			}},
		}},
	}
	if got := messageUEKey(message); got != -1 {
		t.Fatalf("AMF ID zero must have its own UE queue, got %d", got)
	}
}

func TestFullQueueCanBeCancelledDuringAssociationShutdown(t *testing.T) {
	dispatcher := newOrderedDispatcher(1)
	running, release := make(chan struct{}), make(chan struct{})
	dispatcher.enqueue(1, func() { close(running); <-release })
	<-running
	enqueued := make(chan bool)
	go func() { enqueued <- dispatcher.enqueue(1, func() { t.Error("work accepted after shutdown") }) }()
	dispatcher.stop()
	select {
	case accepted := <-enqueued:
		if accepted {
			t.Fatal("full queue accepted work after shutdown")
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not unblock the association reader")
	}
	close(release)
	dispatcher.wait()
}

func TestSaturatedQueueHonorsProcessingDeadline(t *testing.T) {
	dispatcher := newOrderedDispatcher(1)
	running, release := make(chan struct{}), make(chan struct{})
	dispatcher.enqueue(1, func() { close(running); <-release })
	<-running
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if dispatcher.enqueueContext(ctx, 1, func() { t.Error("work accepted after deadline") }) {
		t.Fatal("cancelled enqueue succeeded")
	}
	close(release)
	dispatcher.wait()
}

func TestRANOnlyProcedureFollowsUnpublishedAMFID(t *testing.T) {
	dispatcher := newOrderedDispatcher(8)
	gnb := &gnbcontext.GNBContext{}
	nas := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage, InitiatingMessage: &ngapType.InitiatingMessage{Value: ngapType.InitiatingMessageValue{
		DownlinkNASTransport: &ngapType.DownlinkNASTransport{ProtocolIEs: ngapType.ProtocolIEContainerDownlinkNASTransportIEs{List: []ngapType.DownlinkNASTransportIEs{{Value: ngapType.DownlinkNASTransportIEsValue{RANUENGAPID: &ngapType.RANUENGAPID{Value: 12}, AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 34}}}}}},
	}}}
	indication := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage, InitiatingMessage: &ngapType.InitiatingMessage{Value: ngapType.InitiatingMessageValue{
		ErrorIndication: &ngapType.ErrorIndication{ProtocolIEs: ngapType.ProtocolIEContainerErrorIndicationIEs{List: []ngapType.ErrorIndicationIEs{{Value: ngapType.ErrorIndicationIEsValue{RANUENGAPID: &ngapType.RANUENGAPID{Value: 12}}}}}},
	}}}
	key := dispatcher.messageKey(gnb, nas)
	started, release := make(chan struct{}), make(chan struct{})
	dispatcher.enqueue(key, func() { close(started); <-release })
	<-started
	if dispatcher.messageKey(gnb, indication) != key {
		t.Fatal("RAN-only indication can overtake unpublished AMF identity")
	}
	close(release)
	dispatcher.wait()
	if len(dispatcher.ueAliases) != 0 || len(dispatcher.aliasUEs) != 0 {
		t.Fatal("idle worker retained UE aliases")
	}
}

func orderedNASMessage(ue *gnbcontext.GNBUe, payload byte) *ngapType.NGAPPDU {
	return &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeDownlinkNASTransport},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentDownlinkNASTransport,
				DownlinkNASTransport: &ngapType.DownlinkNASTransport{ProtocolIEs: ngapType.ProtocolIEContainerDownlinkNASTransportIEs{List: []ngapType.DownlinkNASTransportIEs{
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID}, Value: ngapType.DownlinkNASTransportIEsValue{AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 42}}},
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID}, Value: ngapType.DownlinkNASTransportIEsValue{RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()}}},
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDNASPDU}, Value: ngapType.DownlinkNASTransportIEsValue{NASPDU: &ngapType.NASPDU{Value: []byte{payload}}}},
				}}},
			},
		},
	}
}

func TestOrderedProductionNASDeliveryPrecedesAMFOnlyRelease(t *testing.T) {
	gnb := createTestGNBContext()
	tx := make(chan gnbcontext.UEMessage)
	ue, err := gnb.NewGnBUe(tx, make(chan gnbcontext.UEMessage, 1), 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newOrderedDispatcher(8)
	for _, payload := range []byte{1, 2} {
		message := orderedNASMessage(ue, payload)
		dispatcher.enqueue(dispatcher.messageKey(gnb, message), func() { dispatchUEMessage(nil, gnb, message) })
	}
	release := &ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeUEContextRelease},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
				UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{List: []ngapType.UEContextReleaseCommandIEs{{
					Id:    ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDUENGAPIDs},
					Value: ngapType.UEContextReleaseCommandIEsValue{UENGAPIDs: &ngapType.UENGAPIDs{Present: ngapType.UENGAPIDsPresentAMFUENGAPID, AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 42}}},
				}}}},
			},
		},
	}
	dispatcher.enqueue(dispatcher.messageKey(gnb, release), func() { dispatchUEMessage(nil, gnb, release) })
	for _, want := range []byte{1, 2} {
		select {
		case message, open := <-tx:
			if !open || len(message.Nas) != 1 || message.Nas[0] != want {
				t.Fatalf("release overtook NAS %d: open=%v message=%+v", want, open, message)
			}
		case <-time.After(time.Second):
			t.Fatal("ordered NAS delivery stalled")
		}
	}
	select {
	case _, open := <-tx:
		if open {
			t.Fatal("release did not close UE delivery")
		}
	case <-time.After(time.Second):
		t.Fatal("release did not complete after ordered delivery")
	}
	dispatcher.wait()
}
