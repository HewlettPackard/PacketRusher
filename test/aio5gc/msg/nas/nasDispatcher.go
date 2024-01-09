/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/tools"
	nasHandler "my5G-RANTester/test/aio5gc/msg/nas/handler"
	"strconv"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/util/fsm"
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
			log.Fatal(errors.New("[5GC][NAS] message integrity could not be verified"))
		}

	} else {
		msg, err = tools.DecodePlainNasNoIntegrityCheck(payload)
	}

	// Hook for changing 5GC behaviour
	hooks := fgc.GetNasHooks()
	msgHandled := false
	if hooks != nil && len(hooks) > 0 {
		for i := range hooks {
			handled, err := hooks[i](msg, ueContext, gnb, fgc)
			if err != nil {
				log.Fatal(err.Error())
			}
			if handled && msgHandled {
				log.Warn("[5GC][NAS] Message handled several times by hooks")
			} else if handled {
				msgHandled = true
			}
		}
	}
	if msgHandled {
		return
	}

	amf := fgc.GetAMFContext()
	session := fgc.GetSessionContext()
	f := fgc.GetFSM()
	logEntry := log.NewEntry(log.StandardLogger())

	// Default Dispacther
	switch msg.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		log.Info("[5GC][NAS] Received Registration Request")
		err = f.SendEvent(ueContext.GetState().GetState(), fsm.EventType(msg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.RegistrationRequest(msg, amf, ueContext, gnb)

	case nas.MsgTypeIdentityResponse:
		log.Info("[5GC][NAS] Received Identity Response")
		err = nasHandler.IdentityResponse(msg, amf, ueContext, gnb)

	case nas.MsgTypeAuthenticationResponse:
		log.Info("[5GC][NAS] Received Authentication Response")
		err = f.SendEvent(ueContext.GetState().GetState(), fsm.EventType(msg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.AuthenticationResponse(msg, gnb, ueContext, amf)

	case nas.MsgTypeSecurityModeComplete:
		log.Info("[5GC][NAS] Received Security Mode Complete")
		err = f.SendEvent(ueContext.GetState().GetState(), fsm.EventType(msg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.SecurityModeComplete(msg, amf, ueContext, gnb)

	case nas.MsgTypeRegistrationComplete:
		log.Info("[5GC][NAS] Received Registration Complete")
		err = f.SendEvent(ueContext.GetState().GetState(), fsm.EventType(msg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.RegistrationComplete(msg, gnb, ueContext, *amf)

	case nas.MsgTypeULNASTransport:
		log.Info("[5GC][NAS] Received UL NAS Transport")
		err = nasHandler.UlNasTransport(msg, gnb, ueContext, session)

	case nas.MsgTypeConfigurationUpdateComplete:
		log.Info("[5GC][NAS] Received Configuration Update Complete")

	case nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration:
		log.Info("[5GC][NAS] Received Deregistration Request: UE Originating Deregistration")
		err = f.SendEvent(ueContext.GetState().GetState(), fsm.EventType(msg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
		if err != nil {
			log.Warn("[5GC][NAS] Unexpected UE state transition: " + err.Error())
		}
		err = nasHandler.UEOriginatingDeregistration(msg, amf, ueContext, gnb)

	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	if err != nil {
		log.Fatal(err)
	}
}
