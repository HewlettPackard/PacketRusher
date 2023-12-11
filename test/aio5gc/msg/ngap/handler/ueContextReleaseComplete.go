/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
)

func UEContextReleaseComplete(req *ngapType.UEContextReleaseComplete, fgc *context.Aio5gc) error {

	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error
	amf := fgc.GetAMFContext()

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

		case ngapType.ProtocolIEIDUserLocationInformation:

		default:
			return errors.New("[5GC][NGAP] Received unknown ie for UEContextReleaseComplete")
		}
	}
	if ue != ranUe {
		return errors.New("[5GC][NGAP] RanUeNgapId does not match the one registred for this UE")
	}
	return nil
}
