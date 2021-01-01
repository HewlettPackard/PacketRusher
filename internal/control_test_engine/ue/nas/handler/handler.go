package handler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/sm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"time"
)

// 5GMM main states in the UE
const MM5G_NULL = 0x00
const MM5G_DEREGISTERED = 0x01
const MM5G_REGISTERED_INITIATED = 0x02
const MM5G_REGISTERED = 0x03
const MM5G_SERVICE_REQ_INIT = 0x04
const MM5G_DEREGISTERED_INIT = 0x05

func HandlerAuthenticationRequest(ue *context.UEContext, message *nas.Message) {

	// getting RAND and AUTN from the message.
	rand := message.AuthenticationRequest.GetRANDValue()
	autn := message.AuthenticationRequest.GetAUTN()

	// getting resStar
	resStar := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, rand[:], ue.UeSecurity.Snn, autn[:])

	// getting NAS Authentication Response.
	// TODO change name of this function and clean.
	authenticationResponse := mm_5gs.AuthenticationResponse(resStar, "")

	// change state of ue for registered-initiated
	ue.SetState(MM5G_REGISTERED_INITIATED)

	// sending to GNB
	sender.SendToGnb(ue, authenticationResponse)
}

func HandlerSecurityModeCommand(ue *context.UEContext, message *nas.Message) {

	// getting NAS Security Mode Complete.
	securityModeComplete, err := mm_5gs.SecurityModeComplete(ue)
	if err != nil {
		log.Fatal("Error sending Security Mode Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, securityModeComplete)
}

func HandlerRegistrationAccept(ue *context.UEContext, message *nas.Message) {

	// change the state of ue for registered
	ue.SetState(MM5G_REGISTERED)

	// getting NAS registration complete.
	registrationComplete, err := mm_5gs.RegistrationComplete(ue)
	if err != nil {
		log.Fatal("Error sending Registration Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, registrationComplete)

	// waiting receive Configuration Update Command.
	time.Sleep(20 * time.Millisecond)

	// getting ul nas transport and pduSession establishment request.
	ulNasTransport, err := mm_5gs.UlNasTransport(ue, nasMessage.ULNASTransportRequestTypeInitialRequest)
	if err != nil {
		log.Fatal("Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func HandlerDlNasTransportPduaccept(ue *context.UEContext, message *nas.Message) {

	//getting PDU Session establishment accept.
	payloadContainer := nas_control.GetNasPduFromPduAccept(message)
	if payloadContainer.GmmHeader.GetMessageType() == nas.MsgTypePDUSessionEstablishmentAccept {
		fmt.Println("Receiving PDU Session Establishment Accept")

		// update PDU Session information.

		// get UE ip
		UeIp := sm_5gs.GetPduAdress(payloadContainer)
		ue.SetIp(UeIp)
	}
}
