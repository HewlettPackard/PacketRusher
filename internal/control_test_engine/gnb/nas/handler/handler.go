package handler

import (
	"log"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/nas_transport"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
)

func HandlerUeInitialized(ue *context.GNBUe, message []byte, gnb *context.GNBContext) {

	// encode NAS message in NGAP.
	ngap, err := nas_transport.SendInitialUeMessage(message, ue, gnb)
	if err != nil {
		log.Fatal("Error making initial UE message: ", err)
	}

	// change state of UE.
	ue.SetState(context.Ongoing)

	// Send Initial UE Message
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngap, conn)
	if err != nil {
		log.Fatal("Error sending initial UE message: ", err)
	}
}

func HandlerUeOngoing(ue *context.GNBUe, message []byte, gnb *context.GNBContext) {

	ngap, err := nas_transport.SendUplinkNasTransport(message, ue, gnb)
	if err != nil {
		log.Fatal("Error making initial UE message: ", err)
	}

	// Send Uplink Nas Transport
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngap, conn)
	if err != nil {
		log.Fatal("Error sending Uplink Nas Transport: ", err)
	}
}

func HandlerUeReady(ue *context.GNBUe, message []byte, gnb *context.GNBContext) {

	// receive UE ip or other messages.
}
