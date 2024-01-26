package state

import (
	"sync"

	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

const (
	AuthenticationInitiated fsm.StateType = "AuthenticationInitiated"
	Authenticated           fsm.StateType = "Authenticated"
	Registred               fsm.StateType = "Registred"
	DeregistratedInitiated  fsm.StateType = "DeregistratedInitiated"
	Deregistrated           fsm.StateType = "Deregistrated"
)

const (
	RegistrationRequest   fsm.EventType = "InitialRegistrationRequested"
	AuthenticationSuccess fsm.EventType = "InitialRegistrationAccepted"
	RegistrationAccept    fsm.EventType = "InitialRegistrationAccepted"
	DeregistrationRequest fsm.EventType = "DeregistrationRequested"
	Deregistration        fsm.EventType = "Deregistration"
)

var ueLock = &sync.Mutex{}
var UeFsm *fsm.FSM

func InitUEFSM(callbacks fsm.Callbacks) {

	var err error
	states := []fsm.StateType{AuthenticationInitiated, Authenticated, Registred, DeregistratedInitiated, Deregistrated}
	for _, i := range states {
		_, ok := callbacks[i]
		if !ok {
			callbacks[i] = func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}
		}
	}

	ueLock.Lock()
	defer ueLock.Unlock()
	UeFsm, err = fsm.NewFSM(
		fsm.Transitions{
			{Event: RegistrationRequest, From: Deregistrated, To: AuthenticationInitiated},
			{Event: AuthenticationSuccess, From: AuthenticationInitiated, To: Authenticated},
			{Event: RegistrationAccept, From: Authenticated, To: Registred},
			{Event: DeregistrationRequest, From: Registred, To: DeregistratedInitiated},
			{Event: Deregistration, From: DeregistratedInitiated, To: Deregistrated},
		},
		fsm.Callbacks{
			AuthenticationInitiated: callbacks[AuthenticationInitiated],
			Authenticated:           callbacks[Authenticated],
			Registred:               callbacks[Registred],
			DeregistratedInitiated:  callbacks[DeregistratedInitiated],
			Deregistrated:           callbacks[Deregistrated],
		},
	)
	if err != nil {
		log.Fatal("[5GC] Failed to create UE FSM: " + err.Error())
	}
}

func GetUeFsm() *fsm.FSM {
	if UeFsm == nil {
		InitUEFSM(fsm.Callbacks{})
	}
	return UeFsm
}
