/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"fmt"
	"my5G-RANTester/config"
	"strconv"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
)

// All in one 5GC for test purpose
type Aio5gc struct {
	amfContext AMFContext
	session    SessionContext
	nasHooks   map[uint8]func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error)
	ngapHook   []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error)
	conf       config.Config
}

func (a *Aio5gc) GetAMFContext() *AMFContext {
	return &a.amfContext
}

func (a *Aio5gc) GetSessionContext() *SessionContext {
	return &a.session
}

func (a *Aio5gc) Init(conf config.Config, id string, name string, ueCallbacks map[fsm.StateType]fsm.Callback, pduCallbacks map[fsm.StateType]fsm.Callback) error {
	plmn := models.PlmnId{
		Mcc: conf.GNodeB.PlmnList.Mcc,
		Mnc: conf.GNodeB.PlmnList.Mnc,
	}
	i, err := strconv.ParseInt(conf.GNodeB.SliceSupportList.Sst, 10, 32)
	if err != nil {
		err = errors.New("failed to convert config sst to int: " + err.Error())
		return err
	}
	sst := int32(i)
	supportedPlmns := []models.PlmnSnssai{
		{
			PlmnId: &plmn,
			SNssaiList: []models.Snssai{
				{
					Sst: sst,
					Sd:  conf.GNodeB.SliceSupportList.Sd,
				},
			},
		}}
	servedGuami := []models.Guami{
		{
			PlmnId: &plmn,
			AmfId:  id,
		},
	}

	pdufsm, err := initPduFSM(pduCallbacks)
	uefsm, err := initUeFSM(ueCallbacks)

	a.conf = conf
	a.amfContext = AMFContext{}
	a.amfContext.NewAmfContext(
		name,
		id,
		supportedPlmns,
		servedGuami,
		100,
		uefsm,
		pdufsm,
	)

	a.session.NewSessionContext()
	return nil
}

func (a *Aio5gc) GetNasHook(msgType uint8) func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error) {
	return a.nasHooks[msgType]
}

func (a *Aio5gc) SetNasHooks(hooks map[uint8]func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error)) {
	a.nasHooks = hooks
}

func (a *Aio5gc) GetNgapHooks() []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error) {
	return a.ngapHook
}

func (a *Aio5gc) SetNgapHooks(hook []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error)) {
	a.ngapHook = hook
}

func initUeFSM(callbacks fsm.Callbacks) (*fsm.FSM, error) {

	if callbacks == nil {
		callbacks = fsm.Callbacks{}
	}
	states := []fsm.StateType{AuthenticationInitiated, Authenticated, Registered, DeregisteredInitiated, Deregistered}
	for _, i := range states {
		_, ok := callbacks[i]
		if !ok {
			callbacks[i] = func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}
		}
	}

	UeFsm, err := fsm.NewFSM(
		fsm.Transitions{
			{Event: RegistrationRequest, From: Deregistered, To: AuthenticationInitiated},
			{Event: AuthenticationSuccess, From: AuthenticationInitiated, To: Authenticated},
			{Event: RegistrationAccept, From: Authenticated, To: Registered},
			{Event: DeregistrationRequest, From: Registered, To: DeregisteredInitiated},
			{Event: Deregistration, From: DeregisteredInitiated, To: Deregistered},
			{Event: ForceDeregistrationInit, From: AuthenticationInitiated, To: DeregisteredInitiated},
			{Event: ForceDeregistrationInit, From: Authenticated, To: DeregisteredInitiated},
			{Event: ForceDeregistrationInit, From: Registered, To: DeregisteredInitiated},
			{Event: ForceDeregistrationInit, From: DeregisteredInitiated, To: DeregisteredInitiated},
		},
		fsm.Callbacks{
			AuthenticationInitiated: callbacks[AuthenticationInitiated],
			Authenticated:           callbacks[Authenticated],
			Registered:              callbacks[Registered],
			DeregisteredInitiated:   callbacks[DeregisteredInitiated],
			Deregistered:            callbacks[Deregistered],
		},
	)
	if UeFsm == nil || err != nil {
		return nil, fmt.Errorf("[5GC] Failed to create PDU FSM: %v", err.Error())
	}
	return UeFsm, nil
}

func initPduFSM(callbacks fsm.Callbacks) (*fsm.FSM, error) {

	if callbacks == nil {
		callbacks = fsm.Callbacks{}
	}
	states := []fsm.StateType{Inactive, InactivePending, Active, ModificationPending}
	for _, i := range states {
		_, ok := callbacks[i]
		if !ok {
			callbacks[i] = func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {}
		}
	}

	PduFsm, err := fsm.NewFSM(
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
	if PduFsm == nil || err != nil {
		return nil, fmt.Errorf("[5GC] Failed to create PDU FSM: %v", err.Error())
	}
	return PduFsm, nil
}
