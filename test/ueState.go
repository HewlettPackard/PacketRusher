/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"errors"
	"sync"
	"testing"

	"my5G-RANTester/lib/fsm"

	"github.com/stretchr/testify/assert"
)

type UEStateList struct {
	m        sync.Mutex
	ueStates []*UEState
}

func (u *UEStateList) Add(single *UEState) {
	if u.ueStates == nil {
		u.ueStates = make([]*UEState, 0)
	}
	u.ueStates = append(u.ueStates, single)
}

func (u *UEStateList) FindByMsin(msin string) (*UEState, error) {
	u.m.Lock()
	defer u.m.Unlock()
	for i := range u.ueStates {
		if u.ueStates[i].msin == msin {
			return u.ueStates[i], nil
		}
	}
	return nil, errors.New("[TEST] UE with msin " + msin + "not found")
}

func (u *UEStateList) CheckPDU() {
	for i := range u.ueStates {
		u.ueStates[i].CheckPDU()
	}
}

type UEState struct {
	m           sync.Mutex
	t           *testing.T
	msin        string
	FSM         fsm.FSM
	pduSessions *pduSessionState
}

func (s *UEState) Init(t *testing.T, msin string, fsm fsm.FSM, checkPDU bool) {
	s.msin = msin
	s.FSM = fsm
	sessionsS := pduSessionState{}
	s.pduSessions = &sessionsS
	s.t = t
}

func (s *UEState) NewPDURequest() {
	s.pduSessions.addRequest()
}

func (s *UEState) NewPDUProvided() {
	s.pduSessions.addProvided()
}

func (s *UEState) NewPDUReleased() {
	s.pduSessions.addReleased()
}

func (s *UEState) CheckPDU() {
	s.pduSessions.check(s.t, s.msin)
}

type State int

const (
	Fresh State = iota + 1
	RegistrationRequested
	AuthenticationChallengeResponded
	SecurityContextSet
	Registred
	Idle
)

func (s State) String() string {
	switch s {
	case Fresh:
		return "Fresh"
	case RegistrationRequested:
		return "RegistrationRequested"
	case AuthenticationChallengeResponded:
		return "AuthenticationChallengeResponded"
	case SecurityContextSet:
		return "SecurityContextSet"
	case Registred:
		return "Registred"
	case Idle:
		return "Deregistered"
	}
	return "Unknown"
}

type pduSessionState struct {
	m         sync.Mutex
	requested int
	provided  int
	released  int
}

func (s *pduSessionState) addReleased() {
	s.m.Lock()
	defer s.m.Unlock()
	s.released++
}

func (s *pduSessionState) addProvided() {
	s.m.Lock()
	defer s.m.Unlock()
	s.provided++
}

func (s *pduSessionState) addRequest() {
	s.m.Lock()
	defer s.m.Unlock()
	s.requested++
}

func (s *pduSessionState) check(t assert.TestingT, msin string) {
	assert.Equalf(t, s.requested, s.provided, "[UE-%s] %d PDU sessions requested but %d were provided", msin, s.requested, s.provided)
	assert.Equalf(t, s.provided, s.released, "[UE-%s] %d PDU sessions provided but %d were released", msin, s.provided, s.released)
}
