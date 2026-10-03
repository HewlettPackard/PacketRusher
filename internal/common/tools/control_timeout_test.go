// SPDX-License-Identifier: Apache-2.0
package tools

import (
	"context"
	"errors"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	coreTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"os"
	"sync"
	"testing"
	"time"

	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
)

func TestAutomaticTerminationTimeoutProgressesLoopsAndRejectsOldIterations(t *testing.T) {
	results := analytics.NewRecorder()
	analytics.SetCurrent(results)
	t.Cleanup(func() { analytics.SetCurrent(nil) })
	conf := coreTools.GenerateDefaultConf(netip.AddrPortFrom(netip.MustParseAddr("127.0.0.61"), uint16(26000+os.Getpid()%2000)), netip.MustParseAddrPort("127.0.0.61:2164"), []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(39000+os.Getpid()%5000))}}})
	builder := aio5gc.FiveGCBuilder{}
	fgc, err := builder.WithConfig(conf).WithNASDispatcherHook(nas.MsgTypeULNASTransport, func(nas.Message, *core.UEContext, *core.GNBContext, *core.Aio5gc) (bool, error) {
		// Integrity-checked, decoded native PDU requests arrive normally; the
		// core deliberately never accepts them. Terminate must still run.
		return true, nil
	}).Build()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fgc.Close()) })
	security := core.SecurityContext{}
	security.SetMsin(conf.Ue.Msin)
	security.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	security.SetAbba([]byte{0, 0})
	require.NoError(t, fgc.GetAMFContext().Provision(models.Snssai{Sst: 1, Sd: conf.Ue.Snssai.Sd}, security))
	var wg sync.WaitGroup
	gnbs := CreateGnbs(1, conf, &wg)
	t.Cleanup(func() {
		for _, node := range gnbs {
			node.Terminate()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	require.NoError(t, WaitGnbs(ctx, gnbs))
	sim := SimulateSingleUE(UESimulationConfig{UeId: 1, Cfg: conf, Gnbs: gnbs, NumPduSessions: 1, RegistrationLoop: true, LoopCount: 2}, &wg)
	t.Cleanup(func() {
		sim.Send(procedures.UeTesterMessage{Type: procedures.Kill})
		select {
		case <-sim.Done():
		case <-time.After(5 * time.Second):
			t.Error("simulation did not stop")
		}
	})
	waitRegistered := func(generation uint64) {
		t.Helper()
		require.Eventually(t, func() bool {
			a, err := sim.Inspect(ctx)
			return err == nil && a.Generation == generation && a.State == "registered" && !a.Ready
		}, 15*time.Second, 20*time.Millisecond)
	}
	waitRegistered(1)
	cancelled, stop := context.WithCancel(ctx)
	stop()
	require.ErrorIs(t, sim.terminateWhenReady(cancelled, 1, 100*time.Millisecond), context.Canceled)
	require.ErrorIs(t, sim.terminateWhenReady(ctx, 2, 100*time.Millisecond), procedures.ErrGeneration)
	// The production helper's short injected deadline executes the original
	// release/deregistration path and permits the next registration iteration.
	err = sim.terminateWhenReady(ctx, 1, 100*time.Millisecond)
	require.True(t, err == nil || errors.Is(err, context.DeadlineExceeded), "%v", err)
	waitRegistered(2)
	require.ErrorIs(t, sim.terminateWhenReady(ctx, 1, 100*time.Millisecond), procedures.ErrGeneration)
	owningIteration, cancelIteration := context.WithCancel(ctx)
	finished := make(chan error, 1)
	go func() { finished <- sim.terminateWhenReady(owningIteration, 2, time.Second) }()
	require.Eventually(t, func() bool { return len(sim.controlGate) == 1 }, time.Second, time.Millisecond)
	cancelIteration()
	select {
	case err := <-finished:
		require.ErrorIs(t, err, context.Canceled)
	case <-ctx.Done():
		t.Fatal("cancelled automatic readiness wait did not return")
	}
	a, err := sim.Inspect(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), a.Generation)
	require.Equal(t, "registered", a.State, "cancelled owning iteration must not invoke timeout termination")
	err = sim.terminateWhenReady(ctx, 2, 100*time.Millisecond)
	require.True(t, err == nil || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, procedures.ErrStopped), "%v", err)
	select {
	case <-sim.Done():
	case <-ctx.Done():
		t.Fatal("timeout did not finish the final registration iteration")
	}
	require.Eventually(t, func() bool {
		seen, complete := 0, true
		fgc.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) {
			seen++
			complete = complete && ue.GetState().Current() == core.Deregistered
		})
		return seen == 2 && complete
	}, 5*time.Second, 20*time.Millisecond)
	for _, procedure := range results.Snapshot().Procedures {
		require.Equal(t, uint64(2), procedure.Started)
		require.Zero(t, procedure.Failure)
		require.Zero(t, procedure.Pending)
		if procedure.Procedure == analytics.SessionEstablishment {
			require.Zero(t, procedure.Success)
			require.Equal(t, uint64(2), procedure.Cancelled)
		} else {
			require.Equal(t, uint64(2), procedure.Success)
			require.Zero(t, procedure.Cancelled)
		}
	}
}
