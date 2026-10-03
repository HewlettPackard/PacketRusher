/** SPDX-License-Identifier: Apache-2.0 */
package handler

import (
	"fmt"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
	"my5G-RANTester/test/aio5gc/msg/nas"
)

func resolveUE(amf *context.AMFContext, ran *ie.RANUENGAPID, core *ie.AMFUENGAPID, gnbs ...*context.GNBContext) (*context.UEContext, error) {
	if ran == nil || core == nil {
		return nil, fmt.Errorf("missing UE NGAP identifiers")
	}
	ue, err := amf.FindUEById(core.Value)
	if err != nil {
		return nil, err
	}
	// RAN IDs are scoped to a gNB. Two gNBs may both assign ID 1; the
	// globally unique AMF ID selects the UE whose RAN ID must then match.
	if ue.GetRanNgapId() != ran.Value {
		return nil, fmt.Errorf("RAN and AMF UE identifiers do not match")
	}
	if len(gnbs) > 0 && !ue.MatchesGNB(gnbs[0], ran.Value) {
		return nil, fmt.Errorf("UE identifiers belong to another gNB association")
	}
	return ue, nil
}
func UplinkNASTransport(req *message.UplinkNASTransport, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	ue, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID, gnb)
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
	return nas.Dispatch(req.NASPDU, ue, fgc, gnb)
}
