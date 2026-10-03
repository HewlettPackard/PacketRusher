/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	ngapConvert "my5G-RANTester/lib/ngap"

	ngap "github.com/free5gc/ngap/message"
	log "github.com/sirupsen/logrus"

	"github.com/free5gc/ngap/aper"

	ngapType "github.com/free5gc/ngap/ie"
)

type HandoverRequiredBuilder struct {
	pdu *ngap.HandoverRequired
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
	pdu := &ngap.HandoverRequired{}
	return &HandoverRequiredBuilder{pdu}
}

func (builder *HandoverRequiredBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverRequiredBuilder {
	builder.pdu.AMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *HandoverRequiredBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverRequiredBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *HandoverRequiredBuilder) SetHandoverType(handoverTypeValue aper.Enumerated) *HandoverRequiredBuilder {
	builder.pdu.HandoverType = &ngapType.HandoverType{Value: handoverTypeValue}
	return builder
}

func (builder *HandoverRequiredBuilder) SetCause(causeValue aper.Enumerated) *HandoverRequiredBuilder {
	builder.pdu.Cause = &ngapType.Cause{Choice: &ngapType.CauseRadioNetwork{Value: causeValue}}
	return builder
}

func (builder *HandoverRequiredBuilder) SetTargetGnodeB(targetGnb *context.GNBContext) *HandoverRequiredBuilder {
	builder.pdu.TargetID = &ngapType.TargetID{Choice: &ngapType.TargetRANNodeID{
		GlobalRANNodeID: &ngapType.GlobalRANNodeID{Choice: &ngapType.GlobalGNBID{
			PLMNIdentity: targetGnb.GetPLMNIdentity(),
			GNBID:        &ngapType.GNBID{Choice: &ngapType.GNBIDForGNBID{Value: aper.BitString{Bytes: targetGnb.GetGnbIdInBytes(), BitLength: 24}}},
		}},
		SelectedTAI: &ngapType.TAI{PLMNIdentity: targetGnb.GetPLMNIdentity(), TAC: &ngapType.TAC{Value: targetGnb.GetTacInBytes()}},
	}}
	return builder
}

func (builder *HandoverRequiredBuilder) SetPduSessionResourceList(pduSessions [16]*context.GnbPDUSession) *HandoverRequiredBuilder {
	list := &ngapType.PDUSessionResourceListHORqd{}
	for _, session := range pduSessions {
		if session == nil {
			continue
		}
		encoded, err := ngapConvert.Marshal(&ngapType.HandoverRequiredTransfer{})
		if err != nil {
			log.Errorf("[GNB][NGAP] Unable to encode handover transfer: %v", err)
			continue
		}
		list.List = append(list.List, ngapType.PDUSessionResourceItemHORqd{PDUSessionID: &ngapType.PDUSessionID{Value: session.GetPduSessionId()}, HandoverRequiredTransfer: ngapConvert.Octets(encoded)})
	}
	if len(list.List) > 0 {
		builder.pdu.PDUSessionResourceListHORqd = list
	}
	return builder
}

func (builder *HandoverRequiredBuilder) SetSourceToTargetContainer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) *HandoverRequiredBuilder {
	// Source to Target Transparent Container
	ie := ngapType.HandoverRequiredIEs{}

	ie.SourceToTargetTransparentContainer = new(ngapType.SourceToTargetTransparentContainer)

	ie.SourceToTargetTransparentContainer.Value = GetSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, prUeId)

	builder.pdu.SourceToTargetTransparentContainer = ie.SourceToTargetTransparentContainer

	return builder
}

func GetSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) []byte {
	data := buildSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, prUeId)
	encodeData, err := ngapConvert.Marshal(&data)
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetSourceToTargetTransparentTransfer: %+v", err)
	}
	return encodeData
}

func buildSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) (data ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer) {
	data.RRCContainer = &ngapType.RRCContainer{Value: aper.OctetString("\x00\x00\x11")}
	data.IndexToRFSP = &ngapType.IndexToRFSP{Value: prUeId}
	list := &ngapType.PDUSessionResourceInformationList{}
	for _, session := range pduSessions {
		if session == nil {
			continue
		}
		list.List = append(list.List, ngapType.PDUSessionResourceInformationItem{
			PDUSessionID:           &ngapType.PDUSessionID{Value: session.GetPduSessionId()},
			QosFlowInformationList: &ngapType.QosFlowInformationList{List: []ngapType.QosFlowInformationItem{{QosFlowIdentifier: &ngapType.QosFlowIdentifier{Value: 1}}}},
		})
	}
	if len(list.List) > 0 {
		data.PDUSessionResourceInformationList = list
	}
	data.TargetCellID = &ngapType.NGRANCGI{Choice: &ngapType.NRCGI{PLMNIdentity: targetGnb.GetPLMNIdentity(), NRCellIdentity: targetGnb.GetNRCellIdentity()}}
	data.UEHistoryInformation = &ngapType.UEHistoryInformation{List: []ngapType.LastVisitedCellItem{{
		LastVisitedCellInformation: &ngapType.LastVisitedCellInformation{Choice: &ngapType.LastVisitedNGRANCellInformation{
			GlobalCellID:       &ngapType.NGRANCGI{Choice: &ngapType.NRCGI{PLMNIdentity: sourceGnb.GetPLMNIdentity(), NRCellIdentity: sourceGnb.GetNRCellIdentity()}},
			CellType:           &ngapType.CellType{CellSize: &ngapType.CellSize{Value: ngapType.CellSizePresentVerysmall}},
			TimeUEStayedInCell: &ngapType.TimeUEStayedInCell{Value: 10},
		}},
	}}}
	return data
}

func (builder *HandoverRequiredBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}
