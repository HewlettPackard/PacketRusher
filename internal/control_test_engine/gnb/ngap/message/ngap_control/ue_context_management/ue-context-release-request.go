/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Valentin D'Emmanuele
 */
package ue_context_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap/aper"
	ngap "github.com/free5gc/ngap/message"

	ngapType "github.com/free5gc/ngap/ie"
)

type UeContextReleaseRequestBuilder struct {
	pdu *ngap.UEContextReleaseRequest
}

func UeContextReleaseRequest(ue *context.GNBUe) ([]byte, error) {
	return NewUeContextReleaseRequestBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetPduSessionResourceListCxtRelReq(ue.GetPduSessions()).
		SetCause(ngapType.CauseRadioNetworkPresentUserInactivity).
		Build()
}

func NewUeContextReleaseRequestBuilder() *UeContextReleaseRequestBuilder {
	pdu := &ngap.UEContextReleaseRequest{}
	return &UeContextReleaseRequestBuilder{pdu}
}

func (builder *UeContextReleaseRequestBuilder) SetAmfUeNgapId(amfUeNgapID int64) *UeContextReleaseRequestBuilder {
	builder.pdu.AMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *UeContextReleaseRequestBuilder) SetRanUeNgapId(ranUeNgapID int64) *UeContextReleaseRequestBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *UeContextReleaseRequestBuilder) SetPduSessionResourceListCxtRelReq(pduSessions [16]*context.GnbPDUSession) *UeContextReleaseRequestBuilder {
	activePduSession := []*context.GnbPDUSession{}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		activePduSession = append(activePduSession, pduSession)
	}

	// PDU Session Resource List
	if len(activePduSession) > 0 {
		ie := ngapType.UEContextReleaseRequestIEs{}

		ie.PDUSessionResourceListCxtRelReq = new(ngapType.PDUSessionResourceListCxtRelReq)

		pDUSessionResourceListCxtRelReq := ie.PDUSessionResourceListCxtRelReq

		// PDU Session Resource Item in PDU session Resource List
		for _, pduSessionID := range activePduSession {
			pDUSessionResourceItem := ngapType.PDUSessionResourceItemCxtRelReq{}
			pDUSessionResourceItem.PDUSessionID = &ngapType.PDUSessionID{Value: pduSessionID.GetPduSessionId()}
			pDUSessionResourceListCxtRelReq.List = append(pDUSessionResourceListCxtRelReq.List, pDUSessionResourceItem)
		}
		builder.pdu.PDUSessionResourceListCxtRelReq = ie.PDUSessionResourceListCxtRelReq
	}

	return builder
}

func (builder *UeContextReleaseRequestBuilder) SetCause(causeValue aper.Enumerated) *UeContextReleaseRequestBuilder {
	builder.pdu.Cause = &ngapType.Cause{Choice: &ngapType.CauseRadioNetwork{Value: causeValue}}
	return builder
}

func (builder *UeContextReleaseRequestBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}
