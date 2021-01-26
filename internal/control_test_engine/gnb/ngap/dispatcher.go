package ngap

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/handler"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
)

func Dispatch(amf *context.GNBAmf, gnb *context.GNBContext, message []byte) {

	if message == nil {
		// TODO return error
		log.Info("[GNB][NGAP] NGAP message is nil")
	}

	// decode NGAP message.
	ngapMsg, err := ngap.Decoder(message)
	if err != nil {
		log.Info("[GNB][NGAP] Error decoding NGAP message in ", gnb.GetGnbId(), " GNB")
	}

	// check RanUeId and get UE.

	// handle NGAP message.
	switch ngapMsg.Present {

	case ngapType.NGAPPDUPresentInitiatingMessage:

		switch ngapMsg.InitiatingMessage.ProcedureCode.Value {

		case ngapType.ProcedureCodeDownlinkNASTransport:
			// handler NGAP Downlink NAS Transport.
			log.Info("[GNB][NGAP] Receive Downlink NAS Transport")
			handler.HandlerDownlinkNasTransport(gnb, ngapMsg)

		case ngapType.ProcedureCodeInitialContextSetup:
			// handler NGAP Initial Context Setup Request.
			log.Info("[GNB][NGAP] Receive Initial Context Setup Request")
			handler.HandlerInitialContextSetupRequest(gnb, ngapMsg)

		case ngapType.ProcedureCodePDUSessionResourceSetup:
			// handler NGAP PDU Session Resource Setup Request.
			log.Info("[GNB][NGAP] Receive PDU Session Resource Setup Request")
			handler.HandlerPduSessionResourceSetupRequest(gnb, ngapMsg)
		}

	case ngapType.NGAPPDUPresentSuccessfulOutcome:

		switch ngapMsg.SuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeNGSetup:
			// handler NGAP Setup Response.
			log.Info("[GNB][NGAP] Receive Ng Setup Response")
			handler.HandlerNgSetupResponse(amf, gnb, ngapMsg)

		}

	case ngapType.NGAPPDUPresentUnsuccessfulOutcome:

		switch ngapMsg.UnsuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeNGSetup:
			// handler NGAP Setup Failure.
			log.Info("[GNB][NGAP] Receive Ng Setup Failure")
			handler.HandlerNgSetupFailure(amf, gnb, ngapMsg)

		}
	}
}
