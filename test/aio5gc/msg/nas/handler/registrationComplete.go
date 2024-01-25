/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
)

func RegistrationComplete(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf context.AMFContext) error {
	var err error
	switch ue.GetState().Current() {
	case state.AuthenticationInitiated:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for AuthenticationInitiated UE")
	case state.Deregistrated:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for Deregistrated UE")
	case state.DeregistratedInitiated:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for DeregistratedInitiated UE")
	case state.Registred:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for Registred UE")
	case state.SecurityContextAvailable:
		err = DefaultRegistrationComplete(nasMsg, gnb, ue, amf)
	default:
		err = fmt.Errorf("Unknown UE state: %v ", ue.GetState().ToString())
	}
	return err
}

func DefaultRegistrationComplete(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf context.AMFContext) error {

	nwName := amf.GetNetworkName()
	err := state.UpdateUE(ue.GetStatePointer(), state.Registred)
	if err != nil {
		return err
	}
	msg.SendConfigurationUpdateCommand(gnb, ue, &nwName)
	return nil
}
