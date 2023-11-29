/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nasMsgHandler

import (
	"fmt"
	"my5G-RANTester/test/amf/context"
	"my5G-RANTester/test/amf/lib/tools"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/openapi/models"
)

func RegistrationAccept(ue *context.UEContext) (nasPdu []byte, err error) {
	nasMsg, err := buildRegistrationAccept(ue)
	if err != nil {
		return nil, err
	}

	return tools.Encode(ue, nasMsg)
}

func buildRegistrationAccept(ue *context.UEContext) (nasMsg *nas.Message, err error) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeRegistrationAccept)

	m.SecurityHeader = nas.SecurityHeader{
		ProtocolDiscriminator: nasMessage.Epd5GSMobilityManagementMessage,
		SecurityHeaderType:    nas.SecurityHeaderTypeIntegrityProtectedAndCiphered,
	}

	registrationAccept := nasMessage.NewRegistrationAccept(0)
	registrationAccept.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	registrationAccept.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	registrationAccept.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	registrationAccept.RegistrationAcceptMessageIdentity.SetMessageType(nas.MsgTypeRegistrationAccept)

	registrationAccept.RegistrationResult5GS.SetLen(1)
	registrationResult := nasMessage.AccessType3GPP

	registrationAccept.RegistrationResult5GS.SetRegistrationResultValue5GS(registrationResult)

	gutiNas, err := nasConvert.GutiToNasWithError(ue.GetGuti())
	if err != nil {
		return nil, fmt.Errorf("encode GUTI failed: %w", err)
	}
	registrationAccept.GUTI5G = &gutiNas
	registrationAccept.GUTI5G.SetIei(nasMessage.RegistrationAcceptGUTI5GType)

	// Supporting only one TAI per UE for now
	registrationAccept.TAIList = nasType.NewTAIList(nasMessage.RegistrationAcceptTAIListType)
	var taiList []models.Tai
	taiList = append(taiList, *ue.GetUserLocationInfo().Tai)
	taiListNas := nasConvert.TaiListToNas(taiList)
	registrationAccept.TAIList.SetLen(uint8(len(taiListNas)))
	registrationAccept.TAIList.SetPartialTrackingAreaIdentityList(taiListNas)

	nssai := ue.GetNssai()
	var buf []uint8
	buf = append(buf, nasConvert.SnssaiToNas(nssai)...)

	registrationAccept.AllowedNSSAI = nasType.NewAllowedNSSAI(nasMessage.RegistrationAcceptAllowedNSSAIType)
	registrationAccept.AllowedNSSAI.SetLen(uint8(len(buf)))
	registrationAccept.AllowedNSSAI.SetSNSSAIValue(buf)
	registrationAccept.ConfiguredNSSAI = nasType.NewConfiguredNSSAI(nasMessage.RegistrationAcceptConfiguredNSSAIType)
	registrationAccept.ConfiguredNSSAI.SetLen(uint8(len(buf)))
	registrationAccept.ConfiguredNSSAI.SetSNSSAIValue(buf)

	// 5gs network feature support
	registrationAccept.NetworkFeatureSupport5GS = nasType.
		NewNetworkFeatureSupport5GS(nasMessage.RegistrationAcceptNetworkFeatureSupport5GSType)
	registrationAccept.NetworkFeatureSupport5GS.SetLen(2)

	registrationAccept.SetMPSI(0)
	registrationAccept.SetIWKN26(0)
	registrationAccept.SetEMF(0)
	registrationAccept.SetEMC(0)
	registrationAccept.SetIMSVoPS3GPP(1)
	registrationAccept.SetIMSVoPSN3GPP(0)
	registrationAccept.SetEMCN(0)
	registrationAccept.SetMCSI(0)

	registrationAccept.T3512Value = nasType.NewT3512Value(nasMessage.RegistrationAcceptT3512ValueType)
	registrationAccept.T3512Value.SetLen(1)
	t3512 := nasConvert.GPRSTimer3ToNas(60)
	registrationAccept.T3512Value.Octet = t3512

	m.GmmMessage.RegistrationAccept = registrationAccept

	return m, nil
}
