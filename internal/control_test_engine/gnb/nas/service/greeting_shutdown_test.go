// SPDX-License-Identifier: Apache-2.0
package service

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"testing"
	"time"
)

func TestPublishedGreetingFailureRetiresUEAndKeepsListener(t *testing.T) {
	node := &context.GNBContext{}
	node.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"))
	node.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:3")).SetStateActive()
	tx, rx := make(chan context.UEMessage, 1), make(chan context.UEMessage, 1)
	tx <- context.UEMessage{} // The greeting cannot complete before failure.
	lost := make(chan struct{})
	listenerDone := make(chan struct{})
	go func() { gnbListen(node); close(listenerDone) }()
	t.Cleanup(func() {
		node.Terminate()
		close(rx)
		select {
		case <-listenerDone:
		case <-time.After(time.Second):
			t.Error("gNB listener did not join")
		}
	})
	if err := node.QueueUE(context.UEMessage{GNBTx: tx, GNBRx: rx, ConnectionLost: lost, PrUeId: 1}); err != nil {
		t.Fatal(err)
	}
	var published *context.GNBUe
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	poll := time.NewTicker(time.Millisecond)
	defer poll.Stop()
	for published == nil {
		published, _ = node.GetGnbUeByPrUeId(1)
		if published != nil {
			break
		}
		select {
		case <-poll.C:
		case <-timer.C:
			t.Fatal("listener did not publish the new UE")
		}
	}
	published.FailUEChannel() // Also owns closure of the admitted loss signal.
	select {
	case <-lost:
	default:
		t.Fatal("failed published connection did not notify loss")
	}
	// A paging request proves the production listener got past the failed greeting
	// without starting/retaining a connection worker for the rejected publication.
	reply := make(chan context.UEMessage, 1)
	if err := node.QueueUE(context.UEMessage{FetchPagedUEs: true, GNBTx: reply}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reply:
	case <-time.After(time.Second):
		t.Fatal("failed greeting blocked the listener")
	}
	if _, err := node.GetGnbUeByPrUeId(1); err == nil {
		t.Fatal("failed greeting retained its PacketRusher UE entry")
	}
	if _, err := node.GetGnbUe(published.GetRanUeId()); err == nil {
		t.Fatal("failed greeting retained its RAN UE entry")
	}
	for message := range tx {
		if message.Mcc != "" || message.Mnc != "" {
			t.Fatal("failed greeting was delivered after closure")
		}
	}
}
