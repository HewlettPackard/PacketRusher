/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"my5G-RANTester/config"
	"strconv"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
)

// All in one 5GC for test purpose
type Aio5gc struct {
	amfContext AMFContext
	session    SessionContext
	nasHook    []func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error)
	ngapHook   []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error)
	conf       config.Config
}

func (a *Aio5gc) GetAMFContext() *AMFContext {
	return &a.amfContext
}

func (a *Aio5gc) GetSessionContext() *SessionContext {
	return &a.session
}

func (a *Aio5gc) Init(conf config.Config, id string, name string) error {
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

	a.conf = conf
	a.amfContext = AMFContext{}
	a.amfContext.NewAmfContext(
		name,
		id,
		supportedPlmns,
		servedGuami,
		100,
	)

	a.session.NewSessionContext()
	return nil
}

func (a *Aio5gc) GetNasHooks() []func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error) {
	return a.nasHook
}

func (a *Aio5gc) SetNasHooks(hook []func(*nas.Message, *UEContext, *GNBContext, *Aio5gc) (bool, error)) {
	a.nasHook = hook
}

func (a *Aio5gc) GetNgapHooks() []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error) {
	return a.ngapHook
}

func (a *Aio5gc) SetNgapHooks(hook []func(*ngapType.NGAPPDU, *GNBContext, *Aio5gc) (bool, error)) {
	a.ngapHook = hook
}
