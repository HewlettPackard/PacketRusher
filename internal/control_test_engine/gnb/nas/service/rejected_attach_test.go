// SPDX-License-Identifier: Apache-2.0
package service

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"testing"
	"time"
)

func TestGNBListenRejectsMissingTXWithoutActiveAMF(t *testing.T) {
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"), netip.MustParseAddrPort("127.0.0.1:2152"))
	inbound := gnb.GetInboundChannel()
	done := make(chan struct{})
	var panicValue any
	go func() {
		defer func() { panicValue = recover(); close(done) }()
		gnbListen(gnb)
	}()
	t.Cleanup(func() {
		close(inbound)
		select {
		case <-done:
			if panicValue != nil {
				t.Errorf("listener panicked: %v", panicValue)
			}
		case <-time.After(time.Second):
			t.Error("listener did not stop after inbound close")
		}
	})

	// No association is active and this rejected request has no TX channel.
	// The listener must safely reject it and remain available for later work.
	inbound <- context.UEMessage{PrUeId: 1}
	reply := make(chan context.UEMessage, 1)
	inbound <- context.UEMessage{FetchPagedUEs: true, GNBTx: reply}
	select {
	case <-reply:
	case <-done:
		t.Fatalf("rejected attach stopped listener: %v", panicValue)
	case <-time.After(time.Second):
		t.Fatal("listener did not process work after rejected attach")
	}
}
