/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func UEOriginatingDeregistration(nasReq *nas.Message, amf *context.AMFContext, ueContext *context.UEContext, gnb *context.GNBContext) error {
	deregistrationRequest := nasReq.DeregistrationRequestUEOriginatingDeregistration
	if !ueContext.GetInitialContextSetup() {
		return errors.New("[5GC][NGAP] This UE has no security context set up")
	}
	context.ReleaseAllPDUSession(ueContext)

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
