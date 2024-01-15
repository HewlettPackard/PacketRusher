/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

func UEOriginatingDeregistration(nasReq *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {

	// Hook for changing AuthenticationResponse behaviour
	hook := fgc.GetNasHook(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration)
	if hook != nil {
		handled, err := hook(nasReq, ueContext, gnb, fgc)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	deregistrationRequest := nasReq.DeregistrationRequestUEOriginatingDeregistration
	context.ReleaseAllPDUSession(ueContext)

	err := state.GetUeFsm().SendEvent(ueContext.GetState(), state.DeregistrationRequested, fsm.ArgsType{}, log.NewEntry(log.StandardLogger()))
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
		msg.SendUEContextReleaseCommand(gnb, ueContext, ngapType.CausePresentNas, ngapType.CauseNasPresentDeregister)
	default:
		return errors.New("[5GC][NAS] Deregistration procedure: unsupported access type")
	}
	return nil
}
