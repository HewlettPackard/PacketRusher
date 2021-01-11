package trigger

import (
	"fmt"
	"log"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/interface_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_context_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
)

func SendPduSessionResourceSetupResponse(ue *context.GNBUe, gnb *context.GNBContext) error {

	// send PDU Session Resource Setup Response.
	gnbIp := gnb.GetGnbIpByData()
	ngapMsg, err := pdu_session_management.PDUSessionResourceSetupResponse(ue, gnbIp)
	if err != nil {
		return fmt.Errorf("Error sending PDU Session Resource Setup Response.")
	}

	ue.SetStateReady()

	// Send PDU Session Resource Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		return fmt.Errorf("Error sending PDU Session Resource Setup Response.: ", err)
	}

	return nil
}

func SendInitialContextSetupResponse(ue *context.GNBUe) {

	// send Initial Context Setup Response.
	ngapMsg, err := ue_context_management.InitialContextSetupResponse(ue)
	if err != nil {
		log.Fatal("Error sending Initial Context Setup Response")
	}

	// Send Initial Context Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("Error sending Initial Context Setup Response: ", err)
	}
}

func SendNgSetupRequest(gnb *context.GNBContext, amf *context.GNBAmf) {

	// send NG setup response.
	ngapMsg, err := interface_management.NGSetupRequest(gnb, "my5gRANTester")
	if err != nil {
		log.Fatal("Error sending NG Setup Request")

	}

	conn := amf.GetSCTPConn()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		fmt.Println("Error sending NG Setup Request: ", err)

	}

}
