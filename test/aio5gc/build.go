/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
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
)

type FiveGCBuilder struct {
	errors       []error
	config       config.Config
	nasHooks     map[nas.MsgType]func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)
	ngapHook     []func(ngapType.Message, *context.GNBContext, *context.Aio5gc) (bool, error)
	pduCallbacks map[fsm.StateType]fsm.Callback
	ueCallbacks  map[fsm.StateType]fsm.Callback
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = context.CloneConfig(conf)
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(ProcedureCode nas.MsgType, hook func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	if f.nasHooks == nil {
		f.nasHooks = map[nas.MsgType]func(nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error){}
	}
	_, ok := f.nasHooks[ProcedureCode]
	if ok {
		f.errors = append(f.errors, fmt.Errorf("duplicate NAS hook for %s", ProcedureCode))
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
		f.errors = append(f.errors, fmt.Errorf("duplicate UE callback for %s", state))
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
		f.errors = append(f.errors, fmt.Errorf("duplicate PDU callback for %s", state))
		return f
	}
	f.pduCallbacks[state] = callback
	return f
}

func (f *FiveGCBuilder) Build() (*context.Aio5gc, error) {
	if err := errors.Join(f.errors...); err != nil {
		return nil, err
	}
	amfId := "196673"                    // TODO generate ID
	amfName := "amf.5gc.3gppnetwork.org" // TODO generate Name

	fgc := context.Aio5gc{}
	if reflect.DeepEqual(f.config, config.Config{}) {
		return nil, errors.New("no configuration provided")
	}
	err := fgc.Init(f.config, amfId, amfName, f.ueCallbacks, f.pduCallbacks)
	if err != nil {
		return nil, err
	}

	if f.nasHooks != nil {
		fgc.SetNasHooks(f.nasHooks)
	}

	if f.ngapHook != nil {
		fgc.SetNgapHooks(f.ngapHook)
	}
	if len(f.config.AMFs) == 0 {
		return nil, errors.New("no AMF endpoints provided")
	}
	for index, amf := range f.config.AMFs {
		if amf == nil {
			_ = fgc.Close()
			return nil, fmt.Errorf("AMF endpoint %d is nil", index)
		}
		listener, err := service.Listen(amf.AddrPort)
		if err != nil {
			_ = fgc.Close()
			return nil, fmt.Errorf("start mock AMF: %w", err)
		}
		fgc.SetAMFEndpoint(index, config.AMF{IPv4Port: config.IPv4Port{AddrPort: listener.Addr()}})
		fgc.RegisterCloser(listener.Close)
		fgc.Go(func() { service.Serve(listener, &fgc) })
	}
	return &fgc, nil
}
