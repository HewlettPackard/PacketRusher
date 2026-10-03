// SPDX-License-Identifier: Apache-2.0
package testkit

import (
	"context"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	"testing"
	"time"
)

func TestCoreCancellationReleasesReachedGateAndJoinsWorker(t *testing.T) {
	var fgc core.Aio5gc
	gate := NewGate()
	result := make(chan error, 1)
	require.True(t, fgc.Go(func() { result <- gate.Wait(fgc.Context()) }))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-gate.Reached():
	case <-ctx.Done():
		t.Fatal("gate was not reached")
	}
	require.NoError(t, fgc.CloseContext(ctx))
	require.ErrorIs(t, <-result, context.Canceled)
	gate.Open()
	gate.Open()
	require.NoError(t, gate.Wait(context.Background()))
}

func TestCanceledStartupDoesNotBuildOrBindCore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fixture, err := Start(ctx, new(aio5gc.FiveGCBuilder), 1)
	require.Nil(t, fixture)
	require.ErrorIs(t, err, context.Canceled)
}
