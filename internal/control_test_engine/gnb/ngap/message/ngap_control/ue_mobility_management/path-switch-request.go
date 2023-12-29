/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ue_mobility_management

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap"
	log "github.com/sirupsen/logrus"

	"github.com/free5gc/aper"

	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
)

type PathSwitchRequestBuilder struct {
	pdu ngapType.NGAPPDU
	ies *ngapType.ProtocolIEContainerPathSwitchRequestIEs
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
	pdu := ngapType.NGAPPDU{}

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodePathSwitchRequest
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentPathSwitchRequest
	initiatingMessage.Value.PathSwitchRequest = new(ngapType.PathSwitchRequest)

	pathSwitchRequest := initiatingMessage.Value.PathSwitchRequest
	ies := &pathSwitchRequest.ProtocolIEs

	return &PathSwitchRequestBuilder{pdu, ies}
}
func (builder *PathSwitchRequestBuilder) SetSourceAmfUeNgapId(amfUeNgapID int64) *PathSwitchRequestBuilder {
	// SOURCE AMF UE NGAP ID
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDSourceAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentSourceAMFUENGAPID
	ie.Value.SourceAMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.SourceAMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *PathSwitchRequestBuilder) SetRanUeNgapId(ranUeNgapID int64) *PathSwitchRequestBuilder {
	// RAN UE NGAP ID
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *PathSwitchRequestBuilder) SetUserLocation(gnb *context.GNBContext) *PathSwitchRequestBuilder {
	// User Location Information
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUserLocationInformation
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentUserLocationInformation
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

func (builder *PathSwitchRequestBuilder) SetUserSecurityCapabilities(userSecurityCapabilities *ngapType.UESecurityCapabilities) *PathSwitchRequestBuilder {
	// User Location Information
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUESecurityCapabilities
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentUESecurityCapabilities
	ie.Value.UESecurityCapabilities = userSecurityCapabilities

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *PathSwitchRequestBuilder) PathSwitchRequestTransfer(gnbN3Ip string, pduSessions [16]*context.GnbPDUSession) *PathSwitchRequestBuilder {
	// PDUSessionResourceToBeSwitchedDLList
	ie := ngapType.PathSwitchRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDPDUSessionResourceToBeSwitchedDLList
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.PathSwitchRequestIEsPresentPDUSessionResourceToBeSwitchedDLList
	ie.Value.PDUSessionResourceToBeSwitchedDLList = new(ngapType.PDUSessionResourceToBeSwitchedDLList)

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		item := ngapType.PDUSessionResourceToBeSwitchedDLItem{PDUSessionID: ngapType.PDUSessionID{Value: pduSession.GetPduSessionId()}}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, pduSession.GetTeidDownlink())
		transfer := ngapType.PathSwitchRequestTransfer{
			DLNGUUPTNLInformation: ngapType.UPTransportLayerInformation{
				Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
				GTPTunnel: &ngapType.GTPTunnel{
					TransportLayerAddress: ngapConvert.IPAddressToNgap(gnbN3Ip, ""),
					GTPTEID:               ngapType.GTPTEID{Value: aper.OctetString(buf.Bytes())},
					IEExtensions:          nil,
				}},
			DLNGUTNLInformationReused:    nil,
			UserPlaneSecurityInformation: nil,
			QosFlowAcceptedList:          ngapType.QosFlowAcceptedList{List: []ngapType.QosFlowAcceptedItem{{QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: pduSession.GetQosId()}}}},
			IEExtensions:                 nil,
		}
		res, _ := aper.MarshalWithParams(transfer, "valueExt")
		item.PathSwitchRequestTransfer = res

		ie.Value.PDUSessionResourceToBeSwitchedDLList.List = append(ie.Value.PDUSessionResourceToBeSwitchedDLList.List, item)
	}
	if len(ie.Value.PDUSessionResourceToBeSwitchedDLList.List) == 0 {
		log.Error("[GNB][NGAP] No PDU Session to hand over. Xn Handover requires at least a PDU Session.")
		return builder
	}

	builder.ies.List = append(builder.ies.List, ie)

	return builder
}

func (builder *PathSwitchRequestBuilder) Build() ([]byte, error) {
	return ngap.Encoder(builder.pdu)
}
