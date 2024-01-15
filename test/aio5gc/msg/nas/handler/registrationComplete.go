/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
)

func RegistrationComplete(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, fgc *context.Aio5gc) error {

	// Hook for changing RegistrationComplete behaviour
	hook := fgc.GetNasHook(nas.MsgTypeRegistrationComplete)
	if hook != nil {
		handled, err := hook(nasMsg, ue, gnb, fgc)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}
	nwName := fgc.GetAMFContext().GetNetworkName()
	err := state.GetUeFsm().SendEvent(ue.GetState(), state.InitialRegistrationAccepted, fsm.ArgsType{}, log.NewEntry(log.StandardLogger()))
	if err != nil {
		return err
	}
	msg.SendConfigurationUpdateCommand(gnb, ue, &nwName)
	return nil
}
