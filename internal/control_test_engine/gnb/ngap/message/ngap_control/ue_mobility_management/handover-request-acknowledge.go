/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023-2026 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap/aper"
	ngap "github.com/free5gc/ngap/message"
	log "my5G-RANTester/internal/log"

	ngapType "github.com/free5gc/ngap/ie"
	ngapConvert "my5G-RANTester/lib/ngap"
)

type HandoverRequestAcknowledgeBuilder struct {
	pdu *ngap.HandoverRequestAcknowledge
}

func HandoverRequestAcknowledge(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewHandoverRequestAcknowledgeBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetPduSessionResourceAdmittedList(gnb, ue.GetPduSessions()).
		SetTargetToSourceContainer().
		Build()
}

func NewHandoverRequestAcknowledgeBuilder() *HandoverRequestAcknowledgeBuilder {
	pdu := &ngap.HandoverRequestAcknowledge{}
	return &HandoverRequestAcknowledgeBuilder{pdu}
}

func (builder *HandoverRequestAcknowledgeBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverRequestAcknowledgeBuilder {
	builder.pdu.AMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverRequestAcknowledgeBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetPduSessionResourceAdmittedList(gnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession) *HandoverRequestAcknowledgeBuilder {
	ie := ngapType.HandoverRequestAcknowledgeIEs{}

	ie.PDUSessionResourceAdmittedList = new(ngapType.PDUSessionResourceAdmittedList)

	pDUSessionResourceAdmittedList := ie.PDUSessionResourceAdmittedList

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		//PDU SessionResource Admittedy Item
		pDUSessionResourceAdmittedItem := ngapType.PDUSessionResourceAdmittedItem{}
		pDUSessionResourceAdmittedItem.PDUSessionID = &ngapType.PDUSessionID{Value: pduSession.GetPduSessionId()}
		pDUSessionResourceAdmittedItem.HandoverRequestAcknowledgeTransfer = ngapConvert.Octets(GetHandoverRequestAcknowledgeTransfer(gnb, pduSession))

		pDUSessionResourceAdmittedList.List = append(pDUSessionResourceAdmittedList.List, pDUSessionResourceAdmittedItem)
	}

	if len(pDUSessionResourceAdmittedList.List) == 0 {
		log.Info("[GNB][NGAP] No admitted PDU Session")
		return builder
	}

	builder.pdu.PDUSessionResourceAdmittedList = ie.PDUSessionResourceAdmittedList

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) SetTargetToSourceContainer() *HandoverRequestAcknowledgeBuilder {
	// Target To Source TransparentContainer
	ie := ngapType.HandoverRequestAcknowledgeIEs{}

	ie.TargetToSourceTransparentContainer = new(ngapType.TargetToSourceTransparentContainer)

	targetToSourceTransparentContainer := ie.TargetToSourceTransparentContainer
	targetToSourceTransparentContainer.Value = GetTargetToSourceTransparentTransfer()
	builder.pdu.TargetToSourceTransparentContainer = ie.TargetToSourceTransparentContainer

	return builder
}

func (builder *HandoverRequestAcknowledgeBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}

func GetHandoverRequestAcknowledgeTransfer(gnb *context.GNBContext, pduSession *context.GnbPDUSession) []byte {
	data := buildHandoverRequestAcknowledgeTransfer(gnb, pduSession)
	encodeData, err := ngapConvert.Marshal(&data)
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetHandoverRequestAcknowledgeTransfer: %+v", err)
	}
	return encodeData
}

func buildHandoverRequestAcknowledgeTransfer(gnb *context.GNBContext, pduSession *context.GnbPDUSession) (data ngapType.HandoverRequestAcknowledgeTransfer) {
	data.DLNGUUPTNLInformation = ngapConvert.UPTransport(gnb.GetN3GnbIp(), pduSession.GetTeidDownlink())
	data.QosFlowSetupResponseList = &ngapType.QosFlowListWithDataForwarding{List: []ngapType.QosFlowItemWithDataForwarding{{QosFlowIdentifier: &ngapType.QosFlowIdentifier{Value: 1}}}}
	return data
}

func GetTargetToSourceTransparentTransfer() []byte {
	data := buildTargetToSourceTransparentTransfer()
	encodeData, err := ngapConvert.Marshal(&data)
	if err != nil {
		log.Fatalf("aper MarshalWithParams error in GetTargetToSourceTransparentTransfer: %+v", err)
	}
	return encodeData
}

func buildTargetToSourceTransparentTransfer() (data ngapType.TargetNGRANNodeToSourceNGRANNodeTransparentContainer) {
	data.RRCContainer = &ngapType.RRCContainer{Value: aper.OctetString("\x00\x00\x11")}
	return data
}
