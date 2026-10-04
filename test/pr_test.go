/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {
	before := analytics.Results()

	controlIFConfig := netip.MustParseAddrPort("127.0.0.1:9489")
	dataIFConfig := netip.MustParseAddrPort("127.0.0.1:2154")
	amfListConfig := []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38414")}},
	}

	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfListConfig)

	type UECheck struct {
		HasAuthOnce  bool
		PduActivated map[int32]bool
	}

	ueChecks := map[string]*UECheck{}
	var checksMu sync.Mutex

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithPDUCallback(context.Active, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			if event != fsm.EntryEvent {
				return
			}
			checksMu.Lock()
			defer checksMu.Unlock()
			ue := args["ue"].(*context.UEContext)
			sm := args["sm"].(*context.SmContext)
			check := ueChecks[ue.GetSecurityContext().GetMsin()]
			if check.PduActivated == nil {
				check.PduActivated = map[int32]bool{}
			}
			check.PduActivated[sm.GetPduSessionId()] = true
		}).
		WithUeCallback(context.Authenticated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			if event != fsm.EntryEvent {
				return
			}
			checksMu.Lock()
			defer checksMu.Unlock()
			ue := args["ue"].(*context.UEContext)
			check, ok := ueChecks[ue.GetSecurityContext().GetMsin()]
			if !ok {
				check = &UECheck{}
				ueChecks[ue.GetSecurityContext().GetMsin()] = check
			}
			check.HasAuthOnce = true
		}).
		Build()
	require.NoError(t, err)
	t.Cleanup(func() { _ = fiveGC.Close() })

	// Setup gNodeB
	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs, err := tools.CreateGnbs(t.Context(), gnbCount, conf, &wg)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, gnb := range gnbs {
			gnb.Terminate()
		}
	})

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	// Setup UE
	ueCount := 10
	scenarioChans := make([]chan procedures.UeTesterMessage, ueCount+1)
	ueSimCfg := tools.UESimulationConfig{
		Gnbs:                     gnbs,
		Cfg:                      conf,
		TimeBeforeDeregistration: 0,
		TimeBeforeNgapHandover:   0,
		TimeBeforeXnHandover:     0,
		NumPduSessions:           1,
		RegistrationLoop:         false,
	}

	simulations := make([]*tools.UESimulation, 0, ueCount)
	t.Cleanup(func() { stopTestSimulations(t, simulations) })
	for ueSimCfg.UeId = 1; ueSimCfg.UeId <= ueCount; ueSimCfg.UeId++ {
		ueSimCfg.ScenarioChan = scenarioChans[ueSimCfg.UeId]

		imsi := tools.IncrementMsin(ueSimCfg.UeId, ueSimCfg.Cfg.Ue.Msin)

		securityContext := context.SecurityContext{}
		securityContext.SetMsin(imsi)
		securityContext.SetAuthSubscription(ueSimCfg.Cfg.Ue.Key, ueSimCfg.Cfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})

		amfContext := fiveGC.GetAMFContext()
		amfContext.Provision(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd}, securityContext)

		simulations = append(simulations, tools.SimulateSingleUE(ueSimCfg, &wg))

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	// Terminate only after the client has really accepted every PDU session,
	// rather than assuming registration completes before a short wall-clock
	// timer. Keep the original success and deregistration assertions below.
	require.Eventually(t, func() bool {
		return analytics.Results()[analytics.SessionEstablishment].Success-before[analytics.SessionEstablishment].Success == uint64(ueCount)
	}, 30*time.Second, 10*time.Millisecond, "all clients must accept a PDU session")
	for _, simulation := range simulations {
		require.True(t, simulation.Send(procedures.UeTesterMessage{Type: procedures.Terminate}))
	}
	waitTestSimulations(t, simulations, 30*time.Second)
	require.Eventually(t, func() bool {
		allDeregistered := true
		fiveGC.GetAMFContext().ExecuteForAllUe(func(ue *context.UEContext) {
			allDeregistered = allDeregistered && ue.GetState().Is(context.Deregistered)
		})
		return allDeregistered
	}, 5*time.Second, 10*time.Millisecond, "the mock core must process deregistration")
	i := 0
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			i++
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			checksMu.Lock()
			defer checksMu.Unlock()
			check := ueChecks[ue.GetSecurityContext().GetMsin()]
			require.NotNil(t, check)
			assert.Equal(t, map[int32]bool{1: true}, check.PduActivated, "PDU session must have been activated before deregistration")
			assert.True(t, check.HasAuthOnce, "UE has never changed state")
			ue.ExecuteForAllSmContexts(
				func(sm *context.SmContext) {
					assert.Equalf(t, context.Inactive, sm.GetState().Current(), "Expected all pdu sessions to be in Inactive state but was not")
					assert.True(t, check.PduActivated[sm.GetPduSessionId()], "Expected all pdu to be activate once but was not")
				})
		})
	assert.Equalf(t, ueCount, i, "Expected %v ue to created in 5GC state but was %v", ueCount, i)
	for i, procedure := range analytics.Results() {
		assert.Equal(t, uint64(ueCount), procedure.Started-before[i].Started, procedure.Procedure)
		assert.Equal(t, uint64(ueCount), procedure.Success-before[i].Success, procedure.Procedure)
		assert.Equal(t, before[i].Failure, procedure.Failure, procedure.Procedure)
		assert.Equal(t, before[i].Pending, procedure.Pending, procedure.Procedure)
	}

}

func TestUERegistrationLoop(t *testing.T) {
	before := analytics.Results()

	controlIFConfig := netip.MustParseAddrPort("127.0.0.1:9490")
	dataIFConfig := netip.MustParseAddrPort("127.0.0.1:2155")
	amfListConfig := []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38415")}},
	}

	type UECheck struct {
		authCounter int
	}
	ueChecks := map[string]*UECheck{}
	var checksMu sync.Mutex

	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfListConfig)

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithUeCallback(context.Authenticated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			if event != fsm.EntryEvent {
				return
			}
			checksMu.Lock()
			defer checksMu.Unlock()
			ue := args["ue"].(*context.UEContext)
			check, ok := ueChecks[ue.GetSecurityContext().GetMsin()]
			if !ok {
				check = &UECheck{}
				ueChecks[ue.GetSecurityContext().GetMsin()] = check
			}
			check.authCounter++
		}).
		Build()
	require.NoError(t, err)
	t.Cleanup(func() { _ = fiveGC.Close() })

	// Setup gNodeB
	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs, err := tools.CreateGnbs(t.Context(), gnbCount, conf, &wg)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, gnb := range gnbs {
			gnb.Terminate()
		}
	})

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	// Setup UE
	scenarioChans := make([]chan procedures.UeTesterMessage, 2)
	deregistrationTrigger := make(chan struct{})
	ueSimCfg := tools.UESimulationConfig{
		UeId:                     1,
		Gnbs:                     gnbs,
		Cfg:                      conf,
		TimeBeforeDeregistration: 2000,
		DeregistrationTrigger:    deregistrationTrigger,
		TimeBeforeNgapHandover:   0,
		TimeBeforeXnHandover:     0,
		NumPduSessions:           1,
		RegistrationLoop:         true,
		LoopCount:                5,
	}
	scenarioChans[ueSimCfg.UeId] = make(chan procedures.UeTesterMessage)
	ueSimCfg.ScenarioChan = scenarioChans[ueSimCfg.UeId]

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(tools.IncrementMsin(ueSimCfg.UeId, ueSimCfg.Cfg.Ue.Msin))
	securityContext.SetAuthSubscription(ueSimCfg.Cfg.Ue.Key, ueSimCfg.Cfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})

	amfContext := fiveGC.GetAMFContext()
	amfContext.Provision(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd}, securityContext)

	simulation := tools.SimulateSingleUE(ueSimCfg, &wg)
	t.Cleanup(func() { stopTestSimulations(t, []*tools.UESimulation{simulation}) })

	// Each iteration must really finish registration and accept its PDU session
	// before we exercise graceful teardown. A timer started before attach can
	// deliberately abort an authenticated UE during SCTP recovery instead.
	deadline := time.Now().Add(45 * time.Second)
	for iteration := 1; iteration <= ueSimCfg.LoopCount; iteration++ {
		require.Eventually(t, func() bool {
			for i, procedure := range analytics.Results() {
				if procedure.Success-before[i].Success != uint64(iteration) {
					return false
				}
			}
			return true
		}, time.Until(deadline), 10*time.Millisecond, "iteration %d must complete both client procedures", iteration)
		timer := time.NewTimer(time.Until(deadline))
		select {
		case deregistrationTrigger <- struct{}{}:
		case <-simulation.Done():
			timer.Stop()
			t.Fatalf("scenario ended before iteration %d teardown", iteration)
		case <-timer.C:
			t.Fatalf("scenario did not accept iteration %d teardown before the deadline", iteration)
		}
		timer.Stop()
	}

	// Join the whole loop before inspecting it or closing its gNB inbound
	// channel. A fixed sleep can expire while the next UE is still attaching.
	waitTestSimulations(t, []*tools.UESimulation{simulation}, time.Until(deadline))
	require.Eventually(t, func() bool {
		allDeregistered := true
		fiveGC.GetAMFContext().ExecuteForAllUe(func(ue *context.UEContext) {
			allDeregistered = allDeregistered && ue.GetState().Is(context.Deregistered)
		})
		return allDeregistered
	}, 5*time.Second, 10*time.Millisecond, "the mock core must process the last deregistration")
	ueCount := 0
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			ueCount++
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			checksMu.Lock()
			defer checksMu.Unlock()
			check := ueChecks[ue.GetSecurityContext().GetMsin()]
			require.NotNil(t, check)
			assert.Equal(t, 5, check.authCounter, "each loop must authenticate once")
		})
	assert.Equal(t, ueSimCfg.LoopCount, ueCount, "each loop must create a core UE context")
	for i, procedure := range analytics.Results() {
		assert.Equal(t, uint64(ueSimCfg.LoopCount), procedure.Started-before[i].Started, procedure.Procedure)
		assert.Equal(t, uint64(ueSimCfg.LoopCount), procedure.Success-before[i].Success, procedure.Procedure)
		assert.Equal(t, before[i].Failure, procedure.Failure, procedure.Procedure)
		assert.Equal(t, before[i].Pending, procedure.Pending, procedure.Procedure)
	}
}

func waitTestSimulations(t *testing.T, simulations []*tools.UESimulation, timeout time.Duration) {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for _, simulation := range simulations {
		select {
		case <-simulation.Done():
		case <-deadline.C:
			t.Fatal("UE scenarios did not finish before the test deadline")
		}
	}
}

func stopTestSimulations(t *testing.T, simulations []*tools.UESimulation) {
	t.Helper()
	for _, simulation := range simulations {
		simulation.Send(procedures.UeTesterMessage{Type: procedures.Kill})
	}
	waitTestSimulations(t, simulations, 30*time.Second)
}
