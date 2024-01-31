/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
	nasHandler "my5G-RANTester/test/aio5gc/msg/nas/handler"
	"strconv"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func Dispatch(nasPDU *ngapType.NASPDU, ue *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) {
	payload := nasPDU.Value
	m := new(nas.Message)
	m.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & 0x0f
	var msg *nas.Message
	var err error

	switch ue.GetState().Current() {
	case context.Authenticated,
		context.Registered:
		if m.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {
			var integrityProtected bool
			msg, integrityProtected, err = codec.Decode(ue, payload, false)
			if !integrityProtected {
				log.Error("[5GC][NAS] message integrity could not be verified:" + err.Error())
				return
			}
		} else {
			log.Error("[5GC][NAS] Received plain Nas message UE in state" + ue.GetState().Current())
			return
		}

	default:
		if m.SecurityHeaderType == nas.SecurityHeaderTypePlainNas {
			msg, err = codec.DecodePlainNasNoIntegrityCheck(payload)
		} else {
			log.Error("[5GC][NAS] Received non plain Nas message UE in state" + ue.GetState().Current())
			return
		}
	}

	// Hook for changing NASHandler behaviour
	hook := fgc.GetNasHook(msg.GmmHeader.GetMessageType())
	if hook != nil {
		handled, err := hook(msg, ue, gnb, fgc)
		if err != nil {
			log.Error(err)
		}
		if handled {
			return
		}
	}

	amf := fgc.GetAMFContext()
	session := fgc.GetSessionContext()

	// Default Dispacther
	switch msg.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		log.Info("[5GC][NAS] Received Registration Request")
		err = nasHandler.RegistrationRequest(msg, amf, ue, gnb)

	case nas.MsgTypeIdentityResponse:
		log.Info("[5GC][NAS] Received Identity Response")
		err = nasHandler.IdentityResponse(msg, amf, ue, gnb)

	case nas.MsgTypeAuthenticationResponse:
		log.Info("[5GC][NAS] Received Authentication Response")
		err = nasHandler.AuthenticationResponse(msg, gnb, ue, amf)

	case nas.MsgTypeSecurityModeComplete:
		log.Info("[5GC][NAS] Received Security Mode Complete")
		err = nasHandler.SecurityModeComplete(msg, amf, ue, gnb)

	case nas.MsgTypeRegistrationComplete:
		log.Info("[5GC][NAS] Received Registration Complete")
		err = nasHandler.RegistrationComplete(msg, gnb, ue, *amf)

	case nas.MsgTypeULNASTransport:
		log.Info("[5GC][NAS] Received UL NAS Transport")
		err = nasHandler.UlNasTransport(msg, gnb, ue, session)

	case nas.MsgTypeConfigurationUpdateComplete:
		log.Info("[5GC][NAS] Received Configuration Update Complete")

	case nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration:
		log.Info("[5GC][NAS] Received Deregistration Request: UE Originating Deregistration")
		err = nasHandler.UEOriginatingDeregistration(msg, amf, ue, gnb)
	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}

	if err != nil {
		log.Error(err)
	}
}
