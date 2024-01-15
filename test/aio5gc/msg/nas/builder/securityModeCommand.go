/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"bytes"
	"fmt"
	"my5G-RANTester/internal/common/auth"
	"my5G-RANTester/test/aio5gc/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

func SecurityModeCommand(ue *context.UEContext) ([]byte, error) {

	integAlg, cipherAlg := auth.SelectAlgorithms(ue.GetSecurityCapability())
	ue.GetSecurityContext().SetCipheringAlg(cipherAlg)
	ue.GetSecurityContext().SetIntegrityAlg(integAlg)

	ue.GetSecurityContext().DerivateAlgKey()
	nasMsg := buildSecurityModeCommand(ue)
	data := new(bytes.Buffer)
	err := nasMsg.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	msg := data.Bytes()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func buildSecurityModeCommand(ue *context.UEContext) *nas.Message {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeSecurityModeCommand)

	securityModeCommand := nasMessage.NewSecurityModeCommand(0)
	securityModeCommand.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	securityModeCommand.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	securityModeCommand.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	securityModeCommand.SecurityModeCommandMessageIdentity.SetMessageType(nas.MsgTypeSecurityModeCommand)

	securityModeCommand.SelectedNASSecurityAlgorithms.SetTypeOfIntegrityProtectionAlgorithm(ue.GetSecurityContext().GetIntegrityAlg())
	securityModeCommand.SelectedNASSecurityAlgorithms.SetTypeOfCipheringAlgorithm(ue.GetSecurityContext().GetCipheringAlg())

	securityModeCommand.SpareHalfOctetAndNgksi = nasConvert.SpareHalfOctetAndNgksiToNas(ue.GetNgKsi())

	securityModeCommand.ReplayedUESecurityCapabilities.SetLen(ue.GetSecurityCapability().GetLen())
	securityModeCommand.ReplayedUESecurityCapabilities.Buffer = ue.GetSecurityCapability().Buffer

	var isIMEISVRequested uint8
	if ue.GetPei() != "" {
		isIMEISVRequested = nasMessage.IMEISVNotRequested
	} else {
		isIMEISVRequested = nasMessage.IMEISVRequested
	}

	securityModeCommand.IMEISVRequest = nasType.NewIMEISVRequest(nasMessage.SecurityModeCommandIMEISVRequestType)
	securityModeCommand.IMEISVRequest.SetIMEISVRequestValue(isIMEISVRequested)

	securityModeCommand.Additional5GSecurityInformation = nasType.
		NewAdditional5GSecurityInformation(nasMessage.SecurityModeCommandAdditional5GSecurityInformationType)
	securityModeCommand.Additional5GSecurityInformation.SetLen(1)
	securityModeCommand.Additional5GSecurityInformation.SetRINMR(1)
	securityModeCommand.Additional5GSecurityInformation.SetHDP(0)

	m.GmmMessage.SecurityModeCommand = securityModeCommand

	return m
}
