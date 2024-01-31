/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"

	"github.com/free5gc/ngap/ngapType"
)

func InitialContextSetupResponse(req *ngapType.InitialContextSetupResponse, fgc *context.Aio5gc) error {
	amf := fgc.GetAMFContext()
	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error

	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ranUe, err = amf.FindUEByRanId(req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value)
			if err != nil {
				return err
			}
		case ngapType.ProtocolIEIDAMFUENGAPID:
			ue, err = amf.FindUEById(req.ProtocolIEs.List[ie].Value.AMFUENGAPID.Value)
			if err != nil {
				return err
			}
		}
	}
	if ue != ranUe {
		return errors.New("[5GC][NGAP] RanUeNgapId does not match the one Registered for this UE")
	}
	return nil
}
