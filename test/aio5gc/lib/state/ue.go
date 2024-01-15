package state

import (
	"sync"

	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

const (
	Deregistrated            fsm.StateType = "Deregistrated"
	CommonProcedureInitiated fsm.StateType = "CommonProcedureInitiated"
	Registred                fsm.StateType = "Registred"
	DeregistratedInitiated   fsm.StateType = "DeregistratedInitiated"
)

const (
	InitialRegistrationRequested fsm.EventType = "InitialRegistrationRequested"
	InitialRegistrationAccepted  fsm.EventType = "InitialRegistrationAccepted"
	DeregistrationRequested      fsm.EventType = "DeregistrationRequested"
	Deregistration               fsm.EventType = "Deregistration"
)

var ueLock = &sync.Mutex{}
var UeFsm *fsm.FSM

func GetUeFsm() *fsm.FSM {
	if UeFsm == nil {
		ueLock.Lock()
		defer ueLock.Unlock()
		var err error
		UeFsm, err = fsm.NewFSM(
			fsm.Transitions{
				{Event: InitialRegistrationRequested, From: Deregistrated, To: CommonProcedureInitiated},
				{Event: InitialRegistrationAccepted, From: CommonProcedureInitiated, To: Registred},
				{Event: DeregistrationRequested, From: Registred, To: DeregistratedInitiated},
				{Event: Deregistration, From: DeregistratedInitiated, To: Deregistrated},
				{Event: Deregistration, From: Registred, To: Deregistrated},
			},
			fsm.Callbacks{
				Deregistrated:            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				CommonProcedureInitiated: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				Registred:                func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				DeregistratedInitiated:   func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}},
		)
		if err != nil {
			log.Fatal("[5GC] Failed to create UE FSM: " + err.Error())
		}
	}

	return UeFsm
}
