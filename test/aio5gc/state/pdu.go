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
	ForceRelease         fsm.EventType = "ForceRelease"
	ModificationCommand  fsm.EventType = "ModificationCommand"
	ModificationComplete fsm.EventType = "ModificationComplete"
)

var pduLock = &sync.Mutex{}
var PduFsm *fsm.FSM

func InitPduFSM(callbacks fsm.Callbacks) {

	var err error
	states := []fsm.StateType{Inactive, InactivePending, Active, ModificationPending}
	for _, i := range states {
		_, ok := callbacks[i]
		if !ok {
			callbacks[i] = func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}
		}
	}

	pduLock.Lock()
	defer pduLock.Unlock()
	PduFsm, err = fsm.NewFSM(
		fsm.Transitions{
			{Event: EstablishmentReject, From: Inactive, To: Inactive},
			{Event: EstablishmentAccept, From: Inactive, To: Active},
			{Event: ReleaseComplete, From: InactivePending, To: Inactive},
			{Event: ReleaseCommand, From: Active, To: InactivePending},
			{Event: ModificationCommand, From: Active, To: ModificationPending},
			{Event: ModificationComplete, From: ModificationPending, To: Inactive},
			{Event: ForceRelease, From: Active, To: Inactive},
		},
		fsm.Callbacks{
			Inactive:            callbacks[Inactive],
			InactivePending:     callbacks[InactivePending],
			Active:              callbacks[Active],
			ModificationPending: callbacks[ModificationPending],
		},
	)
	if err != nil {
		log.Fatal("[5GC] Failed to create PDU FSM: " + err.Error())
	}
}

func GetPduFSM() *fsm.FSM {
	if PduFsm == nil {

	}

	return PduFsm
}
