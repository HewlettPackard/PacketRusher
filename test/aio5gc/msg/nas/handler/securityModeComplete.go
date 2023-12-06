/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
)

func SecurityModeComplete(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	securityModeComplete := nasReq.SecurityModeComplete
	if securityModeComplete.IMEISV != nil {
		if pei, err := nasConvert.PeiToStringWithError(securityModeComplete.IMEISV.Octet[:]); err != nil {
			return fmt.Errorf("[5GC][NAS] Decode PEI failed: %w", err)
		} else {
			ue.SetPei(pei)
		}
	}

	if securityModeComplete.NASMessageContainer == nil {
		return fmt.Errorf("[5GC][NAS] Empty NASMessageContainer in securityModeComplete message")
	}
	contents := securityModeComplete.NASMessageContainer.GetNASMessageContainerContents()
	m := nas.NewMessage()
	if err := m.GmmMessageDecode(&contents); err != nil {
		return err
	}

	switch m.GmmMessage.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		registrationRequest := m.GmmMessage.RegistrationRequest
		ue.SetSecurityCapability(registrationRequest.UESecurityCapability)
		ue.AllocateGuti(amf)
		ue.GetSecurityContext().UpdateSecurityContext()
		msg.SendRegistrationAccept(gnb, ue, amf)
	default:
		return errors.New("nas message container Iei type error")
	}
	return nil
}
