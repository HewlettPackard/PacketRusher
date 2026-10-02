/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package templates

import (
	"fmt"
	"my5G-RANTester/internal/common/tools"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// churnCohort drives a churn cohort, UE 1..size: on each wave out it deregisters the
// members, and on each wave in it rearms them so they register again.
//
// Waves are driven by the caller rather than by a timer. A per-UE timer runs from that
// UE's own registration, so a cohort's deregistrations smear across the whole
// registration spread, and a measurement taken between waves has no clean interval to
// belong to.
//
// The driver keeps no state of its own about the members. It acts on what each member
// reports, so a wave signalled before the previous one has settled acts on where the
// members really are.
type churnCohort struct {
	size    int
	gap     time.Duration        // between UEs within a wave
	members []*tools.ChurnMember // by UE id; nil outside the cohort

	// registerWait bounds how long a wave out waits for a member to register before
	// it leaves that member in; leaveWait, how long it waits for the members it sent
	// out to deregister.
	registerWait time.Duration
	leaveWait    time.Duration
	poll         time.Duration
}

// newChurnCohort returns the cohort of the first size of numUes UEs, or nil when size
// is 0. A size above numUes churns every UE.
func newChurnCohort(size, numUes int, gap time.Duration) *churnCohort {
	if size > numUes {
		log.Warn("[TESTER] --churnFirstN exceeds the number of UEs; churning all ", numUes)
		size = numUes
	}
	if size <= 0 {
		return nil
	}

	c := &churnCohort{
		size:         size,
		gap:          gap,
		members:      make([]*tools.ChurnMember, size+1),
		registerWait: 30 * time.Second,
		leaveWait:    30 * time.Second,
		poll:         10 * time.Millisecond,
	}
	for id := 1; id <= size; id++ {
		c.members[id] = tools.NewChurnMember()
	}

	return c
}

// memberFor is the cohort member UE id is, or nil if it is not in the cohort.
func (c *churnCohort) memberFor(id int) *tools.ChurnMember {
	if c == nil || id < 1 || id > c.size {
		return nil
	}

	return c.members[id]
}

// waveOut deregisters, through terminate, every member that is registered. Members
// still registering are waited for: a UE told to terminate before it is registered
// skips its deregistration, and the AMF keeps its context. The waiting is shared by
// the whole wave, at most registerWait in all, so members that never register cost
// the wave one wait rather than one each; any still registering after it are left in.
// Members already parked are left alone. The wave then waits, up to leaveWait, for the
// members it sent out to park. It returns true if stop fired, leaving the rest of the
// wave undone.
func (c *churnCohort) waveOut(terminate func(id int), stop <-chan os.Signal) bool {
	log.Info("[TESTER][CHURN] wave out: deregistering up to ", c.size, " UEs")
	var sent, leftIn []int
	budget := c.registerWait
	for id := 1; id <= c.size; id++ {
		if stopped(stop) {
			return true
		}
		m := c.members[id]
		waitStart := time.Now()
		state, stopped := c.await(m, budget, stop, tools.ChurnRegistered, tools.ChurnParked)
		budget -= time.Since(waitStart)
		if stopped {
			return true
		}
		switch state {
		case tools.ChurnRegistered:
			terminate(id)
			sent = append(sent, id)
			time.Sleep(c.gap)
		case tools.ChurnStarting:
			leftIn = append(leftIn, id)
		}
	}
	if len(leftIn) > 0 {
		log.Warn("[TESTER][CHURN] wave out: ", len(leftIn), " UEs had not registered within ", c.registerWait,
			" and were left in: ", firstIds(leftIn))
	}

	deadline := time.Now().Add(c.leaveWait)
	var notParked []int
	for _, id := range sent {
		state, stopped := c.await(c.members[id], time.Until(deadline), stop, tools.ChurnParked)
		if stopped {
			return true
		}
		if state != tools.ChurnParked && state != tools.ChurnGone {
			notParked = append(notParked, id)
		}
	}
	if len(notParked) > 0 {
		log.Warn("[TESTER][CHURN] wave out: ", len(notParked), " of ", len(sent), " UEs had not deregistered within ",
			c.leaveWait, ": ", firstIds(notParked))
	}
	log.Info("[TESTER][CHURN] wave out: complete, ", len(sent), " UEs sent out")

	return false
}

// waveIn rearms every parked member. Registered members are in already. Members still
// registering are left as they are, and reported, since nothing else would show them.
// A member still deregistering from a wave out that stopped waiting for it reads as
// registered until it parks, and is left for the next wave. It returns true if stop
// fired, leaving the rest of the wave undone.
func (c *churnCohort) waveIn(stop <-chan os.Signal) bool {
	log.Info("[TESTER][CHURN] wave in: re-registering up to ", c.size, " UEs")
	rearmed := 0
	var registering []int
	for id := 1; id <= c.size; id++ {
		if stopped(stop) {
			return true
		}
		m := c.members[id]
		switch m.State() {
		case tools.ChurnParked:
			if m.Rearm() {
				rearmed++
				time.Sleep(c.gap)
			}
		case tools.ChurnStarting:
			registering = append(registering, id)
		}
	}
	if len(registering) > 0 {
		log.Warn("[TESTER][CHURN] wave in: ", len(registering), " UEs were still registering and were left as they are: ", firstIds(registering))
	}
	log.Info("[TESTER][CHURN] wave in: complete, ", rearmed, " UEs rearmed")

	return false
}

// firstIds formats up to ten of ids for a log line.
func firstIds(ids []int) string {
	if len(ids) <= 10 {
		return fmt.Sprint(ids)
	}
	return fmt.Sprint(ids[:10], " and ", len(ids)-10, " more")
}

// await waits up to timeout for m to report one of want, or to be gone, and returns the
// state it found. It returns stopped if stop fired first.
func (c *churnCohort) await(m *tools.ChurnMember, timeout time.Duration, stop <-chan os.Signal, want ...tools.ChurnState) (state tools.ChurnState, stopped bool) {
	deadline := time.Now().Add(timeout)
	for {
		state = m.State()
		if state == tools.ChurnGone {
			return state, false
		}
		for _, w := range want {
			if state == w {
				return state, false
			}
		}
		if !time.Now().Before(deadline) {
			return state, false
		}
		select {
		case <-stop:
			return state, true
		case <-time.After(c.poll):
		}
	}
}

func stopped(stop <-chan os.Signal) bool {
	select {
	case <-stop:
		return true
	default:
		return false
	}
}
