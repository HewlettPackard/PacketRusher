package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
)

func PathSwitchRequest(ue *context.GNBUe) ([]byte, error) {
	message := BuildPathSwitchRequest(ue.GetAmfUeId(), ue.GetRanUeId())

	return ngap.Encoder(message)
}

func BuildPathSwitchRequest(amfUeNgapID, ranUeNgapID int64) (pdu ngapType.NGAPPDU) {

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodePathSwitchRequest
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentPathSwitchRequest
	initiatingMessage.Value.PathSwitchRequest = new(ngapType.PathSwitchRequest)

	pathSwitchRequest := initiatingMessage.Value.PathSwitchRequest
	pathSwitchRequestIEs := &pathSwitchRequest.ProtocolIEs

	// SOURCE AMF UE NGAP ID
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentSourceAMFUENGAPID
	ie.Value.SourceAMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.SourceAMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	pathSwitchRequestIEs.List = append(pathSwitchRequestIEs.List, ie)

	// RAN UE NGAP ID
	ie = ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	pathSwitchRequestIEs.List = append(pathSwitchRequestIEs.List, ie)

	return
}
