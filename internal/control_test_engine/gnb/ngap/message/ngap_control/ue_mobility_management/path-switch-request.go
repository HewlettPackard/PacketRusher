/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ue_mobility_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"

	ngap "github.com/free5gc/ngap/message"
	log "github.com/sirupsen/logrus"

	ngapType "github.com/free5gc/ngap/ie"
	ngapConvert "my5G-RANTester/lib/ngap"
)

type PathSwitchRequestBuilder struct {
	pdu *ngap.PathSwitchRequest
}

func PathSwitchRequest(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	return NewPathSwitchRequestBuilder().
		SetSourceAmfUeNgapId(ue.GetAmfUeId()).
		SetRanUeNgapId(ue.GetRanUeId()).
		PathSwitchRequestTransfer(gnb.GetN3GnbIp(), ue.GetPduSessions()).
		SetUserLocation(gnb).
		SetUserSecurityCapabilities(ue.GetUESecurityCapabilities()).
		Build()
}

func NewPathSwitchRequestBuilder() *PathSwitchRequestBuilder {
	pdu := &ngap.PathSwitchRequest{}
	return &PathSwitchRequestBuilder{pdu}
}
func (builder *PathSwitchRequestBuilder) SetSourceAmfUeNgapId(amfUeNgapID int64) *PathSwitchRequestBuilder {
	builder.pdu.SourceAMFUENGAPID = &ngapType.AMFUENGAPID{Value: amfUeNgapID}
	return builder
}

func (builder *PathSwitchRequestBuilder) SetRanUeNgapId(ranUeNgapID int64) *PathSwitchRequestBuilder {
	builder.pdu.RANUENGAPID = &ngapType.RANUENGAPID{Value: ranUeNgapID}
	return builder
}

func (builder *PathSwitchRequestBuilder) SetUserLocation(gnb *context.GNBContext) *PathSwitchRequestBuilder {
	builder.pdu.UserLocationInformation = ngapConvert.UserLocation(gnb.GetPLMNIdentity(), gnb.GetNRCellIdentity(), gnb.GetTacInBytes())
	return builder
}

func (builder *PathSwitchRequestBuilder) SetUserSecurityCapabilities(userSecurityCapabilities *ngapType.UESecurityCapabilities) *PathSwitchRequestBuilder {
	builder.pdu.UESecurityCapabilities = userSecurityCapabilities
	return builder
}

func (builder *PathSwitchRequestBuilder) PathSwitchRequestTransfer(gnbN3Ip netip.Addr, pduSessions [16]*context.GnbPDUSession) *PathSwitchRequestBuilder {
	list := &ngapType.PDUSessionResourceToBeSwitchedDLList{}
	for _, session := range pduSessions {
		if session == nil {
			continue
		}
		transfer := &ngapType.PathSwitchRequestTransfer{
			DLNGUUPTNLInformation: ngapConvert.UPTransport(gnbN3Ip, session.GetTeidDownlink()),
			QosFlowAcceptedList:   &ngapType.QosFlowAcceptedList{List: []ngapType.QosFlowAcceptedItem{{QosFlowIdentifier: &ngapType.QosFlowIdentifier{Value: session.GetQosId()}}}},
		}
		encoded, err := ngapConvert.Marshal(transfer)
		if err != nil {
			log.Errorf("[GNB][NGAP] Unable to encode path switch transfer: %v", err)
			continue
		}
		list.List = append(list.List, ngapType.PDUSessionResourceToBeSwitchedDLItem{PDUSessionID: &ngapType.PDUSessionID{Value: session.GetPduSessionId()}, PathSwitchRequestTransfer: ngapConvert.Octets(encoded)})
	}
	if len(list.List) > 0 {
		builder.pdu.PDUSessionResourceToBeSwitchedDLList = list
	}
	return builder
}

func (builder *PathSwitchRequestBuilder) Build() ([]byte, error) {
	return builder.pdu.MarshalBinary()
}
