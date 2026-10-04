/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	ie "github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	ngapType "github.com/free5gc/ngap/ie"
	"github.com/free5gc/util/fsm"
	"github.com/sirupsen/logrus"
)

func UEOriginatingDeregistration(nasReq *nas.DeregReqUEOrig, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.AuthenticationInitiated:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	case context.Registered:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	case context.Authenticated:
		err = DefaultUEOriginatingDeregistration(nasReq, amf, ue, gnb)
	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for UEOriginatingDeregistration", ue.GetState().Current())
	}
	return err
}

func DefaultUEOriginatingDeregistration(nasReq *nas.DeregReqUEOrig, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {

	deregistrationRequest := nasReq
	context.ForceReleaseAllPDUSession(ue)

	err := ue.GetUeFsm().SendEvent(ue.GetState(), context.DeregistrationRequest, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger()))

	if err != nil {
		return err
	}
	// if Deregistration type is not switch-off, send Deregistration Accept (need to implement)
	if !deregistrationRequest.DeregType.Switchoff {
		return errors.New("[5GC][NAS] Not switch-off deregistration not supported")
	}

	// TS 23.502 4.2.6, 4.12.3
	switch deregistrationRequest.DeregType.AccessType {
	case ie.AccessType_3gpp:
		msg.SendUEContextReleaseCommand(gnb, ue, 3, ngapType.CauseNasPresentDeregister)
	default:
		return errors.New("[5GC][NAS] Deregistration procedure: unsupported access type")
	}
	return nil
}
