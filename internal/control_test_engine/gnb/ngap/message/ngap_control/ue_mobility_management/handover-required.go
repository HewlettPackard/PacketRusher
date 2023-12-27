/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap"
	log "github.com/sirupsen/logrus"

	"github.com/free5gc/aper"

	"github.com/free5gc/ngap/ngapType"
)

type HandoverRequiredBuilder struct {
	pdu ngapType.NGAPPDU
	ies *ngapType.ProtocolIEContainerHandoverRequiredIEs
}

func HandoverRequired(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewHandoverRequiredBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetHandoverType(ngapType.HandoverTypePresentIntra5gs).
		SetCause(ngapType.CauseRadioNetworkPresentHandoverDesirableForRadioReason).
		SetPduSessionResourceList(ue.GetPduSessions()).
		SetTargetGnodeB(targetGnb).
		SetSourceToTargetContainer(sourceGnb, targetGnb, ue.GetPduSessions(), ue.GetPrUeId()).
		Build()
}

func NewHandoverRequiredBuilder() *HandoverRequiredBuilder {
	pdu := ngapType.NGAPPDU{}

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeHandoverPreparation
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentHandoverRequired
	initiatingMessage.Value.HandoverRequired = new(ngapType.HandoverRequired)

	handoverRequired := initiatingMessage.Value.HandoverRequired
	ies := &handoverRequired.ProtocolIEs

	return &HandoverRequiredBuilder{pdu, ies}
}

func (builder *HandoverRequiredBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverRequiredBuilder {
	// AMF UE NGAP ID
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverRequiredBuilder {
	// RAN UE NGAP ID
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetHandoverType(handoverTypeValue aper.Enumerated) *HandoverRequiredBuilder {
	// Handover Type
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDHandoverType
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentHandoverType
	ie.Value.HandoverType = new(ngapType.HandoverType)

	handoverType := ie.Value.HandoverType
	handoverType.Value = handoverTypeValue

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetCause(causeValue aper.Enumerated) *HandoverRequiredBuilder {
	// Cause
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDCause
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentCause
	ie.Value.Cause = new(ngapType.Cause)

	cause := ie.Value.Cause
	cause.Present = ngapType.CausePresentRadioNetwork
	cause.RadioNetwork = new(ngapType.CauseRadioNetwork)
	cause.RadioNetwork.Value = causeValue

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetTargetGnodeB(targetGnb *context.GNBContext) *HandoverRequiredBuilder {
	// Target ID
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDTargetID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentTargetID
	ie.Value.TargetID = new(ngapType.TargetID)

	targetID := ie.Value.TargetID
	targetID.Present = ngapType.TargetIDPresentTargetRANNodeID
	targetID.TargetRANNodeID = new(ngapType.TargetRANNodeID)

	targetRANNodeID := targetID.TargetRANNodeID
	targetRANNodeID.GlobalRANNodeID.Present = ngapType.GlobalRANNodeIDPresentGlobalGNBID
	targetRANNodeID.GlobalRANNodeID.GlobalGNBID = new(ngapType.GlobalGNBID)

	globalRANNodeID := targetRANNodeID.GlobalRANNodeID
	globalRANNodeID.GlobalGNBID.PLMNIdentity = targetGnb.GetPLMNIdentity()
	globalRANNodeID.GlobalGNBID.GNBID.Present = ngapType.GNBIDPresentGNBID

	globalRANNodeID.GlobalGNBID.GNBID.GNBID = new(aper.BitString)

	gNBID := globalRANNodeID.GlobalGNBID.GNBID.GNBID

	*gNBID = aper.BitString{
		Bytes:     targetGnb.GetGnbIdInBytes(),
		BitLength: uint64(len(targetGnb.GetGnbIdInBytes()) * 8),
	}
	globalRANNodeID.GlobalGNBID.PLMNIdentity = targetGnb.GetPLMNIdentity()

	targetRANNodeID.SelectedTAI.PLMNIdentity = targetGnb.GetPLMNIdentity()
	targetRANNodeID.SelectedTAI.TAC.Value = targetGnb.GetTacInBytes()

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetPduSessionResourceList(pduSessions [16]*context.GnbPDUSession) *HandoverRequiredBuilder {
	// PDU Session Resource List
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDPDUSessionResourceListHORqd
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentPDUSessionResourceListHORqd
	ie.Value.PDUSessionResourceListHORqd = new(ngapType.PDUSessionResourceListHORqd)

	pDUSessionResourceListHORqd := ie.Value.PDUSessionResourceListHORqd

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		res, _ := aper.MarshalWithParams(ngapType.HandoverRequiredTransfer{}, "valueExt")

		pDUSessionResourceItem := ngapType.PDUSessionResourceItemHORqd{
			PDUSessionID: ngapType.PDUSessionID{
				Value: pduSession.GetPduSessionId(),
			},
			HandoverRequiredTransfer: res,
		}

		pDUSessionResourceListHORqd.List = append(pDUSessionResourceListHORqd.List, pDUSessionResourceItem)
	}
	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequiredBuilder) SetSourceToTargetContainer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) *HandoverRequiredBuilder {
	// Source to Target Transparent Container
	ie := ngapType.HandoverRequiredIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDSourceToTargetTransparentContainer
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentSourceToTargetTransparentContainer
	ie.Value.SourceToTargetTransparentContainer = new(ngapType.SourceToTargetTransparentContainer)

	ie.Value.SourceToTargetTransparentContainer.Value = GetSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, prUeId)

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func GetSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) []byte {
	data := buildSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, prUeId)
	encodeData, err := aper.MarshalWithParams(data, "valueExt")
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetSourceToTargetTransparentTransfer: %+v", err)
	}
	return encodeData
}

func buildSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) (data ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer) {

	// RRC Container
	data.RRCContainer.Value = aper.OctetString("\x00\x00\x11")

	data.IndexToRFSP = &ngapType.IndexToRFSP{Value: prUeId}

	// PDU Session Resource Information List
	data.PDUSessionResourceInformationList = new(ngapType.PDUSessionResourceInformationList)
	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		infoItem := ngapType.PDUSessionResourceInformationItem{}
		infoItem.PDUSessionID.Value = pduSession.GetPduSessionId()
		qosItem := ngapType.QosFlowInformationItem{}
		qosItem.QosFlowIdentifier.Value = 1
		infoItem.QosFlowInformationList.List = append(infoItem.QosFlowInformationList.List, qosItem)
		data.PDUSessionResourceInformationList.List = append(data.PDUSessionResourceInformationList.List, infoItem)
	}

	// Target Cell ID
	data.TargetCellID.Present = ngapType.TargetIDPresentTargetRANNodeID
	data.TargetCellID.NRCGI = new(ngapType.NRCGI)
	data.TargetCellID.NRCGI.PLMNIdentity = targetGnb.GetPLMNIdentity()
	data.TargetCellID.NRCGI.NRCellIdentity = targetGnb.GetNRCellIdentity()

	// UE History Information
	lastVisitedCellItem := ngapType.LastVisitedCellItem{}
	lastVisitedCellInfo := &lastVisitedCellItem.LastVisitedCellInformation
	lastVisitedCellInfo.Present = ngapType.LastVisitedCellInformationPresentNGRANCell
	lastVisitedCellInfo.NGRANCell = new(ngapType.LastVisitedNGRANCellInformation)
	ngRanCell := lastVisitedCellInfo.NGRANCell
	ngRanCell.GlobalCellID.Present = ngapType.NGRANCGIPresentNRCGI
	ngRanCell.GlobalCellID.NRCGI = new(ngapType.NRCGI)
	ngRanCell.GlobalCellID.NRCGI.PLMNIdentity = sourceGnb.GetPLMNIdentity()
	ngRanCell.GlobalCellID.NRCGI.NRCellIdentity = sourceGnb.GetNRCellIdentity()
	ngRanCell.CellType.CellSize.Value = ngapType.CellSizePresentVerysmall
	ngRanCell.TimeUEStayedInCell.Value = 10

	data.UEHistoryInformation.List = append(data.UEHistoryInformation.List, lastVisitedCellItem)
	return data
}

func (builder *HandoverRequiredBuilder) Build() ([]byte, error) {
	return ngap.Encoder(builder.pdu)
}
