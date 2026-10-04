/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"fmt"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
	"my5G-RANTester/test/aio5gc/msg/nas"
)

func resolveUE(amf *context.AMFContext, ran *ie.RANUENGAPID, core *ie.AMFUENGAPID) (*context.UEContext, error) {
	if ran == nil || core == nil {
		return nil, fmt.Errorf("missing UE NGAP identifiers")
	}
	ue, err := amf.FindUEById(core.Value)
	if err != nil {
		return nil, err
	}
	ranUE, err := amf.FindUEByRanId(ran.Value)
	if err != nil {
		return nil, err
	}
	if ue != ranUE {
		return nil, fmt.Errorf("RAN and AMF UE identifiers do not match")
	}
	return ue, nil
}
func UplinkNASTransport(req *message.UplinkNASTransport, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	ue, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID)
	if err != nil {
		return err
	}
	location, ok := req.UserLocationInformation.Choice.(*ie.UserLocationInformationNR)
	if !ok {
		return fmt.Errorf("mock core requires NR location")
	}
	model := convert.NRLocationToModels(location)
	model.GlobalGnbId = gnb.GetGlobalRanNodeID()
	ue.SetUserLocationInfo(model)
	nas.Dispatch(req.NASPDU, ue, fgc, gnb)
	return nil
}
