/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/util/fsm"
	"github.com/sirupsen/logrus"
)

func RegistrationComplete(nasMsg *nas.RegComplete, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.Authenticated:
		err = DefaultRegistrationComplete(nasMsg, gnb, ue, amf)
	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for RegistrationComplete", ue.GetState().Current())
	}
	return err
}

func DefaultRegistrationComplete(nasMsg *nas.RegComplete, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {

	nwName := amf.GetNetworkName()
	err := ue.GetUeFsm().SendEvent(ue.GetState(), context.RegistrationAccept, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		return err
	}
	msg.SendConfigurationUpdateCommand(gnb, ue, &nwName)
	return nil
}
