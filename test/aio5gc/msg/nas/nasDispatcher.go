/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	"my5G-RANTester/test/aio5gc/lib/tools"
	nasHandler "my5G-RANTester/test/aio5gc/msg/nas/handler"
	"strconv"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func Dispatch(nasPDU *ngapType.NASPDU, ueContext *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) {
	payload := nasPDU.Value
	m := new(nas.Message)
	m.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & 0x0f
	var msg *nas.Message
	var err error

	if m.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {
		var integrityProtected bool
		msg, integrityProtected, err = tools.Decode(ueContext, payload, false)
		if !integrityProtected {
			log.Fatal(errors.New("[5GC][NAS] message integrity could not be verified:" + err.Error()))
		}

	} else {
		msg, err = tools.DecodePlainNasNoIntegrityCheck(payload)
	}
	switch ueContext.GetState().Current() {
	case state.Deregistrated:
		err = dispatchDeregistreted(msg, ueContext, fgc, gnb)
	case state.CommonProcedureInitiated:
		err = dispatchCommonProcedureInitiated(msg, ueContext, fgc, gnb)
	case state.Registred:
		err = dispatchRegistred(msg, ueContext, fgc, gnb)
	case state.DeregistratedInitiated:
		err = dispatchDeregistratedInitiated(msg, ueContext, fgc, gnb)
	default:
		err = errors.New("[5GC][NAS] ue in unknown state: " + string(ueContext.GetState().Current()))
	}

	if err != nil {
		log.Error(err)
	}
}

func dispatchDeregistreted(msg *nas.Message, ueContext *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) error {
	var err error

	// Default when ue is deregistrated
	switch msg.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		log.Info("[5GC][NAS] Received Registration Request")
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.RegistrationRequest(msg, ueContext, gnb, fgc)

	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	return err
}

func dispatchCommonProcedureInitiated(msg *nas.Message, ueContext *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) error {
	var err error

	// Dispacther when ue is in common procedure initiated state
	switch msg.GmmHeader.GetMessageType() {

	case nas.MsgTypeIdentityResponse:
		log.Info("[5GC][NAS] Received Identity Response")
		err = nasHandler.IdentityResponse(msg, fgc, ueContext, gnb)

	case nas.MsgTypeAuthenticationResponse:
		log.Info("[5GC][NAS] Received Authentication Response")
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.AuthenticationResponse(msg, gnb, ueContext, fgc)

	case nas.MsgTypeSecurityModeComplete:
		log.Info("[5GC][NAS] Received Security Mode Complete")
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.SecurityModeComplete(msg, ueContext, gnb, fgc)

	case nas.MsgTypeRegistrationComplete:
		log.Info("[5GC][NAS] Received Registration Complete")
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.RegistrationComplete(msg, gnb, ueContext, fgc)

	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	return err
}

func dispatchRegistred(msg *nas.Message, ueContext *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) error {
	var err error
	session := fgc.GetSessionContext()

	// Dispacther when ue is in registrated state
	switch msg.GmmHeader.GetMessageType() {

	case nas.MsgTypeULNASTransport:
		log.Info("[5GC][NAS] Received UL NAS Transport")
		err = nasHandler.UlNasTransport(msg, gnb, ueContext, session)

	case nas.MsgTypeConfigurationUpdateComplete:
		log.Info("[5GC][NAS] Received Configuration Update Complete")

	case nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration:
		log.Info("[5GC][NAS] Received Deregistration Request: UE Originating Deregistration")
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.UEOriginatingDeregistration(msg, ueContext, gnb, fgc)

	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	return err
}

func dispatchDeregistratedInitiated(msg *nas.Message, ueContext *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) error {
	var err error

	// for  NW initiated Dereg
	switch msg.GmmHeader.GetMessageType() {
	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	return err
}
