/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package templates

import (
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMember stands in for SimulateSingleUE's loop for a cohort member, without a
// network. It reports the same states, and like the real loop it keeps reading its
// scenario channel while registering and while leaving, dropping what it cannot act on.
// That window, a Terminate arriving while the member is leaving, is where a driver
// working from its own record of the members fell out of step with them.
type fakeMember struct {
	m          *tools.ChurnMember
	scenario   chan procedures.UeTesterMessage
	register   time.Duration // time to register; 0 never registers
	deregister time.Duration

	registrations atomic.Int32
	terminates    atomic.Int32 // Terminates acted on
	dropped       atomic.Int32 // messages read while leaving, and dropped
}

func startFake(t *testing.T, m *tools.ChurnMember, register, deregister time.Duration) *fakeMember {
	t.Helper()
	f := &fakeMember{m: m, scenario: make(chan procedures.UeTesterMessage), register: register, deregister: deregister}
	go f.run()
	t.Cleanup(func() { sendUnlessDone(f.scenario, procedures.UeTesterMessage{Type: procedures.Terminate}, m) })
	return f
}

func (f *fakeMember) run() {
	defer f.m.Exit()
	for {
		f.m.Report(tools.ChurnStarting)
		var registered <-chan time.Time
		if f.register > 0 {
			registered = time.After(f.register)
		}
	serving:
		for {
			select {
			case <-registered:
				f.m.Report(tools.ChurnRegistered)
				f.registrations.Add(1)
			case msg := <-f.scenario:
				if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
					f.terminates.Add(1)
					break serving
				}
			}
		}
		leaving := time.After(f.deregister)
	left:
		for {
			select {
			case <-leaving:
				break left
			case <-f.scenario:
				f.dropped.Add(1)
			}
		}
		if !f.m.Park(f.scenario) {
			return
		}
	}
}

func terminateVia(fakes map[int]*fakeMember, sent *[]int, mu *sync.Mutex) func(int) {
	return func(id int) {
		mu.Lock()
		*sent = append(*sent, id)
		mu.Unlock()
		f := fakes[id]
		sendUnlessDone(f.scenario, procedures.UeTesterMessage{Type: procedures.Terminate}, f.m)
	}
}

func newTestCohort(size int) *churnCohort {
	c := newChurnCohort(size, size, 0)
	c.registerWait = 5 * time.Second
	c.leaveWait = 5 * time.Second
	c.poll = time.Millisecond
	return c
}

func TestNewChurnCohort(t *testing.T) {
	assert.Nil(t, newChurnCohort(0, 10, 0), "0 disables churn")
	assert.Nil(t, newChurnCohort(-1, 10, 0), "a negative size disables churn")

	c := newChurnCohort(25, 10, 0)
	require.NotNil(t, c)
	assert.Equal(t, 10, c.size, "a size above the number of UEs churns them all")

	c = newChurnCohort(3, 10, 0)
	for id := 1; id <= 3; id++ {
		assert.NotNil(t, c.memberFor(id), "UE %d is in the cohort", id)
	}
	assert.Nil(t, c.memberFor(4), "UE 4 is outside the cohort")
	assert.Nil(t, c.memberFor(0))

	var none *churnCohort
	assert.Nil(t, none.memberFor(1), "no cohort, no member")
}

// Waves signalled back to back: out, in, out. Each wave acts on where the member really
// is, so it ends parked, no Terminate is lost in its leaving window, and the next wave
// in and out still work.
func TestChurnBackToBackWavesStayInStep(t *testing.T) {
	c := newTestCohort(1)
	f := startFake(t, c.memberFor(1), 20*time.Millisecond, 20*time.Millisecond)
	var sent []int
	var mu sync.Mutex
	terminate := terminateVia(map[int]*fakeMember{1: f}, &sent, &mu)

	require.False(t, c.waveOut(terminate, nil))
	require.False(t, c.waveIn(nil))
	require.False(t, c.waveOut(terminate, nil))

	assert.Equal(t, tools.ChurnParked, f.m.State(), "after out, in, out the member is out")
	assert.EqualValues(t, 2, f.registrations.Load(), "it registered once at the start and once on the wave in")
	assert.EqualValues(t, 2, f.terminates.Load())
	assert.Zero(t, f.dropped.Load(), "no Terminate was sent into its leaving window")

	require.False(t, c.waveIn(nil))
	require.False(t, c.waveOut(terminate, nil))
	assert.Equal(t, tools.ChurnParked, f.m.State())
	assert.EqualValues(t, 3, f.registrations.Load())
	assert.EqualValues(t, 3, f.terminates.Load())
}

// The first wave can come while members are still registering. A UE told to terminate
// then skips its deregistration, so the wave waits for each member to register.
func TestChurnWaveOutWaitsForRegistration(t *testing.T) {
	c := newTestCohort(2)
	fakes := map[int]*fakeMember{}
	for id := 1; id <= 2; id++ {
		fakes[id] = startFake(t, c.memberFor(id), 200*time.Millisecond, time.Millisecond)
	}

	var early atomic.Int32
	terminate := func(id int) {
		if fakes[id].m.State() != tools.ChurnRegistered {
			early.Add(1)
		}
		sendUnlessDone(fakes[id].scenario, procedures.UeTesterMessage{Type: procedures.Terminate}, fakes[id].m)
	}

	require.False(t, c.waveOut(terminate, nil))
	assert.Zero(t, early.Load(), "no member was sent Terminate before it had registered")
	for id := 1; id <= 2; id++ {
		assert.Equal(t, tools.ChurnParked, fakes[id].m.State(), "UE %d", id)
	}
}

// A member that does not register in time is left in rather than terminated, and the
// wave does not wait for it beyond registerWait.
func TestChurnWaveOutLeavesInAMemberThatDoesNotRegister(t *testing.T) {
	c := newTestCohort(1)
	c.registerWait = 50 * time.Millisecond
	f := startFake(t, c.memberFor(1), 0, time.Millisecond)
	var sent []int
	var mu sync.Mutex

	start := time.Now()
	require.False(t, c.waveOut(terminateVia(map[int]*fakeMember{1: f}, &sent, &mu), nil))
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Empty(t, sent)
	assert.Equal(t, tools.ChurnStarting, f.m.State())
}

// Members that never register share one wait: k of them cost a wave out one
// registerWait, not k. A registered member after them is still sent out.
func TestChurnWaveOutSharesOneRegisterWait(t *testing.T) {
	c := newTestCohort(4)
	c.registerWait = 200 * time.Millisecond
	fakes := map[int]*fakeMember{}
	for id := 1; id <= 3; id++ {
		fakes[id] = startFake(t, c.memberFor(id), 0, time.Millisecond)
	}
	fakes[4] = startFake(t, c.memberFor(4), time.Millisecond, time.Millisecond)
	require.Eventually(t, func() bool { return fakes[4].m.State() == tools.ChurnRegistered }, 2*time.Second, time.Millisecond)
	var sent []int
	var mu sync.Mutex

	start := time.Now()
	require.False(t, c.waveOut(terminateVia(fakes, &sent, &mu), nil))
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 2*c.registerWait, "three members that never register should cost one wait, not three")
	assert.Equal(t, []int{4}, sent, "the registered member is still sent out")
	for id := 1; id <= 3; id++ {
		assert.Equal(t, tools.ChurnStarting, fakes[id].m.State(), "UE %d is left in", id)
	}
}

// Out, in and out again: every member is sent Terminate on both waves out.
func TestChurnCycleTerminatesEveryMemberOnEachWaveOut(t *testing.T) {
	c := newTestCohort(3)
	fakes := map[int]*fakeMember{}
	for id := 1; id <= 3; id++ {
		fakes[id] = startFake(t, c.memberFor(id), 5*time.Millisecond, 5*time.Millisecond)
	}
	var sent []int
	var mu sync.Mutex
	terminate := terminateVia(fakes, &sent, &mu)

	require.False(t, c.waveOut(terminate, nil))
	require.False(t, c.waveIn(nil))
	require.False(t, c.waveOut(terminate, nil))

	assert.Equal(t, []int{1, 2, 3, 1, 2, 3}, sent)
	for id := 1; id <= 3; id++ {
		assert.EqualValues(t, 2, fakes[id].registrations.Load(), "UE %d registered again on the wave in", id)
	}
}

// A wave in before any wave out has nothing to rearm, and a second wave out finds
// every member already out.
func TestChurnWavesActOnlyOnMembersInTheRightState(t *testing.T) {
	c := newTestCohort(2)
	fakes := map[int]*fakeMember{}
	for id := 1; id <= 2; id++ {
		fakes[id] = startFake(t, c.memberFor(id), 5*time.Millisecond, 5*time.Millisecond)
	}
	var sent []int
	var mu sync.Mutex
	terminate := terminateVia(fakes, &sent, &mu)

	require.Eventually(t, func() bool {
		return fakes[1].m.State() == tools.ChurnRegistered && fakes[2].m.State() == tools.ChurnRegistered
	}, 2*time.Second, time.Millisecond)
	require.False(t, c.waveIn(nil))
	for id := 1; id <= 2; id++ {
		assert.EqualValues(t, 1, fakes[id].registrations.Load(), "UE %d was in, so it is not rearmed", id)
	}

	require.False(t, c.waveOut(terminate, nil))
	require.False(t, c.waveOut(terminate, nil))
	assert.Equal(t, []int{1, 2}, sent, "the second wave out has no one to send out")
}

// A member that has returned blocks neither wave nor the end-of-run Terminate.
func TestChurnMemberGoneBlocksNothing(t *testing.T) {
	c := newTestCohort(1)
	m := c.memberFor(1)
	m.Exit()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.waveOut(func(int) { t.Error("a member that has returned was sent Terminate") }, nil)
		c.waveIn(nil)
		sendUnlessDone(make(chan procedures.UeTesterMessage), procedures.UeTesterMessage{Type: procedures.Terminate}, m)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("blocked on a member that has returned")
	}
}

// SIGINT during a wave stops it at once instead of after the rest of the cohort.
func TestChurnWaveStopsOnSignal(t *testing.T) {
	c := newTestCohort(2)
	fakes := map[int]*fakeMember{}
	for id := 1; id <= 2; id++ {
		fakes[id] = startFake(t, c.memberFor(id), time.Millisecond, time.Millisecond)
	}
	require.Eventually(t, func() bool {
		return fakes[1].m.State() == tools.ChurnRegistered && fakes[2].m.State() == tools.ChurnRegistered
	}, 2*time.Second, time.Millisecond)
	var sent []int
	var mu sync.Mutex

	stop := make(chan os.Signal, 1)
	stop <- syscall.SIGINT
	assert.True(t, c.waveOut(terminateVia(fakes, &sent, &mu), stop))
	assert.Empty(t, sent)

	// Out, then a wave in that is stopped before it rearms anyone.
	require.False(t, c.waveOut(terminateVia(fakes, &sent, &mu), nil))
	stopIn := make(chan os.Signal, 1)
	stopIn <- syscall.SIGINT
	assert.True(t, c.waveIn(stopIn))
	for id := 1; id <= 2; id++ {
		assert.Equal(t, tools.ChurnParked, fakes[id].m.State(), "UE %d was not rearmed", id)
	}
}
