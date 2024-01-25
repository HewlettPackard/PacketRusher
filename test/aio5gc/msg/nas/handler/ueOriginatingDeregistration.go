/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/ngap/ngapType"
)

func UEOriginatingDeregistration(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	var err error
	switch ue.GetState().Current() {
	case state.AuthenticationInitiated:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	case state.Deregistrated:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received UEOriginatingDeregistration for Deregistrated UE")
	case state.DeregistratedInitiated:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received UEOriginatingDeregistration for DeregistratedInitiated UE")
	case state.Registred:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	case state.SecurityContextAvailable:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	default:
		err = fmt.Errorf("Unknown UE state: %v ", ue.GetState().ToString())
	}
	return err
}

func DefaultUEOriginatingDeregistration(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {

	deregistrationRequest := nasReq.DeregistrationRequestUEOriginatingDeregistration
	context.ReleaseAllPDUSession(ue)

	err := state.UpdateUE(ue.GetStatePointer(), state.DeregistratedInitiated)
	if err != nil {
		return err
	}
	// if Deregistration type is not switch-off, send Deregistration Accept (need to implement)
	if deregistrationRequest.GetSwitchOff() == 0 {
		return errors.New("[5GC][NAS] Not switch-off deregistration not supported")
	}

	// TS 23.502 4.2.6, 4.12.3
	switch deregistrationRequest.GetAccessType() {
	case nasMessage.AccessType3GPP:
		msg.SendUEContextReleaseCommand(gnb, ue, ngapType.CausePresentNas, ngapType.CauseNasPresentDeregister)
	default:
		return errors.New("[5GC][NAS] Deregistration procedure: unsupported access type")
	}
	return nil
}
