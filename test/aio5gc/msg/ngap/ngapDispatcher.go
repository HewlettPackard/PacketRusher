/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"errors"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"

	ngapHandler "my5G-RANTester/test/aio5gc/msg/ngap/handler"
	"os"

	log "github.com/sirupsen/logrus"
)

func Dispatch(buf []byte, gnb *context.GNBContext, fgc *context.Aio5gc) {

	ngapMsg, err := ngap.Decoder(buf)
	if err != nil {
		log.Printf("[5GC] Ngap Decode failed: %v", err)
		os.Exit(1)
	}

	// Hooks for changing 5GC behaviour
	hooks := fgc.GetNgapHooks()
	msgHandled := false
	if hooks != nil && len(hooks) > 0 {
		for i := range hooks {
			handled, err := hooks[i](ngapMsg, gnb, fgc)
			if err != nil {
				log.Fatal(err.Error())
			}
			if handled && msgHandled {
				log.Warn("[5GC][NGAP] Message handled several times by hooks")
			} else if handled {
				msgHandled = true
			}
		}
	}
	if msgHandled {
		return
	}

	// Default Dispacther
	switch ngapMsg.Present {
	case ngapType.NGAPPDUPresentInitiatingMessage:

		switch ngapMsg.InitiatingMessage.ProcedureCode.Value {

		case ngapType.ProcedureCodeNGSetup:
			log.Info("[5GC][NGAP] Received NG Setup Request")
			err = ngapHandler.NGSetupRequest(ngapMsg.InitiatingMessage.Value.NGSetupRequest, gnb, fgc)

		case ngapType.ProcedureCodeInitialUEMessage:
			log.Info("[5GC][NGAP] Received Initial UE Message Request")
			err = ngapHandler.InitialUEMessage(ngapMsg.InitiatingMessage.Value.InitialUEMessage, gnb, fgc)

		case ngapType.ProcedureCodeUplinkNASTransport:
			log.Info("[5GC][NGAP] Received uplink NAS Transport")
			err = ngapHandler.UplinkNASTransport(ngapMsg.InitiatingMessage.Value.UplinkNASTransport, gnb, fgc)

		default:
			err = errors.New("[5GC][NGAP] Received unknown NGAP NGAPPDUPresentInitiatingMessage ProcedureCode")
		}

	case ngapType.NGAPPDUPresentSuccessfulOutcome:

		switch ngapMsg.SuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeInitialContextSetup:
			log.Info("[5GC][NGAP] Received Successful Initial Context Setup Response")
			ngapHandler.InitialContextSetupResponse(ngapMsg.SuccessfulOutcome.Value.InitialContextSetupResponse, fgc)

		case ngapType.ProcedureCodePDUSessionResourceSetup:
			log.Info("[5GC][NGAP] Received Successful PDU Session Resource Setup response")
			ngapHandler.PDUSessionResourceSetup(ngapMsg.SuccessfulOutcome.Value.PDUSessionResourceSetupResponse, fgc)

		case ngapType.ProcedureCodePDUSessionResourceRelease:
			log.Info("[5GC][NGAP] Received Successful Procedure Code PDU Session Resource Release")

		default:
			err = errors.New("[5GC][NGAP] Received unknown NGAP NGAPPDUPresentSuccessfulOutcome ProcedureCode")
		}

	case ngapType.NGAPPDUPresentUnsuccessfulOutcome:

		switch ngapMsg.UnsuccessfulOutcome.ProcedureCode.Value {

		default:
			err = errors.New("[5GC][NGAP] Received unknown NGAP NGAPPDUPresentUnsuccessfulOutcome ProcedureCode")
		}
	default:
		err = errors.New("[5GC][NGAP] Received unknown NGAP message")
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
