/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package aio5gc

import (
	"errors"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/service"
	"reflect"

	nas "github.com/free5gc/nas/message"
	ngapType "github.com/free5gc/ngap/message"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

type FiveGCBuilder struct {
	config       config.Config
	nasHooks     map[nas.MsgType]func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)
	ngapHook     []func(ngapType.Message, *context.GNBContext, *context.Aio5gc) (bool, error)
	pduCallbacks map[fsm.StateType]fsm.Callback
	ueCallbacks  map[fsm.StateType]fsm.Callback
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = conf
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(ProcedureCode nas.MsgType, hook func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	if f.nasHooks == nil {
		f.nasHooks = map[nas.MsgType]func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error){}
	}
	_, ok := f.nasHooks[ProcedureCode]
	if ok {
		log.Errorf("[5GC] Coudln't add NAS Hook with procedure code %d: already exist", ProcedureCode)
		return f
	}
	f.nasHooks[ProcedureCode] = hook
	return f
}

func (f *FiveGCBuilder) WithNGAPDispatcherHook(hook func(ngapType.Message, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
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
	if reflect.DeepEqual(f.config, config.Config{}) {
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
	for _, amf := range f.config.AMFs {
		listener, err := service.Listen(amf.AddrPort)
		if err != nil {
			_ = fgc.Close()
			return nil, fmt.Errorf("start mock AMF: %w", err)
		}
		fgc.SetAddr(listener.Addr())
		fgc.RegisterCloser(listener.Close)
		go service.Serve(listener, &fgc)
	}
	return &fgc, nil
}
