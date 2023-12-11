/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
)

func RegistrationComplete(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf context.AMFContext) error {
	if !ue.GetInitialContextSetup() {
		return errors.New("[5GC][NGAP] This UE has no security context set up")
	}
	nwName := amf.GetNetworkName()
	msg.SendConfigurationUpdateCommand(gnb, ue, &nwName)
	return nil
}
