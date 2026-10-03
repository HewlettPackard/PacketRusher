package context

import (
	"github.com/free5gc/util/fsm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestForceReleaseAllPDUSessionsAllowsCallbackRemoval(t *testing.T) {
	machine, err := initPduFSM(nil)
	require.NoError(t, err)
	ue := &UEContext{smContexts: make(map[int32]*SmContext), pduFsm: machine, securityContext: &SecurityContext{}}
	for _, id := range []int32{1, 2} {
		session := NewSmContext(id)
		session.state = fsm.NewState(Active)
		require.NoError(t, ue.AddSmContext(session))
	}
	done := make(chan struct{})
	go func() { ForceReleaseAllPDUSession(ue); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("session release callbacks deadlocked")
	}
	for _, id := range []int32{1, 2} {
		_, err := ue.GetSmContext(id)
		require.Error(t, err)
	}
}

func TestConfirmReleasePreservesOtherPDUSessions(t *testing.T) {
	machine, err := initPduFSM(nil)
	require.NoError(t, err)
	ue := &UEContext{smContexts: make(map[int32]*SmContext), pduFsm: machine}
	first, second := NewSmContext(1), NewSmContext(2)
	first.state = fsm.NewState(InactivePending)
	second.state = fsm.NewState(Active)
	require.NoError(t, ue.AddSmContext(first))
	require.NoError(t, ue.AddSmContext(second))
	require.NoError(t, ConfirmPDUSessionRelease(ue, 1))
	_, err = ue.GetSmContext(1)
	require.Error(t, err)
	remaining, err := ue.GetSmContext(2)
	require.NoError(t, err)
	require.Same(t, second, remaining)
}

// Switch-off deregistration may overtake a release ACK on the same UE. Retire
// pending sessions coherently and accept a later ACK only for a known retirement.
func TestForceReleasePendingSessionAndLateConfirmation(t *testing.T) {
	machine, err := initPduFSM(nil)
	require.NoError(t, err)
	ue := &UEContext{smContexts: make(map[int32]*SmContext), pduFsm: machine, securityContext: &SecurityContext{}}
	session := NewSmContext(1)
	session.state.Set(InactivePending)
	require.NoError(t, ue.AddSmContext(session))
	ForceReleaseAllPDUSession(ue)
	require.Equal(t, Inactive, session.state.Current())
	require.NoError(t, ConfirmPDUSessionRelease(ue, 1))
	require.Error(t, ConfirmPDUSessionRelease(ue, 2), "unknown session must remain rejected")
	replacement := NewSmContext(1)
	replacement.state.Set(Active)
	require.NoError(t, ue.AddSmContext(replacement))
	require.Error(t, ConfirmPDUSessionRelease(ue, 1), "stale confirmation must not release a new active session")
	remaining, err := ue.GetSmContext(1)
	require.NoError(t, err)
	require.Same(t, replacement, remaining)
}
