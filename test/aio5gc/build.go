package aio5gc

import (
	"errors"
	"my5G-RANTester/config"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/service"

	"github.com/free5gc/nas"
)

type FiveGCBuilder struct {
	config   config.Config
	nasHook  func(*nas.Message, *context.UEContext, *context.Aio5gc) (bool, error)
	ngapHook func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)
}

func (f *FiveGCBuilder) WithConfig(conf config.Config) *FiveGCBuilder {
	f.config = conf
	return f
}

func (f *FiveGCBuilder) WithNASDispatcherHook(hook func(*nas.Message, *context.UEContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.nasHook = hook
	return f
}

func (f *FiveGCBuilder) WithNGAPDispatcherHook(hook func(*ngapType.NGAPPDU, *context.GNBContext, *context.Aio5gc) (bool, error)) *FiveGCBuilder {
	f.ngapHook = hook
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
	if f.nasHook != nil {
		fgc.SetNasHook(f.nasHook)
	}

	if f.ngapHook != nil {
		fgc.SetNgapHook(f.ngapHook)
	}
	go service.RunServer(f.config.AMF.Ip, f.config.AMF.Port, &fgc)
	return &fgc, nil
}
