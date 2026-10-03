// SPDX-License-Identifier: Apache-2.0
package context

import (
	"net/netip"
	"sync"
	"testing"
	"time"
)

func TestUEAdmissionSerializesWithTermination(t *testing.T) {
	node := &GNBContext{}
	node.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"))
	const requests = 64
	start := make(chan struct{})
	results := make(chan error, requests)
	var workers sync.WaitGroup
	for range requests {
		workers.Add(1)
		go func() { defer workers.Done(); <-start; results <- node.QueueUE(UEMessage{}) }()
	}
	workers.Add(1)
	go func() { defer workers.Done(); <-start; node.Terminate() }()
	close(start)
	joined := make(chan struct{})
	go func() { workers.Wait(); close(joined) }()
	select {
	case <-joined:
	case <-time.After(time.Second):
		t.Fatal("admission blocked termination")
	}
	close(results)
	accepted := 0
	for err := range results {
		if err == nil {
			accepted++
		}
	}
	delivered := 0
	for range node.GetInboundChannel() {
		delivered++
	}
	if delivered != accepted {
		t.Fatalf("delivered %d requests; accepted %d", delivered, accepted)
	}
	if err := node.QueueUE(UEMessage{}); err == nil {
		t.Fatal("late admission succeeded")
	}
	if ue, err := node.NewGnBUe(make(chan UEMessage, 1), make(chan UEMessage, 1), 1, nil); ue != nil || err == nil {
		t.Fatal("late gNB UE publication succeeded")
	}
	select {
	case <-node.Done():
	default:
		t.Fatal("termination did not broadcast shutdown")
	}
}
