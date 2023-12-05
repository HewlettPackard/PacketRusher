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

	// Hook for changing 5GC behaviour
	hook := fgc.GetNgapHook()
	if hook != nil {
		handled, err := hook(ngapMsg, gnb, fgc)
		if err != nil {
			log.Fatal(err.Error())
		} else if handled {
			return
		}
	}

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
			err = ngapHandler.UplinkNASTransport(ngapMsg.InitiatingMessage.Value.UplinkNASTransport, *gnb, fgc)

		default:
			err = errors.New("[5GC][NGAP] Received unknown NGAP NGAPPDUPresentInitiatingMessage ProcedureCode")
		}

	case ngapType.NGAPPDUPresentSuccessfulOutcome:

		switch ngapMsg.SuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeInitialContextSetup:
			log.Info("[5GC][NGAP] Received initial context setup response")
			ngapHandler.InitialContextSetupResponse(ngapMsg.SuccessfulOutcome.Value.InitialContextSetupResponse, fgc)

		case ngapType.ProcedureCodePDUSessionResourceSetup:
			log.Info("[5GC][NGAP] Received PDU Session Resource Setup response")
			ngapHandler.PDUSessionResourceSetup(ngapMsg.SuccessfulOutcome.Value.PDUSessionResourceSetupResponse, fgc)

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
