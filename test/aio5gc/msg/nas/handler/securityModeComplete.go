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

	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
)

func SecurityModeComplete(nasReq *nas.SecModeComplete, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.Authenticated:
		err = DefaultSecurityModeComplete(nasReq, ue, gnb, amf)
	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for SecurityModeComplete", ue.GetState().Current())
	}
	return err
}

func DefaultSecurityModeComplete(nasReq *nas.SecModeComplete, ue *context.UEContext, gnb *context.GNBContext, amf *context.AMFContext) error {

	securityModeComplete := nasReq
	if securityModeComplete.IMEISV != nil {
		if pei, err := peiString(securityModeComplete.IMEISV); err != nil {
			return fmt.Errorf("[5GC][NAS] Decode PEI failed: %w", err)
		} else {
			ue.SetPei(pei)
		}
	}

	if securityModeComplete.NASMsgCntr == nil {
		return fmt.Errorf("[5GC][NAS] Empty NASMsgCntr in securityModeComplete message")
	}
	contents := securityModeComplete.NASMsgCntr.Contents
	m, err := nas.Parse(contents, nil)
	if err != nil {
		return err
	}

	switch m.MsgType() {
	case nas.MsgTypeRegReq:
		registrationRequest := m.(*nas.RegReq)
		ue.SetSecurityCapability(registrationRequest.UESecCapability)
		ue.AllocateGuti(amf)
		ue.GetSecurityContext().UpdateSecurityContext()
		msg.SendRegistrationAccept(gnb, ue, amf)
	default:
		return errors.New("nas message container Iei type error")
	}
	return nil
}

func peiString(identity *ie.MobileId5GS) (string, error) { return identity.PEIStr(), nil }
