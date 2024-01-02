/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap"

	"github.com/free5gc/ngap/ngapType"
)

type HandoverNotifyBuilder struct {
	pdu ngapType.NGAPPDU
	ies *ngapType.ProtocolIEContainerHandoverNotifyIEs
}

func HandoverNotify(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewHandoverNotifyBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetUserLocation(gnb).
		Build()
}

func NewHandoverNotifyBuilder() *HandoverNotifyBuilder {
	pdu := ngapType.NGAPPDU{}

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeHandoverNotification
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentHandoverNotify
	initiatingMessage.Value.HandoverNotify = new(ngapType.HandoverNotify)

	handoverNotify := initiatingMessage.Value.HandoverNotify
	ies := &handoverNotify.ProtocolIEs

	return &HandoverNotifyBuilder{pdu, ies}
}

func (builder *HandoverNotifyBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverNotifyBuilder {
	// AMF UE NGAP ID
	ie := ngapType.HandoverNotifyIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverNotifyBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverNotifyBuilder {
	// RAN UE NGAP ID
	ie := ngapType.HandoverNotifyIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.HandoverRequiredIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverNotifyBuilder) SetUserLocation(gnb *context.GNBContext) *HandoverNotifyBuilder {
	// User Location Information
	ie := ngapType.HandoverNotifyIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUserLocationInformation
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.HandoverNotifyIEsPresentUserLocationInformation
	ie.Value.UserLocationInformation = new(ngapType.UserLocationInformation)

	userLocationInformation := ie.Value.UserLocationInformation
	userLocationInformation.Present = ngapType.UserLocationInformationPresentUserLocationInformationNR
	userLocationInformation.UserLocationInformationNR = new(ngapType.UserLocationInformationNR)

	userLocationInformationNR := userLocationInformation.UserLocationInformationNR
	userLocationInformationNR.NRCGI.PLMNIdentity = gnb.GetPLMNIdentity()
	userLocationInformationNR.NRCGI.NRCellIdentity = gnb.GetNRCellIdentity()

	userLocationInformationNR.TAI.PLMNIdentity = gnb.GetPLMNIdentity()
	userLocationInformationNR.TAI.TAC.Value = gnb.GetTacInBytes()

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *HandoverNotifyBuilder) Build() ([]byte, error) {
	return ngap.Encoder(builder.pdu)
}
