/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"encoding/hex"
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/tools"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

func PDUSessionEstablishmentAccept(ue *context.UEContext, smContext *context.SmContext) (msg []byte, err error) {

	nasMsg, err := buildSessionEstablishmentAccept(ue, smContext)
	if err != nil {
		return nil, err
	}
	nasPdu, err := nasMsg.PlainNasEncode()
	if err != nil {
		return nil, err
	}
	nasMsg, err = buildDLNASTransport(ue, nasPdu, uint8(smContext.GetPduSessionId()))
	if err != nil {
		return nil, err
	}
	return tools.Encode(ue, nasMsg)
}

func buildDLNASTransport(ue *context.UEContext, nasPdu []byte, pduSessionId uint8) (msg *nas.Message, err error) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeDLNASTransport)

	m.SecurityHeader = nas.SecurityHeader{
		ProtocolDiscriminator: nasMessage.Epd5GSMobilityManagementMessage,
		SecurityHeaderType:    nas.SecurityHeaderTypeIntegrityProtectedAndCiphered,
	}

	dLNASTransport := nasMessage.NewDLNASTransport(0)
	dLNASTransport.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	dLNASTransport.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	dLNASTransport.SetMessageType(nas.MsgTypeDLNASTransport)
	dLNASTransport.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(nasMessage.PayloadContainerTypeN1SMInfo)
	dLNASTransport.PayloadContainer.SetLen(uint16(len(nasPdu)))
	dLNASTransport.PayloadContainer.SetPayloadContainerContents(nasPdu)
	if pduSessionId != 0 {
		dLNASTransport.PduSessionID2Value = new(nasType.PduSessionID2Value)
		dLNASTransport.PduSessionID2Value.SetIei(nasMessage.DLNASTransportPduSessionID2ValueType)
		dLNASTransport.PduSessionID2Value.SetPduSessionID2Value(pduSessionId)
	}

	m.GmmMessage.DLNASTransport = dLNASTransport

	return m, nil
}

func buildSessionEstablishmentAccept(ue *context.UEContext, smContext *context.SmContext) (msg *nas.Message, err error) {
	m := nas.NewMessage()
	m.GsmMessage = nas.NewGsmMessage()
	m.GsmHeader.SetMessageType(nas.MsgTypePDUSessionEstablishmentAccept)
	m.GsmHeader.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSSessionManagementMessage)
	m.PDUSessionEstablishmentAccept = nasMessage.NewPDUSessionEstablishmentAccept(0x0)
	pDUSessionEstablishmentAccept := m.PDUSessionEstablishmentAccept

	sessRule := smContext.GetSessionRule()
	authDefQos := sessRule.AuthDefQos

	pDUSessionEstablishmentAccept.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSSessionManagementMessage)
	pDUSessionEstablishmentAccept.SetPDUSessionID(uint8(smContext.GetPduSessionId()))
	pDUSessionEstablishmentAccept.SetPTI(smContext.GetPti())
	pDUSessionEstablishmentAccept.SetMessageType(nas.MsgTypePDUSessionEstablishmentAccept)
	pDUSessionEstablishmentAccept.SetSSCMode(1)

	pDUSessionEstablishmentAccept.SetPDUSessionType(smContext.GetPduSessionType())

	qoSRules := nasType.QoSRules{
		{
			Identifier: 0x01,
			DQR:        true,
			Operation:  nasType.OperationCodeCreateNewQoSRule,
			Precedence: 255,
			QFI:        smContext.GetDefQosQFI(),
			PacketFilterList: nasType.PacketFilterList{
				{
					Identifier: 1,
					Direction:  nasType.PacketFilterDirectionBidirectional,
					Components: nasType.PacketFilterComponentList{
						&nasType.PacketFilterMatchAll{},
					},
				},
			},
		},
	}

	qosRulesBytes, err := qoSRules.MarshalBinary()
	if err != nil {
		return nil, errors.New("[5GC] Error while encoding qoSRules: " + err.Error())
	}

	pDUSessionEstablishmentAccept.AuthorizedQosRules.SetLen(uint16(len(qosRulesBytes)))
	pDUSessionEstablishmentAccept.AuthorizedQosRules.SetQosRule(qosRulesBytes)

	pDUSessionEstablishmentAccept.SessionAMBR = nasConvert.ModelsToSessionAMBR(sessRule.AuthSessAmbr)
	pDUSessionEstablishmentAccept.SessionAMBR.SetLen(uint8(len(pDUSessionEstablishmentAccept.SessionAMBR.Octet)))

	addr, addrLen := smContext.PDUAddressToNAS()
	pDUSessionEstablishmentAccept.PDUAddress = nasType.
		NewPDUAddress(nasMessage.PDUSessionEstablishmentAcceptPDUAddressType)
	pDUSessionEstablishmentAccept.PDUAddress.SetLen(addrLen)
	pDUSessionEstablishmentAccept.PDUAddress.SetPDUSessionTypeValue(smContext.GetPduSessionType())
	pDUSessionEstablishmentAccept.PDUAddress.SetPDUAddressInformation(addr)

	pDUSessionEstablishmentAccept.SNSSAI = nasType.NewSNSSAI(nasMessage.ULNASTransportSNSSAIType)
	nssai := smContext.GetSnnsai()
	var sd [3]uint8
	if byteArray, err := hex.DecodeString(nssai.Sd); err != nil {
		return nil, errors.New("[5GC] error while decoding nssai sd: " + err.Error())
	} else {
		copy(sd[:], byteArray)
	}
	pDUSessionEstablishmentAccept.SNSSAI.SetLen(4)
	pDUSessionEstablishmentAccept.SNSSAI.SetSST(uint8(nssai.Sst))
	pDUSessionEstablishmentAccept.SNSSAI.SetSD(sd)

	authDescs := nasType.QoSFlowDescs{}
	defaultAuthDesc := nasType.QoSFlowDesc{}
	defaultAuthDesc.QFI = smContext.GetDefQosQFI()
	defaultAuthDesc.OperationCode = nasType.OperationCodeCreateNewQoSFlowDescription
	parameter := new(nasType.QoSFlow5QI)
	parameter.FiveQI = uint8(authDefQos.Var5qi)
	defaultAuthDesc.Parameters = append(defaultAuthDesc.Parameters, parameter)
	authDescs = append(authDescs, defaultAuthDesc)

	qosDescBytes, err := authDescs.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pDUSessionEstablishmentAccept.AuthorizedQosFlowDescriptions = nasType.
		NewAuthorizedQosFlowDescriptions(nasMessage.PDUSessionEstablishmentAcceptAuthorizedQosFlowDescriptionsType)
	pDUSessionEstablishmentAccept.AuthorizedQosFlowDescriptions.SetLen(uint16(len(qosDescBytes)))
	pDUSessionEstablishmentAccept.SetQoSFlowDescriptions(qosDescBytes)

	if smContext.ProtocolConfigurationOptions.DNSIPv4Request ||
		smContext.ProtocolConfigurationOptions.DNSIPv6Request ||
		smContext.ProtocolConfigurationOptions.PCSCFIPv4Request ||
		smContext.ProtocolConfigurationOptions.IPv4LinkMTURequest {
		pDUSessionEstablishmentAccept.ExtendedProtocolConfigurationOptions = nasType.NewExtendedProtocolConfigurationOptions(
			nasMessage.PDUSessionEstablishmentAcceptExtendedProtocolConfigurationOptionsType,
		)
		protocolConfigurationOptions := nasConvert.NewProtocolConfigurationOptions()

		// IPv4 DNS
		if smContext.ProtocolConfigurationOptions.DNSIPv4Request {
			err := protocolConfigurationOptions.AddDNSServerIPv4Address(smContext.GetDataNetwork().Dns.IPv4Addr)
			if err != nil {
				return nil, errors.New("[5GC][NAS] Error while adding DNS IPv4 Addr: " + err.Error())
			}
		}

		// IPv6 DNS
		if smContext.ProtocolConfigurationOptions.DNSIPv6Request {
			err := protocolConfigurationOptions.AddDNSServerIPv6Address(smContext.GetDataNetwork().Dns.IPv6Addr)
			if err != nil {
				return nil, errors.New("[5GC][NAS] Error while adding DNS IPv6 Addr: " + err.Error())
			}
		}

		pcoContents := protocolConfigurationOptions.Marshal()
		pcoContentsLength := len(pcoContents)
		pDUSessionEstablishmentAccept.
			ExtendedProtocolConfigurationOptions.
			SetLen(uint16(pcoContentsLength))
		pDUSessionEstablishmentAccept.
			ExtendedProtocolConfigurationOptions.
			SetExtendedProtocolConfigurationOptionsContents(pcoContents)
	}

	pDUSessionEstablishmentAccept.DNN = nasType.NewDNN(nasMessage.ULNASTransportDNNType)
	pDUSessionEstablishmentAccept.DNN.SetDNN(smContext.GetDataNetwork().Dnn)

	return m, nil
}
