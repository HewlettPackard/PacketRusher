/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package aio5gc

import (
	"errors"
	"my5G-RANTester/config"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/service"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

type FiveGCBuilder struct {
	config       config.Config
	nasHooks     map[uint8]func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)
	ngapHook     []func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)
	pduCallbacks map[fsm.StateType]fsm.Callback
	ueCallbacks  map[fsm.StateType]fsm.Callback
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = conf
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(ProcedureCode uint8, hook func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	if f.nasHooks == nil {
		f.nasHooks = map[uint8]func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error){}
	}
	_, ok := f.nasHooks[ProcedureCode]
	if ok {
		log.Errorf("[5GC] Coudln't add NAS Hook with procedure code %d: already exist", ProcedureCode)
		return f
	}
	f.nasHooks[ProcedureCode] = hook
	return f
}

func (f *FiveGCBuilder) WithNGAPDispatcherHook(hook func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.ngapHook = append(f.ngapHook, hook)
	return f
}

func (f *FiveGCBuilder) WithUeCallback(state fsm.StateType, callback fsm.Callback) *FiveGCBuilder {
	if f.ueCallbacks == nil {
		f.ueCallbacks = map[fsm.StateType]fsm.Callback{}
	}
	_, ok := f.ueCallbacks[state]
	if ok {
		log.Errorf("[5GC] Coudln't add ue state change callback for state %v: already exist", state)
		return f
	}
	f.ueCallbacks[state] = callback
	return f
}

func (f *FiveGCBuilder) WithPDUCallback(state fsm.StateType, callback fsm.Callback) *FiveGCBuilder {
	if f.pduCallbacks == nil {
		f.pduCallbacks = map[fsm.StateType]fsm.Callback{}
	}
	_, ok := f.pduCallbacks[state]
	if ok {
		log.Errorf("[5GC] Coudln't add pdu session state change callback for state %v: already exist", state)
		return f
	}
	f.pduCallbacks[state] = callback
	return f
}

func (f *FiveGCBuilder) Build() (*context.Aio5gc, error) {
	amfId := "196673"                    // TODO generate ID
	amfName := "amf.5gc.3gppnetwork.org" // TODO generate Name

	fgc := context.Aio5gc{}
	if (f.config == config.Config{}) {
		return &context.Aio5gc{}, errors.New("No configuration provided")
	}
	err := fgc.Init(f.config, amfId, amfName, f.ueCallbacks, f.pduCallbacks)
	if err != nil {
		return &context.Aio5gc{}, err
	}

	if f.nasHooks != nil {
		fgc.SetNasHooks(f.nasHooks)
	}

	if f.ngapHook != nil {
		fgc.SetNgapHooks(f.ngapHook)
	}
	go service.RunServer(f.config.AMF.Ip, f.config.AMF.Port, &fgc)
	return &fgc, nil
}
