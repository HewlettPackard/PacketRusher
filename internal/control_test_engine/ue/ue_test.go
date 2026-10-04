/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 * © Copyright 2026 Valentin D'Emmanuele
 */
package ue

import (
	"github.com/free5gc/nas/ie"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"
)

func startUELoop(t *testing.T, ue *context.UEContext) (chan procedures.UeTesterMessage, <-chan struct{}) {
	t.Helper()
	manager := make(chan procedures.UeTesterMessage)
	done := make(chan struct{})
	go func() {
		handleUE(ue, manager)
		close(done)
	}()
	t.Cleanup(func() {
		close(manager)
		waitUELoop(t, done)
	})
	return manager, done
}

func waitUELoop(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("UE loop did not stop")
	}
}

func TestAssociationLossStopsRegisteredOrIdleUE(t *testing.T) {
	for _, mmState := range []int{context.MM5G_REGISTERED, context.MM5G_IDLE} {
		t.Run(map[int]string{context.MM5G_REGISTERED: "registered", context.MM5G_IDLE: "idle"}[mmState], func(t *testing.T) {
			ue := &context.UEContext{StateMM: mmState}
			lost := make(chan struct{})
			ue.SetGnbConnectionLost(lost)
			// No TX channel is needed: a normal context release may already have
			// disabled it. Fault cancellation remains independently selectable.
			_, done := startUELoop(t, ue)
			close(lost)
			waitUELoop(t, done)
		})
	}
}

func TestNormalIdleContextReleaseKeepsUEAvailableToScenario(t *testing.T) {
	ue := &context.UEContext{StateMM: context.MM5G_IDLE}
	tx := make(chan gnbContext.UEMessage)
	rx := make(chan gnbContext.UEMessage)
	ue.SetGnbTx(tx)
	ue.SetGnbRx(rx)
	ue.SetGnbConnectionLost(make(chan struct{}))
	manager, done := startUELoop(t, ue)
	close(tx)
	select {
	case _, open := <-rx:
		if open {
			t.Fatal("old uplink was not closed after context release")
		}
	case <-time.After(time.Second):
		t.Fatal("context release did not stop the old gNB listener")
	}
	select {
	case <-done:
		t.Fatal("normal Idle context release terminated the UE")
	case manager <- procedures.UeTesterMessage{Type: procedures.Kill}:
	case <-time.After(time.Second):
		t.Fatal("UE stopped receiving scenario actions after context release")
	}
	waitUELoop(t, done)
	if ue.GetGnbTx() != nil || ue.GetGnbRx() != nil {
		t.Fatal("released gNB channels were retained")
	}
}

func TestHandoverReplacesAssociationLossSignal(t *testing.T) {
	ue := &context.UEContext{StateMM: context.MM5G_REGISTERED}
	oldLost := make(chan struct{})
	newLost := make(chan struct{})
	ue.SetGnbConnectionLost(oldLost)
	ue.SetGnbRx(make(chan gnbContext.UEMessage, 1))
	gnbMsgHandler(gnbContext.UEMessage{
		GNBRx: make(chan gnbContext.UEMessage), GNBTx: make(chan gnbContext.UEMessage),
		GNBInboundChannel: make(chan gnbContext.UEMessage), ConnectionLost: newLost,
	}, ue)
	close(oldLost)
	manager, done := startUELoop(t, ue)
	// A scenario action rendezvous proves the old association cannot cancel
	// the target connection. ServiceRequest is inert for a registered UE.
	select {
	case <-done:
		t.Fatal("source association loss terminated the target UE")
	case manager <- procedures.UeTesterMessage{Type: procedures.ServiceRequest}:
	case <-time.After(time.Second):
		t.Fatal("UE did not receive scenario action after handover")
	}
	close(newLost)
	waitUELoop(t, done)
}

// Work handed to the UE with RunOnUE is run by the UE's own goroutine, between the
// messages it handles.
func TestHandleUERunsHandedOverWork(t *testing.T) {
	capability := &ie.UESecCapability{Length: 2, EA05G: true, IA05G: true}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)

	mgr := make(chan procedures.UeTesterMessage)
	stopped := make(chan struct{})
	go func() { handleUE(ue, mgr); close(stopped) }()

	ran := make(chan struct{})
	handed := make(chan bool, 1)
	go func() { handed <- ue.RunOnUE(func() { close(ran) }) }()
	select {
	case <-ran:
	case <-time.After(2 * time.Second):
		t.Fatal("the UE's goroutine should run work handed to it")
	}
	if !<-handed {
		t.Fatal("RunOnUE should report the work as handed over")
	}

	close(mgr)
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("the UE's goroutine should stop when the scenario closes its channel")
	}
}
