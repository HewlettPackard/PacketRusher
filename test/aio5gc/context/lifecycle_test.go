package context

import (
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestCloseDrainsWorkersAndRejectsNewWork(t *testing.T) {
	var core Aio5gc
	done := make(chan struct{})
	var closes atomic.Int32
	core.RegisterCloser(func() error { closes.Add(1); close(done); return nil })
	require.True(t, core.BeginWork())
	go func() { defer core.EndWork(); <-done }()
	require.NoError(t, core.Close())
	require.NoError(t, core.Close())
	require.Equal(t, int32(1), closes.Load())
	require.False(t, core.BeginWork())
	core.RegisterCloser(func() error { closes.Add(1); return nil })
	require.Equal(t, int32(2), closes.Load(), "late associations must close immediately")
}

func TestConcurrentCloseWaitsForAllClosers(t *testing.T) {
	var core Aio5gc
	started, release := make(chan struct{}), make(chan struct{})
	core.RegisterCloser(func() error { close(started); <-release; return nil })
	first := make(chan struct{})
	go func() { _ = core.Close(); close(first) }()
	<-started
	second := make(chan struct{})
	go func() { _ = core.Close(); close(second) }()
	select {
	case <-second:
		t.Fatal("concurrent Close returned before shutdown finished")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	<-first
	<-second
}
