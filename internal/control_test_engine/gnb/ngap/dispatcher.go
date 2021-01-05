package ngap

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/handler"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
)

func Dispatch(amf *context.GNBAmf, gnb *context.GNBContext, message []byte) {

	if message == nil {
		// TODO return error
		fmt.Println("NGAP message is nill")
	}

	// decode NGAP message.
	ngapMsg, err := ngap.Decoder(message)
	if err != nil {
		fmt.Println("Error decoding NGAP message in %d GNB", gnb.GetGnbId())
	}

	// check RanUeId and get UE.

	// handle NGAP message.
	switch ngapMsg.Present {

	case ngapType.NGAPPDUPresentInitiatingMessage:

		switch ngapMsg.InitiatingMessage.ProcedureCode.Value {

		case ngapType.ProcedureCodeDownlinkNASTransport:
			// handler NGAP Downlink NAS Transport.
			fmt.Println("[NGAP]Receive Downlink NAS Transport")
			handler.HandlerDownlinkNasTransport(gnb, ngapMsg)

		case ngapType.ProcedureCodeInitialContextSetup:
			// handler NGAP Initial Context Setup Request.
			fmt.Println("[NGAP]Receive Initial Context Setup Request")
			handler.HandlerInitialContextSetupRequest(gnb, ngapMsg)

		case ngapType.ProcedureCodePDUSessionResourceSetup:
			// handler NGAP PDU Session Resource Setup Request.
			fmt.Println("[NGAP]Receive PDU Session Resource Setup Request")
			handler.HandlerPduSessionResourceSetupRequest(gnb, ngapMsg)
		}

	case ngapType.NGAPPDUPresentSuccessfulOutcome:

		switch ngapMsg.SuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeNGSetup:
			// handler NGAP Setup Response.
			fmt.Println("[NGAP]Receive Ng Setup Response")
			handler.HandlerNgSetupResponse(amf, gnb, ngapMsg)

		}
	}
}
