package ngapMsgHandler

import (
	"errors"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"
	"my5G-RANTester/test/amf/lib/types"

	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
)

func NGSetupResponse(amf context.AMFContext) ([]byte, error) {
	// keeping it simple for now, amf will return all saved Guami and nssais. Will probably need to make amf return those based on asked PLMN/Nssai
	message := BuilNGSetupResponse(amf.GetName(), amf.GetId(), amf.GetServedGuami(), amf.GetSupportedPlmnSnssai(), amf.GetRelativeCapacity())
	return ngap.Encoder(message)
}

func BuilNGSetupResponse(amfName string, amfId string, servedGuamis []models.Guami, supportedPlmnSnssai []models.PlmnSnssai, relativeCapacity int64) (pdu ngapType.NGAPPDU) {
	pdu.Present = ngapType.NGAPPDUPresentSuccessfulOutcome
	//assuming it's always successful for now
	pdu.SuccessfulOutcome = new(ngapType.SuccessfulOutcome)

	pdu.SuccessfulOutcome.ProcedureCode.Value = ngapType.ProcedureCodeNGSetup
	pdu.SuccessfulOutcome.Criticality.Value = ngapType.CriticalityPresentReject

	pdu.SuccessfulOutcome.Value.Present = ngapType.SuccessfulOutcomePresentNGSetupResponse
	pdu.SuccessfulOutcome.Value.NGSetupResponse = new(ngapType.NGSetupResponse)

	nGSetupResponseIEs := &pdu.SuccessfulOutcome.Value.NGSetupResponse.ProtocolIEs

	// AMFName
	ie := addProtocolIEAMFName(amfName)
	nGSetupResponseIEs.List = append(nGSetupResponseIEs.List, ie)

	// SERVEDGUAMIList
	ie = addProtocolIEServedGUAMIList(servedGuamis)
	nGSetupResponseIEs.List = append(nGSetupResponseIEs.List, ie)

	// RelativeAMFCapacity
	ie = addProtocolIERelativeAMFCapacity(relativeCapacity)
	nGSetupResponseIEs.List = append(nGSetupResponseIEs.List, ie)

	// PLMNSupportList
	ie = addProtocolIEPLMNSupportList(supportedPlmnSnssai)
	nGSetupResponseIEs.List = append(nGSetupResponseIEs.List, ie)

	return
}

func addProtocolIEAMFName(amfName string) ngapType.NGSetupResponseIEs {
	ie := ngapType.NGSetupResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFName
	ie.Criticality.Value = ngapType.CriticalityPresentReject

	ie.Value.Present = ngapType.NGSetupResponseIEsPresentAMFName
	ie.Value.AMFName = new(ngapType.AMFName)
	ie.Value.AMFName.Value = amfName
	return ie
}

func addProtocolIEServedGUAMIList(servedGuamis []models.Guami) ngapType.NGSetupResponseIEs {
	ie := ngapType.NGSetupResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDServedGUAMIList
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.NGSetupResponseIEsPresentServedGUAMIList

	ie.Value.ServedGUAMIList = new(ngapType.ServedGUAMIList)
	servedGUAMIList := ie.Value.ServedGUAMIList
	for i := 0; i < len(servedGuamis); i++ {
		regionId, setId, pointerId := ngapConvert.AmfIdToNgap(servedGuamis[i].AmfId)
		servedGUAMIItem := ngapType.ServedGUAMIItem{
			GUAMI: ngapType.GUAMI{
				PLMNIdentity: ngapConvert.PlmnIdToNgap(*servedGuamis[i].PlmnId),
				AMFRegionID: ngapType.AMFRegionID{
					Value: regionId,
				},
				AMFSetID: ngapType.AMFSetID{
					Value: setId,
				},
				AMFPointer: ngapType.AMFPointer{
					Value: pointerId,
				},
			},
		}
		servedGUAMIList.List = append(servedGUAMIList.List, servedGUAMIItem)
	}
	return ie
}

func addProtocolIERelativeAMFCapacity(relativeCapacity int64) ngapType.NGSetupResponseIEs {
	ie := ngapType.NGSetupResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRelativeAMFCapacity
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.NGSetupResponseIEsPresentRelativeAMFCapacity
	ie.Value.RelativeAMFCapacity = new(ngapType.RelativeAMFCapacity)
	ie.Value.RelativeAMFCapacity.Value = relativeCapacity
	return ie
}

func addProtocolIEPLMNSupportList(supportedPlmnSnssai []models.PlmnSnssai) ngapType.NGSetupResponseIEs {
	ie := ngapType.NGSetupResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDPLMNSupportList
	ie.Value.Present = ngapType.NGSetupResponseIEsPresentPLMNSupportList
	ie.Criticality.Value = ngapType.CriticalityPresentReject

	ie.Value.PLMNSupportList = new(ngapType.PLMNSupportList)
	pLMNSupportList := ie.Value.PLMNSupportList
	for plmn := range supportedPlmnSnssai {
		sliceSupportList := ngapType.SliceSupportList{}

		for nssai := range supportedPlmnSnssai[plmn].SNssaiList {
			sliceSupportItem := ngapType.SliceSupportItem{
				SNSSAI: ngapConvert.SNssaiToNgap(supportedPlmnSnssai[plmn].SNssaiList[nssai]),
			}
			sliceSupportList.List = append(sliceSupportList.List, sliceSupportItem)
		}

		pLMNSupportItem := ngapType.PLMNSupportItem{
			PLMNIdentity:     ngapConvert.PlmnIdToNgap(*supportedPlmnSnssai[plmn].PlmnId),
			SliceSupportList: sliceSupportList,
		}
		pLMNSupportList.List = append(pLMNSupportList.List, pLMNSupportItem)
	}
	return ie
}

func NGSetupRequest(req *ngapType.NGSetupRequest, amf context.AMFContext, gnb context.GNBContext) (msg []byte, err error) {
	// assert req contains wanted values?
	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDGlobalRANNodeID:
			globalRANNodeID := ngapConvert.RanIdToModels(*req.ProtocolIEs.List[ie].Value.GlobalRANNodeID)
			gnb.SetGlobalRanNodeID(globalRANNodeID)
		case ngapType.ProtocolIEIDRANNodeName:
			gnb.SetRanNodename(req.ProtocolIEs.List[ie].Value.RANNodeName.Value)
		case ngapType.ProtocolIEIDSupportedTAList:
			supportedTaiList := types.TaiListToModels(*req.ProtocolIEs.List[ie].Value.SupportedTAList)
			gnb.SetSuportedTAList(supportedTaiList)
		case ngapType.ProtocolIEIDDefaultPagingDRX:
			gnb.SetDefautlPagingDRX(*req.ProtocolIEs.List[ie].Value.DefaultPagingDRX)
		default:
			return msg, errors.New("[AMF][NGAP] Received unknown ie for NGSetupRequest")
		}
	}

	msg, err = NGSetupResponse(amf)
	if err != nil {
		return msg, err
	}
	log.Info("[AMF][NGAP] Sent NG Setup Response")

	return msg, err
}
