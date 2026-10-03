// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"net/netip"
	"sync"
	"testing"
	"time"

	"my5G-RANTester/internal/common/tools"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	fixture "my5G-RANTester/test/aio5gc/lib/tools"
)

func startShutdownSimulation(t *testing.T, loop bool, loopCount int) (*tools.UESimulation, gnbcontext.UEMessage, *gnbcontext.GNBContext, *sync.WaitGroup) {
	t.Helper()
	cfg := fixture.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:9999"), netip.MustParseAddrPort("127.0.0.1:2152"), nil)
	gnb := &gnbcontext.GNBContext{}
	gnb.NewRanGnbContext("000008", "999", "70", "000001", "1", "000001", cfg.GNodeB.ControlIF.AddrPort, cfg.GNodeB.DataIF.AddrPort)
	wg := &sync.WaitGroup{}
	simulation := tools.SimulateSingleUE(tools.UESimulationConfig{
		UeId: 1, Gnbs: map[string]*gnbcontext.GNBContext{"000008": gnb}, Cfg: cfg,
		RegistrationLoop: loop, LoopCount: loopCount, TimeBeforeReregistration: 60_000,
	}, wg)
	var connection gnbcontext.UEMessage
	select {
	case connection = <-gnb.GetInboundChannel():
	case <-time.After(time.Second):
		t.Fatal("UE did not connect")
	}
	connection.GNBTx <- gnbcontext.UEMessage{Mcc: "999", Mnc: "70"}
	select {
	case <-connection.GNBRx:
	case <-time.After(time.Second):
		t.Fatal("UE did not send registration")
	}
	t.Cleanup(func() {
		simulation.Send(procedures.UeTesterMessage{Type: procedures.Kill})
		waitShutdown(t, simulation.Done())
		finished := make(chan struct{})
		go func() { wg.Wait(); close(finished) }()
		waitShutdown(t, finished)
	})
	return simulation, connection, gnb, wg
}

func waitShutdown(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scenario shutdown blocked")
	}
}

func stopSimulationsForTest(t *testing.T, simulations ...*tools.UESimulation) {
	t.Helper()
	done := make(chan struct{})
	go func() { stopUESimulations(simulations); close(done) }()
	waitShutdown(t, done)
}

func TestScenarioShutdownAfterAssociationLoss(t *testing.T) {
	for _, looping := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "exhausted-loop"}[looping], func(t *testing.T) {
			simulation, connection, _, _ := startShutdownSimulation(t, looping, 1)
			close(connection.ConnectionLost)
			close(connection.GNBTx)
			waitShutdown(t, simulation.Done())
			stopSimulationsForTest(t, simulation)
			if simulation.Send(procedures.UeTesterMessage{Type: procedures.Terminate}) {
				t.Fatal("completed scenario accepted another command")
			}
		})
	}
}

func TestScenarioShutdownDuringRegistrationLoopDelay(t *testing.T) {
	simulation, connection, gnb, _ := startShutdownSimulation(t, true, 0)
	close(connection.ConnectionLost)
	close(connection.GNBTx)
	// The old RX closes during UE cleanup, before the 60-second restart delay.
	for range connection.GNBRx {
	}
	stopSimulationsForTest(t, simulation)
	select {
	case <-gnb.GetInboundChannel():
		t.Fatal("global shutdown restarted the registration loop")
	default:
	}
}

func TestScenarioShutdownAfterNormalIdleRelease(t *testing.T) {
	simulation, connection, _, _ := startShutdownSimulation(t, true, 0)
	simulation.Send(procedures.UeTesterMessage{Type: procedures.Idle})
	select {
	case message := <-connection.GNBRx:
		if !message.Idle {
			t.Fatalf("expected Idle request, got %+v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("UE did not process Idle")
	}
	close(connection.GNBTx)
	stopSimulationsForTest(t, simulation)
}

func TestScenarioShutdownWaitsForBusyLiveUE(t *testing.T) {
	simulation, connection, _, _ := startShutdownSimulation(t, false, 0)
	for range cap(connection.GNBRx) {
		connection.GNBRx <- gnbcontext.UEMessage{}
	}
	// The UE can block sending this registration, but its scenario must still
	// accept shutdown and observe completion after the association fails.
	simulation.Send(procedures.UeTesterMessage{Type: procedures.Registration})
	stopped := make(chan struct{})
	go func() { stopUESimulations([]*tools.UESimulation{simulation}); close(stopped) }()
	select {
	case <-stopped:
		t.Fatal("shutdown skipped a live UE with pending work")
	case <-time.After(20 * time.Millisecond):
	}
	close(connection.ConnectionLost)
	close(connection.GNBTx)
	waitShutdown(t, stopped)
}

func TestScenarioShutdownHandlesCompletedAndLiveScenarios(t *testing.T) {
	finished, firstConnection, _, _ := startShutdownSimulation(t, false, 0)
	live, _, _, _ := startShutdownSimulation(t, true, 0)
	close(firstConnection.ConnectionLost)
	close(firstConnection.GNBTx)
	waitShutdown(t, finished.Done())
	stopSimulationsForTest(t, finished, live)
}

func TestScenarioShutdownFollowsHandoverTargetAssociation(t *testing.T) {
	simulation, source, _, _ := startShutdownSimulation(t, false, 0)
	targetRX, targetTX := make(chan gnbcontext.UEMessage, 1), make(chan gnbcontext.UEMessage, 1)
	targetLost := make(chan struct{})
	source.GNBTx <- gnbcontext.UEMessage{
		GNBRx: targetRX, GNBTx: targetTX, ConnectionLost: targetLost,
		GNBInboundChannel: make(chan gnbcontext.UEMessage, 1),
	}
	select {
	case message := <-source.GNBRx:
		if !message.ConnectionClosed {
			t.Fatalf("source did not observe handover: %+v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("UE did not switch to the handover target")
	}
	close(source.ConnectionLost)
	close(source.GNBTx)
	simulation.Send(procedures.UeTesterMessage{Type: procedures.Registration})
	select {
	case message := <-targetRX:
		if !message.IsNas {
			t.Fatal("target did not receive the next UE command")
		}
	case <-simulation.Done():
		t.Fatal("source association loss terminated the handover target UE")
	case <-time.After(time.Second):
		t.Fatal("handover target UE stopped processing scenario commands")
	}
	close(targetLost)
	close(targetTX)
	waitShutdown(t, simulation.Done())
	stopSimulationsForTest(t, simulation)
}

// The next attempt is deliberately delayed for a minute by this fixture.
// Shutdown must release it without waiting for that timer or starting a new UE.
func TestScenarioLoopDelayStopsWhenGNBTerminates(t *testing.T) {
	simulation, connection, node, _ := startShutdownSimulation(t, true, 0)
	close(connection.ConnectionLost)
	close(connection.GNBTx)
	for range connection.GNBRx {
	}
	node.Terminate()
	waitShutdown(t, simulation.Done())
	for range node.GetInboundChannel() {
		t.Fatal("gNB termination created another registration attempt")
	}
}
