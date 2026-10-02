/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package interface_management

import (
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	codec "my5G-RANTester/lib/ngap"
)

func BuildNGSetupRequest(gnb *context.GNBContext) *message.NGSetupRequest {
	sst, sd := gnb.GetSliceInBytes()
	return &message.NGSetupRequest{
		GlobalRANNodeID: &ie.GlobalRANNodeID{Choice: &ie.GlobalGNBID{
			PLMNIdentity: gnb.GetPLMNIdentity(),
			GNBID:        &ie.GNBID{Choice: &ie.GNBIDForGNBID{Value: aper.BitString{Bytes: gnb.GetGnbIdInBytes(), BitLength: 24}}},
		}},
		RANNodeName: &ie.RANNodeName{Value: "my5gRANTester"},
		SupportedTAList: &ie.SupportedTAList{List: []ie.SupportedTAItem{{
			TAC: &ie.TAC{Value: gnb.GetTacInBytes()},
			BroadcastPLMNList: &ie.BroadcastPLMNList{List: []ie.BroadcastPLMNItem{{
				PLMNIdentity:        gnb.GetPLMNIdentity(),
				TAISliceSupportList: &ie.SliceSupportList{List: []ie.SliceSupportItem{{SNSSAI: codec.Slice(sst, sd)}}},
			}}},
		}}},
		DefaultPagingDRX: &ie.PagingDRX{Value: ie.PagingDRXPresentV128},
	}
}
func NGSetupRequest(gnb *context.GNBContext, name string) ([]byte, error) {
	value := BuildNGSetupRequest(gnb)
	value.RANNodeName.Value = aper.PrintableString(name)
	return value.MarshalBinary()
}
