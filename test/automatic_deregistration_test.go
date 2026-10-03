// SPDX-License-Identifier: Apache-2.0
package test

import (
	stdcontext "context"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/tools"
	core "my5G-RANTester/test/aio5gc/context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticRegistrationToCtxReleaseWithPDUSession(t *testing.T) {
	results := analytics.NewRecorder()
	analytics.SetCurrent(results)
	t.Cleanup(func() { analytics.SetCurrent(nil) })
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 30*time.Second)
	defer cancel()
	fixture := registrationFixture(t, ctx)
	const count = 10
	simulations := make([]*tools.UESimulation, 0, count)
	for id := 1; id <= count; id++ {
		require.NoError(t, fixture.Provision(id))
		simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: id, NumPduSessions: 1, TimeBeforeDeregistration: 400})
		require.NoError(t, err)
		simulations = append(simulations, simulation)
	}
	// Preserve the actual automatic timers. Join the actors, then observe core
	// deregistration and inactive sessions rather than sleeping for an estimate.
	waitTestSimulations(t, simulations, time.Until(deadlineOf(ctx)))
	snapshot, err := fixture.Core.Wait(ctx, allCoreDeregistered(count))
	require.NoError(t, err)
	require.Empty(t, snapshot.Errors)
	require.Len(t, snapshot.UEs, count)
	for _, ue := range snapshot.UEs {
		assert.Equal(t, core.Deregistered, ue.State)
		assert.Equal(t, uint64(1), ue.Entries[core.Authenticated], "UE must authenticate once")
		require.Len(t, ue.Sessions, 1)
		assert.Equal(t, uint64(1), ue.Sessions[1].Entries[core.Active], "PDU session must have been activated before deregistration")
		assert.Equal(t, core.Inactive, ue.Sessions[1].State)
	}
	observed := 0
	fixture.Core.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) {
		observed++
		assert.Equal(t, core.Deregistered, ue.GetState().Current())
		ue.ExecuteForAllSmContexts(func(sm *core.SmContext) { assert.Equal(t, core.Inactive, sm.GetState().Current()) })
	})
	assert.Equal(t, count, observed)
	for _, procedure := range results.Snapshot().Procedures {
		assert.Equal(t, uint64(count), procedure.Started)
		assert.Equal(t, uint64(count), procedure.Success)
		assert.Zero(t, procedure.Failure)
		assert.Zero(t, procedure.Pending)
	}
}
func TestAutomaticUERegistrationLoop(t *testing.T) {
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 30*time.Second)
	defer cancel()
	fixture := registrationFixture(t, ctx)
	require.NoError(t, fixture.Provision(1))
	const loops = 5
	simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: 1, NumPduSessions: 1, TimeBeforeDeregistration: 2000, RegistrationLoop: true, LoopCount: loops})
	require.NoError(t, err)
	waitTestSimulations(t, []*tools.UESimulation{simulation}, time.Until(deadlineOf(ctx)))
	snapshot, err := fixture.Core.Wait(ctx, allCoreDeregistered(loops))
	require.NoError(t, err)
	require.Empty(t, snapshot.Errors)
	require.Len(t, snapshot.UEs, loops)
	authCount := uint64(0)
	for _, ue := range snapshot.UEs {
		assert.Equal(t, core.Deregistered, ue.State)
		authCount += ue.Entries[core.Authenticated]
		assert.Equal(t, uint64(1), ue.Entries[core.Authenticated])
		assert.Equal(t, uint64(1), ue.Sessions[1].Entries[core.Active])
	}
	assert.Equal(t, uint64(loops), authCount, "each loop must authenticate once")
	fixture.Core.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) { assert.Equal(t, core.Deregistered, ue.GetState().Current()) })
}
