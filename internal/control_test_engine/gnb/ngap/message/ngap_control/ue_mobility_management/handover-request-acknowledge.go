/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"encoding/binary"
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap"
	log "github.com/sirupsen/logrus"

	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
)

type HandoverRequestAcknowledgeBuilder struct {
	pdu ngapType.NGAPPDU
	ies *ngapType.ProtocolIEContainerHandoverRequestAcknowledgeIEs
}

func HandoverRequestAcknowledge(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewHandoverRequestAcknowledgeBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetPduSessionResourceAdmittedList(gnb, ue.GetPduSessions()).
		SetTargetToSourceContainer().
		Build()
}

func NewHandoverRequestAcknowledgeBuilder() *HandoverRequestAcknowledgeBuilder {
	pdu := ngapType.NGAPPDU{}

	pdu.Present = ngapType.NGAPPDUPresentSuccessfulOutcome
	pdu.SuccessfulOutcome = new(ngapType.SuccessfulOutcome)

	successfulOutcome := pdu.SuccessfulOutcome
	successfulOutcome.ProcedureCode.Value = ngapType.ProcedureCodeHandoverResourceAllocation
	successfulOutcome.Criticality.Value = ngapType.CriticalityPresentReject

	successfulOutcome.Value.Present = ngapType.SuccessfulOutcomePresentHandoverRequestAcknowledge
	successfulOutcome.Value.HandoverRequestAcknowledge = new(ngapType.HandoverRequestAcknowledge)

	handoverRequestAcknowledge := successfulOutcome.Value.HandoverRequestAcknowledge
	ies := &handoverRequestAcknowledge.ProtocolIEs

	return &HandoverRequestAcknowledgeBuilder{pdu, ies}
}

func (builder *HandoverRequestAcknowledgeBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverRequestAcknowledgeBuilder {
	// AMF UE NGAP ID
	ie := ngapType.HandoverRequestAcknowledgeIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverRequestAcknowledgeBuilder {
	// RAN UE NGAP ID
	ie := ngapType.HandoverRequestAcknowledgeIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetPduSessionResourceAdmittedList(gnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession) *HandoverRequestAcknowledgeBuilder {
	ie := ngapType.HandoverRequestAcknowledgeIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDPDUSessionResourceAdmittedList
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.HandoverRequestAcknowledgeIEsPresentPDUSessionResourceAdmittedList
	ie.Value.PDUSessionResourceAdmittedList = new(ngapType.PDUSessionResourceAdmittedList)

	pDUSessionResourceAdmittedList := ie.Value.PDUSessionResourceAdmittedList

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		//PDU SessionResource Admittedy Item
		pDUSessionResourceAdmittedItem := ngapType.PDUSessionResourceAdmittedItem{}
		pDUSessionResourceAdmittedItem.PDUSessionID.Value = pduSession.GetPduSessionId()
		pDUSessionResourceAdmittedItem.HandoverRequestAcknowledgeTransfer = GetHandoverRequestAcknowledgeTransfer(gnb, pduSession)

		pDUSessionResourceAdmittedList.List = append(pDUSessionResourceAdmittedList.List, pDUSessionResourceAdmittedItem)
	}

	if len(pDUSessionResourceAdmittedList.List) == 0 {
		log.Info("[GNB][NGAP] No admitted PDU Session")
		return builder
	}

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetTargetToSourceContainer() *HandoverRequestAcknowledgeBuilder {
	// Target To Source TransparentContainer
	ie := ngapType.HandoverRequestAcknowledgeIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDTargetToSourceTransparentContainer
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequestAcknowledgeIEsPresentTargetToSourceTransparentContainer
	ie.Value.TargetToSourceTransparentContainer = new(ngapType.TargetToSourceTransparentContainer)

	targetToSourceTransparentContainer := ie.Value.TargetToSourceTransparentContainer
	targetToSourceTransparentContainer.Value = GetTargetToSourceTransparentTransfer()
	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) Build() ([]byte, error) {
	return ngap.Encoder(builder.pdu)
}

func GetHandoverRequestAcknowledgeTransfer(gnb *context.GNBContext, pduSession *context.GnbPDUSession) []byte {
	data := buildHandoverRequestAcknowledgeTransfer(gnb, pduSession)
	encodeData, err := aper.MarshalWithParams(data, "valueExt")
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetHandoverRequestAcknowledgeTransfer: %+v", err)
	}
	return encodeData
}

func buildHandoverRequestAcknowledgeTransfer(gnb *context.GNBContext, pduSession *context.GnbPDUSession) (data ngapType.HandoverRequestAcknowledgeTransfer) {

	// DL NG-U UP TNL information
	dlTransportLayerInformation := &data.DLNGUUPTNLInformation
	dlTransportLayerInformation.Present = ngapType.UPTransportLayerInformationPresentGTPTunnel
	dlTransportLayerInformation.GTPTunnel = new(ngapType.GTPTunnel)
	downlinkTeid := make([]byte, 4)
	binary.BigEndian.PutUint32(downlinkTeid, pduSession.GetTeidDownlink())
	dlTransportLayerInformation.GTPTunnel.GTPTEID.Value = downlinkTeid
	dlTransportLayerInformation.GTPTunnel.TransportLayerAddress = ngapConvert.IPAddressToNgap(gnb.GetN3GnbIp().String(), "")

	// Qos Flow Setup Response List
	qosFlowSetupResponseItem := ngapType.QosFlowItemWithDataForwarding{
		QosFlowIdentifier: ngapType.QosFlowIdentifier{
			Value: 1,
		},
	}

	data.QosFlowSetupResponseList.List = append(data.QosFlowSetupResponseList.List, qosFlowSetupResponseItem)

	return data
}

func GetTargetToSourceTransparentTransfer() []byte {
	data := buildTargetToSourceTransparentTransfer()
	encodeData, err := aper.MarshalWithParams(data, "valueExt")
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetTargetToSourceTransparentTransfer: %+v", err)
	}
	return encodeData
}

func buildTargetToSourceTransparentTransfer() (data ngapType.TargetNGRANNodeToSourceNGRANNodeTransparentContainer) {
	// RRC Container
	data.RRCContainer.Value = aper.OctetString("\x00\x00\x11")

	return data
}
