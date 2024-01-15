package state

import (
	"sync"

	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

const (
	Inactive            fsm.StateType = "Inactive"
	InactivePending     fsm.StateType = "InactivePending"
	Active              fsm.StateType = "Active"
	ModificationPending fsm.StateType = "ModificationPending"
)

const (
	EstablishmentReject  fsm.EventType = "EstablishmentReject"
	EstablishmentAccept  fsm.EventType = "EstablishmentAccept"
	ReleaseComplete      fsm.EventType = "ReleaseComplete"
	ReleaseCommand       fsm.EventType = "ReleaseCommand"
	ModificationCommand  fsm.EventType = "ModificationCommand"
	ModificationComplete fsm.EventType = "ModificationComplete"
)

var pduLock = &sync.Mutex{}
var PduFsm *fsm.FSM

func GetPduFSM() *fsm.FSM {
	if UeFsm == nil {
		pduLock.Lock()
		defer pduLock.Unlock()
		var err error
		UeFsm, err = fsm.NewFSM(
			fsm.Transitions{
				{Event: EstablishmentReject, From: Inactive, To: Inactive},
				{Event: EstablishmentAccept, From: Inactive, To: Active},
				{Event: ReleaseComplete, From: InactivePending, To: Inactive},
				{Event: ReleaseCommand, From: Active, To: Inactive},
				{Event: ModificationCommand, From: Active, To: ModificationPending},
				{Event: ModificationComplete, From: ModificationPending, To: Inactive},
			},
			fsm.Callbacks{
				Inactive:            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				InactivePending:     func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				Active:              func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
				ModificationPending: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}},
		)
		if err != nil {
			log.Fatal("[5GC] Failed to create PDU FSM: " + err.Error())
		}
	}

	return UeFsm
}
