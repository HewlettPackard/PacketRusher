// SPDX-License-Identifier: Apache-2.0
package service

import (
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	uecontext "my5G-RANTester/internal/control_test_engine/ue/context"
	"net/netip"
	"testing"
	"time"
)

func admissionGNB() *gnb.GNBContext {
	node := &gnb.GNBContext{}
	node.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"))
	return node
}

func waitAdmission(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("UE admission did not finish")
	}
}

func TestInitConnRejectsTerminatedOrFullGNB(t *testing.T) {
	for _, stopped := range []bool{true, false} {
		t.Run(map[bool]string{true: "terminated", false: "full"}[stopped], func(t *testing.T) {
			node := admissionGNB()
			t.Cleanup(node.Terminate)
			if stopped {
				node.Terminate()
			} else {
				for range cap(node.GetInboundChannel()) {
					if err := node.QueueUE(gnb.UEMessage{}); err != nil {
						t.Fatal(err)
					}
				}
			}
			ue := &uecontext.UEContext{}
			ue.SetGnbContext(node)
			done := make(chan struct{})
			go func() { InitConn(ue, node.GetInboundChannel()); close(done) }()
			waitAdmission(t, done)
			select {
			case <-ue.GetGnbConnectionLost():
			default:
				t.Fatal("rejected connection did not notify its UE")
			}
		})
	}
}

func TestInitConnAcceptedGreetingIsCancelledByGNBShutdown(t *testing.T) {
	node := admissionGNB()
	t.Cleanup(node.Terminate)
	ue := &uecontext.UEContext{}
	ue.SetGnbContext(node)
	done := make(chan struct{})
	go func() { InitConn(ue, node.GetInboundChannel()); close(done) }()
	select {
	case <-node.GetInboundChannel():
	case <-time.After(time.Second):
		t.Fatal("gNB did not receive admission")
	}
	node.Terminate()
	waitAdmission(t, done)
	// Admission transferred loss ownership; only the gNB binding may close it.
	select {
	case <-ue.GetGnbConnectionLost():
		t.Fatal("UE closed an accepted connection's owned loss signal")
	default:
	}
}
