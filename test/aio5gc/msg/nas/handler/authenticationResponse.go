/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"encoding/hex"
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/msg"
	"strings"

	"fmt"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func AuthenticationResponse(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {
	var err error
	switch ue.GetState().Current() {
	case state.AuthenticationInitiated:
		err = DefaultAuthenticationResponse(nasMsg, gnb, ue, amf)
	case state.Deregistrated:
		return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for Deregistrated UE")
	case state.DeregistratedInitiated:
		return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for DeregistratedInitiated UE")
	case state.Registred:
		return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for Registred UE")
	case state.SecurityContextAvailable:
		return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for SecurityContextAvailable UE")
	default:
		err = fmt.Errorf("Unknown UE state: %v ", ue.GetState().ToString())
	}
	return err
}

func DefaultAuthenticationResponse(nasMsg *nas.Message, gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {

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

	err := state.UpdateUE(ue.GetStatePointer(), state.SecurityContextAvailable)
	if err != nil {
		return err
	}
	msg.SendSecurityModeCommand(gnb, ue)
	return nil
}
