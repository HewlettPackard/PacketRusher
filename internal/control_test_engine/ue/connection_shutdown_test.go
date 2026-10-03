// SPDX-License-Identifier: Apache-2.0
package ue

import (
	"my5G-RANTester/internal/analytics"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	uecontext "my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/service"
	fixture "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"
)

func shutdownGNB() *gnb.GNBContext {
	node := &gnb.GNBContext{}
	node.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"))
	return node
}

func TestNewUEAfterGNBTerminationJoins(t *testing.T) {
	recorder := analytics.NewRecorder()
	previous := analytics.Current()
	analytics.SetCurrent(recorder)
	t.Cleanup(func() { analytics.SetCurrent(previous); recorder.Close() })
	node := shutdownGNB()
	node.Terminate() // Deterministically closes the production inbound channel first.
	cfg := fixture.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"), nil)
	var wg sync.WaitGroup
	wg.Add(1)
	events := NewUE(cfg, 1, make(chan procedures.UeTesterMessage), node, &wg)
	drained := make(chan struct{})
	go func() {
		for range events {
		}
		close(drained)
	}()
	joined := make(chan struct{})
	go func() { wg.Wait(); close(joined) }()
	for _, done := range []<-chan struct{}{joined, drained} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("late UE constructor did not join cleanup")
		}
	}
	for _, procedure := range recorder.Snapshot().Procedures {
		if procedure.Started != 0 || procedure.Pending != 0 {
			t.Fatalf("rejected constructor created procedure attempts: %+v", procedure)
		}
	}
}

func TestHandoverConnectionRetainsTargetLifecycleOwner(t *testing.T) {
	source, target := shutdownGNB(), shutdownGNB()
	t.Cleanup(source.Terminate)
	t.Cleanup(target.Terminate)
	ue := &uecontext.UEContext{}
	ue.SetGnbContext(source)
	ue.SetGnbRx(make(chan gnb.UEMessage, 1))
	gnbMsgHandler(gnb.UEMessage{GNBRx: make(chan gnb.UEMessage, 1), GNBTx: make(chan gnb.UEMessage, 1), GNBInboundChannel: target.GetInboundChannel(), GNB: target}, ue)
	if ue.GetGnbContext() != target {
		t.Fatal("handover kept the source admission owner")
	}
	source.Terminate()
	done := make(chan struct{})
	go func() { service.InitConn(ue, ue.GetGnbInboundChannel()); close(done) }()
	select {
	case message := <-target.GetInboundChannel():
		message.GNBTx <- gnb.UEMessage{Mcc: "001", Mnc: "01"}
	case <-time.After(time.Second):
		t.Fatal("reconnect did not use the live target")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("target greeting did not finish")
	}
	target.Terminate()
	verifyPaging(ue) // Also exercises the idle polling path after channel closure.
	service.InitConn(ue, ue.GetGnbInboundChannel())
	select {
	case <-ue.GetGnbConnectionLost():
	default:
		t.Fatal("late target reconnect did not reject admission")
	}
}
