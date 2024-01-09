/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package state

import (
	"sync"

	"github.com/free5gc/util/fsm"
)

type State int

const (
	Fresh                            fsm.StateType = "Fresh"
	RegistrationRequested            fsm.StateType = "RegistrationRequested"
	AuthenticationChallengeResponded fsm.StateType = "AuthenticationChallengeResponded"
	SecurityContextSet               fsm.StateType = "SecurityContextSet"
	Registred                        fsm.StateType = "Registred"
	DeregistrationRequested          fsm.StateType = "DeregistrationRequested"
	Idle                             fsm.StateType = "Idle"
)

type UEState struct {
	m           sync.Mutex
	state       *fsm.State
	pduSessions *pduSessionCount
}

func (s *UEState) Init() {
	s.state = fsm.NewState(Fresh)
	sessions := pduSessionCount{}
	s.pduSessions = &sessions
}

func (s *UEState) GetState() *fsm.State {
	return s.state
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

func (s *UEState) GetRequestedPDUCount() int {
	return s.pduSessions.getRequested()
}

func (s *UEState) GetProvidedPDUCount() int {
	return s.pduSessions.getProvided()
}

func (s *UEState) GetReleasedPDUCount() int {
	return s.pduSessions.getReleased()
}

type pduSessionCount struct {
	m         sync.Mutex
	requested int
	provided  int
	released  int
}

func (s *pduSessionCount) addReleased() {
	s.m.Lock()
	defer s.m.Unlock()
	s.released++
}

func (s *pduSessionCount) addProvided() {
	s.m.Lock()
	defer s.m.Unlock()
	s.provided++
}

func (s *pduSessionCount) addRequest() {
	s.m.Lock()
	defer s.m.Unlock()
	s.requested++
}

func (s *pduSessionCount) getRequested() int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.requested
}

func (s *pduSessionCount) getProvided() int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.provided
}

func (s *pduSessionCount) getReleased() int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.released
}
