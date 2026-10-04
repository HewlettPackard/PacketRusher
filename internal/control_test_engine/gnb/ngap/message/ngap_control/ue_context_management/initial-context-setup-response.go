/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2024-2026 Valentin D'Emmanuele
 */
package ue_context_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"
	ngapConvert "my5G-RANTester/lib/ngap"

	ngapType "github.com/free5gc/ngap/ie"
	ngap "github.com/free5gc/ngap/message"
	log "my5G-RANTester/internal/log"
)

type InitialContextSetupResponseBuilder struct {
	pdu *ngap.InitialContextSetupResponse
}

func InitialContextSetupResponse(ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	return NewInitialContextSetupResponseBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetPDUSessionResourceSetupListCxtRes(gnb, ue.GetPduSessions()).
		Build()
}

func NewInitialContextSetupResponseBuilder() *InitialContextSetupResponseBuilder {
	pdu := &ngap.InitialContextSetupResponse{}
	return &InitialContextSetupResponseBuilder{pdu}
}

func (builder *InitialContextSetupResponseBuilder) SetAmfUeNgapId(amfUeNgapID int64) *InitialContextSetupResponseBuilder {
	builder.pdu.AMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *InitialContextSetupResponseBuilder) SetRanUeNgapId(ranUeNgapID int64) *InitialContextSetupResponseBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *InitialContextSetupResponseBuilder) SetPDUSessionResourceSetupListCxtRes(gnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession) *InitialContextSetupResponseBuilder {
	// PDU Session Resource Setup List Cxt Res
	ie := ngapType.InitialContextSetupResponseIEs{}

	ie.PDUSessionResourceSetupListCxtRes = new(ngapType.PDUSessionResourceSetupListCxtRes)

	PDUSessionResourceSetupListCxtRes := ie.PDUSessionResourceSetupListCxtRes
	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		pDUSessionResourceSetupItemCxtRes := ngapType.PDUSessionResourceSetupItemCxtRes{}

		pDUSessionResourceSetupItemCxtRes.PDUSessionID = &ngapType.PDUSessionID{Value: pduSession.GetPduSessionId()}
		pDUSessionResourceSetupItemCxtRes.PDUSessionResourceSetupResponseTransfer = ngapConvert.Octets(pdu_session_management.GetPDUSessionResourceSetupResponseTransfer(gnb.GetN3GnbIp(), pduSession.GetTeidDownlink(), pduSession.GetQosId()))
		PDUSessionResourceSetupListCxtRes.List = append(PDUSessionResourceSetupListCxtRes.List, pDUSessionResourceSetupItemCxtRes)
	}

	if len(PDUSessionResourceSetupListCxtRes.List) == 0 {
		log.Info("[GNB][NGAP] No PDU Session to set up in InitialContextSetupResponse.")
		return builder
	}
	builder.pdu.PDUSessionResourceSetupListCxtRes = ie.PDUSessionResourceSetupListCxtRes

	return builder
}

func (builder *InitialContextSetupResponseBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}
