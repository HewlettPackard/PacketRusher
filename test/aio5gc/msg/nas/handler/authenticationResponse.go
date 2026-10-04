/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"encoding/hex"
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"
	"strings"

	"fmt"

	nas "github.com/free5gc/nas/message"
	ngapType "github.com/free5gc/ngap/ie"
	"github.com/free5gc/util/fsm"
	"github.com/sirupsen/logrus"
	log "my5G-RANTester/internal/log"
)

func AuthenticationResponse(nasMsg *nas.AuthRsp, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.AuthenticationInitiated:
		err = DefaultAuthenticationResponse(nasMsg, gnb, ue, amf)
	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for AuthenticationResponse", ue.GetState().Current())
	}
	return err
}

func DefaultAuthenticationResponse(nasMsg *nas.AuthRsp, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {

	if nasMsg.AuthRspParam == nil {
		return errors.New("AuthenticationResponseParameter is nil")
	}
	resStarb := nasMsg.AuthRspParam.Res
	resStar := hex.EncodeToString(resStarb[:])

	xresStar := ue.GetSecurityContext().GetXresStar()

	if strings.EqualFold(resStar, xresStar) {
		log.Info("[5GC] 5G AKA confirmation succeeded")
		ue.DerivateKamf()

		oldUe, err := amf.FindRegisteredUEByMsin(ue.GetSecurityContext().GetMsin())
		if err == nil && oldUe.GetAmfNgapId() != ue.GetAmfNgapId() {
			err := ue.GetUeFsm().SendEvent(oldUe.GetState(), context.ForceDeregistrationInit, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger()))
			if err != nil {
				log.Error(err)
			}
			msg.SendUEContextReleaseCommand(gnb, oldUe, 3, ngapType.CauseNasPresentUnspecified)
		}
	} else {
		return errors.New(("5G AKA confirmation failed, expected res* " + xresStar + " but got " + resStar))
	}
	err := ue.GetUeFsm().SendEvent(ue.GetState(), context.AuthenticationSuccess, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		return err
	}
	prov, err := amf.FindProvisionedData(ue.GetSecurityContext().GetMsin())
	if err != nil {
		return err
	}
	ue.SetDefaultSNssai(prov.GetDefaultSNssai())
	msg.SendSecurityModeCommand(gnb, ue)
	return nil
}
