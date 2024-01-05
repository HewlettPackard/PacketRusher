/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"errors"
	"fmt"
	"sync"

	"my5G-RANTester/lib/fsm"
)

type UEStateList struct {
	m        sync.Mutex
	Fsm      *fsm.FSM
	UeStates map[int64]*UEState
}

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

func (u *UEStateList) New(amfId int64) (*UEState, error) {
	if u.UeStates == nil {
		u.UeStates = make(map[int64]*UEState, 0)
	}
	_, exist := u.UeStates[amfId]
	if exist {
		return nil, fmt.Errorf("[TEST] UE with AMFID %d already assigned", amfId)
	}
	u.UeStates[amfId] = &UEState{
		state:       fsm.NewState(Fresh),
		pduSessions: &pduSessionCount{},
	}
	return u.UeStates[amfId], nil
}

func (u *UEStateList) FindByAmfId(amfId int64) (*UEState, error) {
	u.m.Lock()
	defer u.m.Unlock()
	ue, exist := u.UeStates[amfId]
	if exist {
		return ue, nil
	}
	return nil, fmt.Errorf("[TEST] UE with AMFID %d not found", amfId)
}

func (u *UEStateList) FindByMsin(msin string) (*UEState, error) {
	u.m.Lock()
	defer u.m.Unlock()
	for i := range u.UeStates {
		if u.UeStates[i].msin == msin {
			return u.UeStates[i], nil
		}
	}
	return nil, errors.New("[TEST] UE with msin " + msin + "not found")
}

type UEState struct {
	m           sync.Mutex
	msin        string
	state       *fsm.State
	pduSessions *pduSessionCount
}

func (s *UEState) Init(msin string, state *fsm.State, checkPDU bool) {
	s.state = state
	sessions := pduSessionCount{}
	s.pduSessions = &sessions
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
