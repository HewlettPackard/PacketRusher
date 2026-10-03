// SPDX-License-Identifier: Apache-2.0
package ue

import (
	"context"
	"errors"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	ueContext "my5G-RANTester/internal/control_test_engine/ue/context"
	ueScenario "my5G-RANTester/internal/control_test_engine/ue/scenario"
	"net/netip"
	"testing"
	"time"
)

func TestControlRunsOnUELoopAndRejectsCancelledOrStaleActions(t *testing.T) {
	ue := &ueContext.UEContext{StateMM: ueContext.MM5G_REGISTERED}
	manager, _ := startUELoop(t, ue)
	for _, tc := range []struct {
		action    string
		stale     bool
		cancelled bool
		want      error
	}{
		{"inspect", false, false, nil}, {"idle", true, false, procedures.ErrGeneration}, {"idle", false, true, context.Canceled}, {"idle", false, false, procedures.ErrNotReady},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		if tc.cancelled {
			cancel()
		}
		r := &procedures.ControlRequest{Context: ctx, Action: tc.action, Generation: 2, Reply: make(chan procedures.ControlResult, 1)}
		if tc.stale {
			r.ExpectedGeneration = 1
		}
		manager <- procedures.UeTesterMessage{Type: procedures.Control, Control: r}
		select {
		case result := <-r.Reply:
			if !errors.Is(result.Err, tc.want) {
				t.Fatalf("%s: %v", tc.action, result.Err)
			}
			if result.Attachment.State != "registered" {
				t.Fatalf("invalid action mutated state: %+v", result.Attachment)
			}
		case <-time.After(time.Second):
			t.Fatal("control request did not return")
		}
		cancel()
	}
}

func TestIdleAndReconnectControlsUseProductionConnectionHandshake(t *testing.T) {
	for _, withGuti := range []bool{false, true} {
		t.Run(map[bool]string{false: "registration fallback", true: "service request"}[withGuti], func(t *testing.T) {
			node := &gnb.GNBContext{}
			node.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:1"), netip.MustParseAddrPort("127.0.0.1:2"))
			amf := node.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:3"))
			amf.SetStateActive()
			ue := &ueContext.UEContext{}
			ue.NewRanUeContext("0000000001", &ie.UESecCapability{Length: 2, EA05G: true, IA05G: true}, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{}, "0000", "internet", 1, "", config.TunnelDisabled, make(chan ueScenario.ScenarioMessage, 16), node.GetInboundChannel(), 1)
			ue.StateMM = ueContext.MM5G_REGISTERED
			rx, tx := make(chan gnb.UEMessage, 10), make(chan gnb.UEMessage, 10)
			ue.SetGnbRx(rx)
			ue.SetGnbTx(tx)
			ue.SetGnbConnectionLost(make(chan struct{}))
			gu, err := node.NewGnBUe(tx, rx, 1, nil)
			if err != nil {
				t.Fatal(err)
			}
			gu.SetStateReady()
			if withGuti {
				guti := &ie.MobileId5GS{}
				if err := guti.FromGUTIStr("0010100000000000001"); err != nil {
					t.Fatal(err)
				}
				ue.Set5gGuti(guti)
			}
			t.Cleanup(func() { ue.StopDRX() })
			manager, _ := startUELoop(t, ue)
			request := func(action string) *procedures.ControlRequest {
				r := &procedures.ControlRequest{Context: context.Background(), Action: action, Gnbs: map[string]*gnb.GNBContext{"000008": node}, Generation: 1, Reply: make(chan procedures.ControlResult, 1)}
				manager <- procedures.UeTesterMessage{Type: procedures.Control, Control: r}
				return r
			}
			idle := request("idle")
			if result := <-idle.Reply; result.Err != nil {
				t.Fatal(result.Err)
			}
			if message := <-rx; !message.Idle {
				t.Fatal("idle control did not trigger context release")
			}
			close(tx)
			select {
			case _, open := <-rx:
				if open {
					t.Fatal("uplink not closed")
				}
			case <-time.After(time.Second):
				t.Fatal("normal Idle release not received")
			}
			// Reconnect must install fresh channels before its NAS request. The local
			// handshake carries PLMN information exactly as the production gNB does.
			var reconnect *procedures.ControlRequest
			sent := make(chan struct{})
			go func() { reconnect = request("reconnect"); close(sent) }()
			var connection gnb.UEMessage
			select {
			case connection = <-node.GetInboundChannel():
			case <-time.After(time.Second):
				t.Fatal("reconnect did not request a gNB connection")
			}
			connection.GNBTx <- gnb.UEMessage{Mcc: "001", Mnc: "01"}
			<-sent
			if result := <-reconnect.Reply; result.Err != nil {
				t.Fatal(result.Err)
			}
			message := <-connection.GNBRx
			if !message.IsNas {
				t.Fatal("reconnect did not send NAS")
			}
			payload := message.Nas
			if nas.GetSecHdrType(payload) != nas.SecHdrTypePlainNas {
				payload = payload[nas.SecHdrLen:]
			}
			decoded, err := nas.Parse(payload, nil)
			if err != nil {
				t.Fatal(err)
			}
			if withGuti {
				if _, ok := decoded.(*nas.SvcReq); !ok {
					t.Fatalf("expected service request, got %T", decoded)
				}
			} else {
				if _, ok := decoded.(*nas.RegReq); !ok {
					t.Fatalf("expected registration fallback, got %T", decoded)
				}
			}
			inspected := <-request("inspect").Reply
			if inspected.Attachment.ConnectionGeneration != 2 {
				t.Fatalf("old connection generation retained: %+v", inspected.Attachment)
			}
		})
	}
}
