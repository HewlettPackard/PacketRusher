/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"encoding/hex"
	"errors"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"
	"strings"

	"github.com/free5gc/nas"

	log "github.com/sirupsen/logrus"
)

func AuthenticationResponse(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {
	if nasMsg.AuthenticationResponse.AuthenticationResponseParameter == nil {
		return errors.New("AuthenticationResponseParameter is nil")
	}
	resStarb := nasMsg.AuthenticationResponse.AuthenticationResponseParameter.GetRES()
	resStar := hex.EncodeToString(resStarb[:])

	xresStar := ue.GetSecurityContext().GetXresStar()

	if strings.EqualFold(resStar, xresStar) {
		log.Info("[5GC] 5G AKA confirmation succeeded")
		ue.DerivateKamf()
		oldUe, err := amf.FindRegistredUEByMsin(ue.GetSecurityContext().GetMsin())
		if err == nil && oldUe.GetAmfNgapId() != ue.GetAmfNgapId() {
			msg.SendUEContextReleaseCommand(gnb, oldUe, ngapType.CausePresentNas, ngapType.CauseNasPresentUnspecified)
		}
	} else {
		return errors.New(("5G AKA confirmation failed, expected res* " + xresStar + " but got " + resStar))
	}

	msg.SendSecurityModeCommand(gnb, ue)
	return nil
}
