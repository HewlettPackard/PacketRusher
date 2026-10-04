/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
)

// Read immutable setup configuration without copying the concurrently updated
// AMF UE pool and ID allocator.
func NGSetupResponse(amf *context.AMFContext) ([]byte, error) {
	return BuilNGSetupResponse(amf.GetName(), amf.GetId(), amf.GetServedGuami(), amf.GetSupportedPlmnSnssai(), amf.GetRelativeCapacity()).MarshalBinary()
}
func BuilNGSetupResponse(name, id string, guamis []models.Guami, plmns []models.Nrf_NFMgmt_PlmnSnssai, capacity int64) *message.NGSetupResponse {
	msg := &message.NGSetupResponse{AMFName: &ie.AMFName{Value: aper.PrintableString(name)}, ServedGUAMIList: new(ie.ServedGUAMIList), RelativeAMFCapacity: &ie.RelativeAMFCapacity{Value: capacity}, PLMNSupportList: new(ie.PLMNSupportList)}
	for _, guami := range guamis {
		msg.ServedGUAMIList.List = append(msg.ServedGUAMIList.List, ie.ServedGUAMIItem{GUAMI: convert.GUAMIToNGAP(guami)})
	}
	for _, plmn := range plmns {
		item := ie.PLMNSupportItem{PLMNIdentity: convert.PLMNToNGAP(*plmn.PlmnId), SliceSupportList: new(ie.SliceSupportList)}
		for _, slice := range plmn.SNssaiList {
			item.SliceSupportList.List = append(item.SliceSupportList.List, ie.SliceSupportItem{SNSSAI: convert.SNSSAIToNGAP(models.Snssai{Sst: slice.Sst, Sd: slice.Sd})})
		}
		msg.PLMNSupportList.List = append(msg.PLMNSupportList.List, item)
	}
	return msg
}
