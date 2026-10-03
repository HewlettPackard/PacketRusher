/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package test

import (
	stdcontext "context"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/testkit"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registrationFixture(t *testing.T, ctx stdcontext.Context) *testkit.Fixture {
	t.Helper()
	builder := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig())
	fixture, err := testkit.Start(ctx, builder, 1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fixture.Close()) })
	return fixture
}
func allCoreDeregistered(count int) func(core.Snapshot) bool {
	return func(snapshot core.Snapshot) bool {
		if len(snapshot.UEs) != count {
			return false
		}
		for _, ue := range snapshot.UEs {
			if ue.State != core.Deregistered {
				return false
			}
			for _, pdu := range ue.Sessions {
				if pdu.State != core.Inactive {
					return false
				}
			}
		}
		return true
	}
}

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {
	results := analytics.NewRecorder()
	analytics.SetCurrent(results)
	t.Cleanup(func() { analytics.SetCurrent(nil) })
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 30*time.Second)
	defer cancel()
	fixture := registrationFixture(t, ctx)
	const ueCount = 10
	simulations := make([]*tools.UESimulation, 0, ueCount)
	for id := 1; id <= ueCount; id++ {
		require.NoError(t, fixture.Provision(id))
		simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: id, NumPduSessions: 1})
		require.NoError(t, err)
		simulations = append(simulations, simulation)
	}
	// Actual UE actors must accept every encoded PDU response before teardown.
	for _, simulation := range simulations {
		attachment, err := simulation.Execute(ctx, "wait", "")
		require.NoError(t, err)
		require.Equal(t, []uint8{1}, attachment.ActivePDUSessions)
	}
	for _, simulation := range simulations {
		require.True(t, simulation.Send(procedures.UeTesterMessage{Type: procedures.Terminate}))
	}
	waitTestSimulations(t, simulations, time.Until(deadlineOf(ctx)))
	snapshot, err := fixture.Core.Wait(ctx, allCoreDeregistered(ueCount))
	require.NoError(t, err)
	assert.Len(t, snapshot.UEs, ueCount)
	require.Empty(t, snapshot.Errors)
	for _, ue := range snapshot.UEs {
		assert.Equal(t, core.Deregistered, ue.State)
		assert.Equal(t, uint64(1), ue.Entries[core.Authenticated], "each UE must authenticate once")
		require.Len(t, ue.Sessions, 1)
		assert.Equal(t, uint64(1), ue.Sessions[1].Entries[core.Active], "PDU session must have been active before deregistration")
		assert.Equal(t, core.Inactive, ue.Sessions[1].State)
	}
	// Retain assertions on the actual live context pool and analytics recorder.
	fixture.Core.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) {
		assert.Equal(t, core.Deregistered, ue.GetState().Current())
		ue.ExecuteForAllSmContexts(func(sm *core.SmContext) { assert.Equal(t, core.Inactive, sm.GetState().Current()) })
	})
	for _, procedure := range results.Snapshot().Procedures {
		assert.Equal(t, uint64(ueCount), procedure.Started)
		assert.Equal(t, uint64(ueCount), procedure.Success)
		assert.Zero(t, procedure.Failure)
		assert.Zero(t, procedure.Pending)
	}
}

func TestUERegistrationLoop(t *testing.T) {
	results := analytics.NewRecorder()
	analytics.SetCurrent(results)
	t.Cleanup(func() { analytics.SetCurrent(nil) })
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 45*time.Second)
	defer cancel()
	fixture := registrationFixture(t, ctx)
	require.NoError(t, fixture.Provision(1))
	trigger := make(chan struct{})
	const loops = 5
	simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: 1, NumPduSessions: 1, TimeBeforeDeregistration: 2000, DeregistrationTrigger: trigger, RegistrationLoop: true, LoopCount: loops})
	require.NoError(t, err)
	for iteration := 1; iteration <= loops; iteration++ {
		// Core transitions identify the expected iteration; client readiness then
		// proves its native NAS accept has been processed before the explicit gate.
		_, err := fixture.Core.Wait(ctx, func(snapshot core.Snapshot) bool {
			return len(snapshot.UEs) == iteration && snapshot.UEs[iteration-1].State == core.Registered && snapshot.UEs[iteration-1].Sessions[1].State == core.Active
		})
		require.NoError(t, err)
		attachment, err := simulation.Execute(ctx, "wait", "")
		require.NoError(t, err)
		require.Equal(t, []uint8{1}, attachment.ActivePDUSessions)
		select {
		case trigger <- struct{}{}:
		case <-simulation.Done():
			t.Fatalf("scenario ended before iteration %d teardown", iteration)
		case <-ctx.Done():
			t.Fatalf("iteration %d teardown: %v", iteration, ctx.Err())
		}
	}
	waitTestSimulations(t, []*tools.UESimulation{simulation}, time.Until(deadlineOf(ctx)))
	snapshot, err := fixture.Core.Wait(ctx, allCoreDeregistered(loops))
	require.NoError(t, err)
	require.Empty(t, snapshot.Errors)
	authCount := uint64(0)
	for _, ue := range snapshot.UEs {
		assert.Equal(t, core.Deregistered, ue.State)
		authCount += ue.Entries[core.Authenticated]
	}
	assert.Equal(t, uint64(loops), authCount, "each loop must authenticate once")
	assert.Len(t, snapshot.UEs, loops)
	fixture.Core.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) { assert.Equal(t, core.Deregistered, ue.GetState().Current()) })
	for _, procedure := range results.Snapshot().Procedures {
		assert.Equal(t, uint64(loops), procedure.Started)
		assert.Equal(t, uint64(loops), procedure.Success)
		assert.Zero(t, procedure.Failure)
		assert.Zero(t, procedure.Cancelled)
		assert.Zero(t, procedure.Pending)
	}
}
func deadlineOf(ctx stdcontext.Context) time.Time { deadline, _ := ctx.Deadline(); return deadline }

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
