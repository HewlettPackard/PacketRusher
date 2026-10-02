/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package tools

import (
	"my5G-RANTester/internal/control_test_engine/procedures"
	"sync/atomic"
)

// ChurnState is where a churn cohort member is, as the member itself reports it.
type ChurnState int32

const (
	// ChurnStarting: registering, or about to.
	ChurnStarting ChurnState = iota
	// ChurnRegistered: NAS registration has completed. Its PDU sessions may still be
	// being established.
	ChurnRegistered
	// ChurnParked: its UE has ended and the member is waiting to be rearmed. The UE
	// deregisters with switch-off, so this does not mean the AMF has finished.
	ChurnParked
	// ChurnGone: its goroutine has returned.
	ChurnGone
)

// ChurnMember connects one churn cohort member's goroutine with the wave driver. The
// member reports its own state, and the driver acts on that state rather than on its
// own record of what it last asked for: a request that crossed a state change is then
// seen as not having happened, instead of leaving the two out of step.
type ChurnMember struct {
	rearm chan struct{} // unbuffered: a rearm is taken only by a parked member, or not at all
	done  chan struct{} // closed when the member's goroutine returns
	state atomic.Int32
}

func NewChurnMember() *ChurnMember {
	return &ChurnMember{rearm: make(chan struct{}), done: make(chan struct{})}
}

// State is the member's state as it last reported it.
func (m *ChurnMember) State() ChurnState {
	return ChurnState(m.state.Load())
}

// Done is closed once the member's goroutine has returned.
func (m *ChurnMember) Done() <-chan struct{} {
	return m.done
}

// Rearm brings a parked member back. It returns false, without blocking, if the member
// has returned. Call it only for a member whose State is ChurnParked: a parked member
// leaves that state only through Rearm or through a scenario message, so it is in Park
// to take the rearm. By the time Rearm returns true the member no longer reports
// ChurnParked, so a caller that looks again sees it starting.
func (m *ChurnMember) Rearm() bool {
	select {
	case m.rearm <- struct{}{}:
		m.unpark()
		return true
	case <-m.done:
		return false
	}
}

// unpark moves a member that took a rearm from ChurnParked to ChurnStarting. Both the
// driver and the member do it, and whichever is first wins: neither may overwrite a
// state the member has reported since, such as ChurnRegistered.
func (m *ChurnMember) unpark() {
	m.state.CompareAndSwap(int32(ChurnParked), int32(ChurnStarting))
}

// The member side: SimulateSingleUE reports the member's state, parks it and marks it
// gone. They are exported so a test can stand in for a member.

// Report records the member's state.
func (m *ChurnMember) Report(s ChurnState) {
	m.state.Store(int32(s))
}

// Park holds a deregistered member until it is rearmed. It returns false if the member
// should stop instead: on Terminate or Kill, which end the run, or when its scenario
// channel is closed. Any other message leaves it parked.
func (m *ChurnMember) Park(scenarioChan <-chan procedures.UeTesterMessage) bool {
	m.Report(ChurnParked)
	for {
		select {
		case <-m.rearm:
			m.unpark()
			return true
		case msg, ok := <-scenarioChan:
			if !ok || msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
				return false
			}
		}
	}
}

// Exit marks the member gone; call it as its goroutine returns.
func (m *ChurnMember) Exit() {
	m.Report(ChurnGone)
	close(m.done)
}
