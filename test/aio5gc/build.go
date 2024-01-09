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
)

type FiveGCBuilder struct {
	config   config.Config
	fsm      *fsm.FSM
	nasHook  []func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)
	ngapHook []func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = conf
	return f
}

func (f *FiveGCBuilder) WithFSM(fsm *fsm.FSM) *FiveGCBuilder {
	f.fsm = fsm
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(hook func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.nasHook = append(f.nasHook, hook)
	return f
}

func (f *FiveGCBuilder) WithNGAPDispatcherHook(hook func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.ngapHook = append(f.ngapHook, hook)
	return f
}

func (f *FiveGCBuilder) Build() (*context.Aio5gc, error) {
	amfId := "196673"                    // TODO generate ID
	amfName := "amf.5gc.3gppnetwork.org" // TODO generate Name

	fgc := context.Aio5gc{}
	if (f.config == config.Config{}) {
		return &context.Aio5gc{}, errors.New("No configuration provided")
	}
	if f.fsm == nil {
		return &context.Aio5gc{}, errors.New("No FSM provided")
	}
	err := fgc.Init(f.config, amfId, amfName, f.fsm)
	if err != nil {
		return &context.Aio5gc{}, err
	}
	if f.nasHook != nil {
		fgc.SetNasHooks(f.nasHook)
	}

	if f.ngapHook != nil {
		fgc.SetNgapHooks(f.ngapHook)
	}
	go service.RunServer(f.config.AMF.Ip, f.config.AMF.Port, &fgc)
	return &fgc, nil
}
