// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/netip"
	"testing"
	"time"

	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func TestGNBListenReturnsWhenInboundChannelCloses(t *testing.T) {
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001", netip.MustParseAddrPort("127.0.0.1:9999"), netip.MustParseAddrPort("127.0.0.1:2152"))
	close(gnb.GetInboundChannel())
	done := make(chan struct{})
	go func() { gnbListen(gnb); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("gNB listener kept processing after its inbound channel closed")
	}
}
