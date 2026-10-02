/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	ngapConvert "my5G-RANTester/lib/ngap"

	ngap "github.com/free5gc/ngap/message"

	ngapType "github.com/free5gc/ngap/ie"
)

type HandoverNotifyBuilder struct {
	pdu *ngap.HandoverNotify
}

func HandoverNotify(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewHandoverNotifyBuilder().
		SetAmfUeNgapId(ue.GetAmfUeId()).SetRanUeNgapId(ue.GetRanUeId()).
		SetUserLocation(gnb).
		Build()
}

func NewHandoverNotifyBuilder() *HandoverNotifyBuilder {
	pdu := &ngap.HandoverNotify{}
	return &HandoverNotifyBuilder{pdu}
}

func (builder *HandoverNotifyBuilder) SetAmfUeNgapId(amfUeNgapID int64) *HandoverNotifyBuilder {
	builder.pdu.AMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *HandoverNotifyBuilder) SetRanUeNgapId(ranUeNgapID int64) *HandoverNotifyBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *HandoverNotifyBuilder) SetUserLocation(gnb *context.GNBContext) *HandoverNotifyBuilder {
	builder.pdu.UserLocationInformation = ngapConvert.UserLocation(gnb.GetPLMNIdentity(), gnb.GetNRCellIdentity(), gnb.GetTacInBytes())
	return builder
}

func (builder *HandoverNotifyBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}
