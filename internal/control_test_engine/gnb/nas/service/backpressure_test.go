// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/netip"
	"runtime"
	"testing"
	"time"

	free5gcngap "github.com/free5gc/ngap"
	"github.com/free5gc/ngap/ngapType"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"
	uecontext "my5G-RANTester/internal/control_test_engine/ue/context"
	uesender "my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"
)

func backpressureUE(t *testing.T, capacity int) (*gnbcontext.GNBContext, *gnbcontext.GNBUe, *uecontext.UEContext, []byte) {
	t.Helper()
	gnb := &gnbcontext.GNBContext{}
	gnb.NewRanGnbContext("test", "001", "01", "000001", "1", "000001", netip.MustParseAddrPort("127.0.0.1:9999"), netip.MustParseAddrPort("127.0.0.1:2152"))
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	amf.SetStateActive()
	rx, tx := make(chan gnbcontext.UEMessage, capacity), make(chan gnbcontext.UEMessage, capacity)
	lost := make(chan struct{})
	gnbUE, err := gnb.NewGnBUe(tx, rx, 1, nil, lost)
	if err != nil {
		t.Fatal(err)
	}
	gnbUE.SetStateDown() // Exercise channel dispatch without opening SCTP sockets.
	ue := &uecontext.UEContext{}
	ue.SetGnbRx(rx)
	ue.SetGnbConnectionLost(lost)
	for range capacity {
		tx <- gnbcontext.UEMessage{IsNas: true, Nas: []byte{0}}
		rx <- gnbcontext.UEMessage{IsNas: true}
	}
	pdu := ngapType.NGAPPDU{Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeDownlinkNASTransport},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentDownlinkNASTransport,
				DownlinkNASTransport: &ngapType.DownlinkNASTransport{ProtocolIEs: ngapType.ProtocolIEContainerDownlinkNASTransportIEs{List: []ngapType.DownlinkNASTransportIEs{
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID}, Value: ngapType.DownlinkNASTransportIEsValue{Present: ngapType.DownlinkNASTransportIEsPresentAMFUENGAPID, AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 42}}},
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID}, Value: ngapType.DownlinkNASTransportIEsValue{Present: ngapType.DownlinkNASTransportIEsPresentRANUENGAPID, RANUENGAPID: &ngapType.RANUENGAPID{Value: gnbUE.GetRanUeId()}}},
					{Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDNASPDU}, Value: ngapType.DownlinkNASTransportIEsValue{Present: ngapType.DownlinkNASTransportIEsPresentNASPDU, NASPDU: &ngapType.NASPDU{Value: []byte{1}}}},
				}}},
			},
		},
	}
	encoded, err := free5gcngap.Encoder(pdu)
	if err != nil {
		t.Fatal(err)
	}
	return gnb, gnbUE, ue, encoded
}

func awaitBackpressureDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("backpressured UE did not make progress")
	}
}

func startBackpressuredDownlink(t *testing.T, gnb *gnbcontext.GNBContext, ue *gnbcontext.GNBUe, encoded []byte) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() { ngap.Dispatch(nil, gnb, encoded); close(done) }()
	deadline := time.Now().Add(time.Second)
	for !ue.HasAmfUeId() && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if !ue.HasAmfUeId() {
		t.Fatal("downlink handler did not publish AMF UE ID")
	}
	return done
}

func TestBidirectionalBackpressureAllowsUplinkProgress(t *testing.T) {
	for _, capacity := range []int{10, 1} {
		t.Run(map[int]string{10: "attach", 1: "handover"}[capacity], func(t *testing.T) {
			gnb, gnbUE, ue, encoded := backpressureUE(t, capacity)
			downlinkDone := startBackpressuredDownlink(t, gnb, gnbUE, encoded)
			ueDone := make(chan struct{})
			go func() {
				uesender.SendToGnb(ue, nil)
				<-gnbUE.GetGnbTx() // A UE can receive downlink again after sending its response.
				close(ueDone)
			}()
			processorDone := make(chan struct{})
			go func() { processingConn(gnbUE, gnb); close(processorDone) }()
			awaitBackpressureDone(t, ueDone)
			awaitBackpressureDone(t, downlinkDone)
			for index := 0; index < capacity; index++ {
				message := <-gnbUE.GetGnbTx()
				want := byte(0)
				if index == capacity-1 {
					want = 1
				}
				if len(message.Nas) != 1 || message.Nas[0] != want {
					t.Fatalf("downlink order changed: got %v, want %d", message.Nas, want)
				}
			}
			close(gnbUE.GetGnbRx())
			awaitBackpressureDone(t, processorDone)
			gnb.DeleteGnBUe(gnbUE)
		})
	}
}

func TestAssociationLossCancelsBothBackpressuredDirections(t *testing.T) {
	gnb, gnbUE, ue, encoded := backpressureUE(t, 1)
	downlinkDone := startBackpressuredDownlink(t, gnb, gnbUE, encoded)
	uplinkDone := make(chan struct{})
	go func() { uesender.SendToGnb(ue, nil); close(uplinkDone) }()
	gnbUE.FailUEChannel()
	awaitBackpressureDone(t, downlinkDone)
	awaitBackpressureDone(t, uplinkDone)
	close(gnbUE.GetGnbRx())
	gnb.DeleteGnBUe(gnbUE)
}
