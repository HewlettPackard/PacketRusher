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
)

type FiveGCBuilder struct {
	config   config.Config
	nasHooks map[uint8]func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)
	ngapHook []func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = conf
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(hooks map[uint8]func(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.nasHooks = hooks
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
	err := fgc.Init(f.config, amfId, amfName)
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
