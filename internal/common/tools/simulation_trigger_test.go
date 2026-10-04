// SPDX-License-Identifier: Apache-2.0
package tools

import (
	"context"
	"net/netip"
	"sync"
	"testing"
	"time"

	nas "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
)

type simulationHarness struct {
	simulation *UESimulation
	inbound    chan gnbcontext.UEMessage
	gnb        *gnbcontext.GNBContext
}

// Exercise the real scenario and UE goroutines without an SCTP listener. The
// harness owns only the gNB greeting and NAS frames needed by these lifecycle
// checks; complete authentication/PDU establishment remains covered by ./test.
func newSimulationHarness(t *testing.T, trigger <-chan struct{}, interval, loops int) *simulationHarness {
	return newSimulationHarnessWithRestartDelay(t, trigger, interval, loops, 0)
}

func newSimulationHarnessWithRestartDelay(t *testing.T, trigger <-chan struct{}, interval, loops, restartDelay int) *simulationHarness {
	t.Helper()
	gnb := &gnbcontext.GNBContext{}
	gnb.NewRanGnbContext("000008", "001", "01", "000001", "01", "", netip.MustParseAddrPort("127.0.0.1:9487"), netip.MustParseAddrPort("127.0.0.1:2152"))
	cfg := config.Config{
		GNodeB: config.GNodeB{PlmnList: config.PlmnList{GnbId: "000008"}},
		Ue: config.Ue{
			Msin: "0000000120", Hplmn: config.Hplmn{Mcc: "001", Mnc: "01"},
			RoutingIndicator: "0000", Snssai: config.Snssai{Sst: 1},
			Ciphering: config.Ciphering{Nea0: true}, Integrity: config.Integrity{Nia0: true},
		},
	}
	// Unit fixture for successful NG Setup/attachment. Native wire readiness is
	// covered separately by the real SCTP scenario fixtures.
	gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412")).SetStateActive()
	t.Cleanup(gnb.Terminate)
	wg := &sync.WaitGroup{}
	h := &simulationHarness{inbound: gnb.GetInboundChannel(), gnb: gnb}
	h.simulation = SimulateSingleUE(UESimulationConfig{
		UeId: 1, Gnbs: map[string]*gnbcontext.GNBContext{"000008": gnb}, Cfg: cfg,
		TimeBeforeDeregistration: interval, DeregistrationTrigger: trigger,
		RegistrationLoop: true, LoopCount: loops,
		TimeBeforeReregistration: restartDelay,
	}, wg)
	t.Cleanup(func() {
		finished := make(chan struct{})
		go func() {
			h.simulation.Send(procedures.UeTesterMessage{Type: procedures.Kill})
			wg.Wait()
			close(finished)
		}()
		select {
		case <-finished:
		case <-time.After(10 * time.Second): // Also covers InitConn's bounded greeting wait.
			t.Error("scenario cleanup did not join all UE goroutines")
		}
	})
	return h
}

func TestRegistrationRestartRepliesToControlRequests(t *testing.T) {
	h := newSimulationHarnessWithRestartDelay(t, nil, 0, 2, 60_000)
	first := h.nextUE(t)
	close(first.ConnectionLost)
	expectConnectionClosed(t, first)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var a procedures.Attachment
	var err error
	for {
		a, err = h.simulation.Inspect(ctx)
		if err == nil && a.Generation == 2 && a.State == "starting" {
			break
		}
		// The UE closes its connection before its scenario channel. A request
		// admitted just before actor exit may receive ErrGeneration; it must
		// still reply, and subsequent inspection must reach the restart delay.
		if err != nil {
			require.ErrorIs(t, err, procedures.ErrGeneration)
		}
		select {
		case <-ctx.Done():
			t.Fatal("inspection did not receive a reply during the restart delay")
		case <-time.After(time.Millisecond):
		}
	}
	require.Equal(t, uint64(2), a.Generation)
	require.Equal(t, "starting", a.State)
	require.False(t, a.Connected)
	require.False(t, a.Ready)
	require.Empty(t, a.ActivePDUSessions)
	_, err = h.simulation.request(ctx, "terminate-after-timeout", "", 1, 0)
	require.ErrorIs(t, err, procedures.ErrGeneration)
	_, err = h.simulation.request(ctx, "idle", "", 2, 0)
	require.ErrorIs(t, err, procedures.ErrNotReady)
	_, err = h.simulation.request(ctx, "inspect", "", 2, 1)
	require.ErrorIs(t, err, procedures.ErrGeneration)
	cancelled, stop := context.WithCancel(ctx)
	stop()
	reply := make(chan procedures.ControlResult, 1)
	require.True(t, h.simulation.Send(procedures.UeTesterMessage{Type: procedures.Control,
		Control: &procedures.ControlRequest{Context: cancelled, Action: "inspect", Reply: reply}}))
	select {
	case result := <-reply:
		require.ErrorIs(t, result.Err, context.Canceled)
	case <-ctx.Done():
		t.Fatal("cancelled request was consumed without a reply")
	}
	h.send(t, procedures.UeTesterMessage{Type: procedures.Kill})
	waitSimulationDone(t, h.simulation)
}

func (h *simulationHarness) nextUE(t *testing.T) gnbcontext.UEMessage {
	t.Helper()
	select {
	case connection := <-h.inbound:
		connection.GNBTx <- gnbcontext.UEMessage{Mcc: "001", Mnc: "01"}
		expectUplink(t, connection, nas.MsgTypeRegReq)
		gu, err := h.gnb.NewGnBUe(connection.GNBTx, connection.GNBRx, connection.PrUeId, nil)
		if err != nil {
			t.Fatal(err)
		}
		gu.SetStateReady()
		return connection
	case <-time.After(2 * time.Second):
		t.Fatal("scenario did not create its next UE")
		return gnbcontext.UEMessage{}
	}
}

func (h *simulationHarness) send(t *testing.T, command procedures.UeTesterMessage) {
	t.Helper()
	accepted := make(chan bool, 1)
	go func() { accepted <- h.simulation.Send(command) }()
	select {
	case ok := <-accepted:
		if !ok {
			t.Fatal("global shutdown was not accepted")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("global shutdown stopped receiving commands")
	}
}

func expectUplink(t *testing.T, connection gnbcontext.UEMessage, want nas.MsgType) {
	t.Helper()
	select {
	case message, open := <-connection.GNBRx:
		if !open || !message.IsNas {
			t.Fatal("expected a NAS uplink before connection closure")
		}
		payload := message.Nas
		if nas.GetSecHdrType(payload) != nas.SecHdrTypePlainNas {
			// This harness uses null cipher/integrity algorithms. Do not use this
			// shortcut in protocol/security tests, which verify protected NAS.
			if len(payload) < 7 {
				t.Fatal("truncated protected NAS")
			}
			payload = payload[7:]
		}
		decoded, err := nas.Parse(payload, nil)
		if err != nil || decoded == nil || decoded.MsgType() != want {
			t.Fatalf("uplink = %x (%v), want NAS type %v", payload, err, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("UE did not send the expected NAS uplink")
	}
}

func registerHarnessUE(t *testing.T, connection gnbcontext.UEMessage) {
	t.Helper()
	// Independent plain Registration Accept literal, for orchestration only.
	connection.GNBTx <- gnbcontext.UEMessage{IsNas: true, Nas: []byte{0x7e, 0, 0x42, 1, 1}}
	expectUplink(t, connection, nas.MsgTypeRegComplete)
}

func expectConnectionClosed(t *testing.T, connection gnbcontext.UEMessage) {
	t.Helper()
	select {
	case _, open := <-connection.GNBRx:
		if open {
			t.Fatal("unexpected uplink before expected connection closure")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("UE connection did not close")
	}
}

func expectConnectionAlive(t *testing.T, connection gnbcontext.UEMessage) {
	t.Helper()
	select {
	case <-connection.GNBRx:
		t.Fatal("explicit trigger unexpectedly used the one-millisecond timer")
	case <-time.After(20 * time.Millisecond):
	}
}

func sendIterationTrigger(t *testing.T, trigger chan<- struct{}) {
	t.Helper()
	select {
	case trigger <- struct{}{}:
	case <-time.After(2 * time.Second):
		t.Fatal("scenario did not accept its iteration trigger")
	}
}

func waitSimulationDone(t *testing.T, simulation *UESimulation) {
	t.Helper()
	select {
	case <-simulation.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("scenario did not finish")
	}
}

func TestNilDeregistrationTriggerWaitsForRegistrationBeforeTimedLoops(t *testing.T) {
	h := newSimulationHarness(t, nil, 25, 3)
	for iteration := 0; iteration < 3; iteration++ {
		connection := h.nextUE(t)
		// The scenario controls intentionally start td after registration. This
		// is longer than td and must not abort an unregistered attachment.
		select {
		case <-connection.GNBRx:
			t.Fatal("timer ended an unregistered iteration")
		case <-time.After(50 * time.Millisecond):
		}
		registerHarnessUE(t, connection)
		expectUplink(t, connection, nas.MsgTypeDeregReqUEOrig)
		expectConnectionClosed(t, connection)
	}
	waitSimulationDone(t, h.simulation)
}

func TestUnregisteredTimedSimulationRemainsGloballyCancellable(t *testing.T) {
	h := newSimulationHarness(t, nil, 1, 0)
	connection := h.nextUE(t)
	expectConnectionAlive(t, connection)
	h.send(t, procedures.UeTesterMessage{Type: procedures.Kill})
	expectConnectionClosed(t, connection)
	waitSimulationDone(t, h.simulation)
}

func TestDeregistrationTriggerOverridesTimerAndGracefullyRepeats(t *testing.T) {
	trigger := make(chan struct{})
	h := newSimulationHarness(t, trigger, 1, 3)
	for iteration := 0; iteration < 3; iteration++ {
		connection := h.nextUE(t)
		expectConnectionAlive(t, connection)
		registerHarnessUE(t, connection)
		sendIterationTrigger(t, trigger)
		expectUplink(t, connection, nas.MsgTypeDeregReqUEOrig)
		expectConnectionClosed(t, connection)
	}
	waitSimulationDone(t, h.simulation)
}

func TestClosedDeregistrationTriggerStaysDisabledAcrossIterations(t *testing.T) {
	trigger := make(chan struct{})
	h := newSimulationHarness(t, trigger, 1, 3)
	first := h.nextUE(t)
	close(trigger)
	expectConnectionAlive(t, first)
	// An association loss ends this UE independently of a scenario stop, so the
	// real loop restarts and must not re-arm the closed trigger or default timer.
	close(first.ConnectionLost)
	expectConnectionClosed(t, first)
	second := h.nextUE(t)
	expectConnectionAlive(t, second)
	registerHarnessUE(t, second)
	h.send(t, procedures.UeTesterMessage{Type: procedures.Terminate})
	expectUplink(t, second, nas.MsgTypeDeregReqUEOrig)
	expectConnectionClosed(t, second)
	waitSimulationDone(t, h.simulation)
}

func TestPendingDeregistrationTriggerAllowsGlobalShutdown(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command procedures.UeTesterMessage
	}{{"kill", procedures.UeTesterMessage{Type: procedures.Kill}}, {"terminate", procedures.UeTesterMessage{Type: procedures.Terminate}}} {
		t.Run(tc.name, func(t *testing.T) {
			h := newSimulationHarness(t, make(chan struct{}), 1, 0)
			connection := h.nextUE(t)
			registerHarnessUE(t, connection)
			h.send(t, tc.command)
			if tc.command.Type == procedures.Terminate {
				expectUplink(t, connection, nas.MsgTypeDeregReqUEOrig)
			}
			expectConnectionClosed(t, connection)
			waitSimulationDone(t, h.simulation)
		})
	}
}

func TestRegistrationLoopStopsWhenGNBTerminates(t *testing.T) {
	h := newSimulationHarness(t, nil, 1, 0)
	connection := h.nextUE(t)
	h.gnb.Terminate()
	expectConnectionClosed(t, connection)
	waitSimulationDone(t, h.simulation)
	for range h.inbound {
		t.Fatal("registration loop created another UE after gNB termination")
	}
}
