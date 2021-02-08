package handler

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"time"
)

func HandlerAuthenticationRequest(ue *context.UEContext, message *nas.Message) {
	var authenticationResponse []byte

	// getting RAND and AUTN from the message.
	rand := message.AuthenticationRequest.GetRANDValue()
	autn := message.AuthenticationRequest.GetAUTN()

	// getting resStar
	resStar, check := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, rand[:], ue.UeSecurity.Snn, autn[:])
	switch check {

	case "MAC failure":
		authenticationResponse = mm_5gs.AuthenticationFailure("")
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: FAILED")
		log.Info("[UE][NAS] Send authentication failure with MAC failure")
		// not change the state of UE.
		break

	case "SQN failure":
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message range: FAILED")
		log.Info("[UE][NAS] Send authentication failure with Synch failure")
		// not change the state of UE.
		break

	case "successful":
		// getting NAS Authentication Response.
		authenticationResponse = mm_5gs.AuthenticationResponse(resStar, "")
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message range: OK")
		log.Info("[UE][NAS] Send authentication response")

		// change state of UE for registered-initiated
		ue.SetStateMM_REGISTERED_INITIATED()
		break
	}

	// sending to GNB
	sender.SendToGnb(ue, authenticationResponse)
}

func HandlerSecurityModeCommand(ue *context.UEContext, message *nas.Message) {

	// TODO checking BIT RINMR that triggered registration request in security mode complete
	// message.SecurityModeCommand.Additional5GSecurityInformation.GetRINMR()

	// getting NAS Security Mode Complete.
	securityModeComplete, err := mm_5gs.SecurityModeComplete(ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending Security Mode Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, securityModeComplete)
}

func HandlerRegistrationAccept(ue *context.UEContext, message *nas.Message) {

	// change the state of ue for registered
	ue.SetStateMM_REGISTERED()

	// getting NAS registration complete.
	registrationComplete, err := mm_5gs.RegistrationComplete(ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending Registration Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, registrationComplete)

	// waiting receive Configuration Update Command.
	time.Sleep(20 * time.Millisecond)

	// getting ul nas transport and pduSession establishment request.
	ulNasTransport, err := mm_5gs.UlNasTransport(ue, nasMessage.ULNASTransportRequestTypeInitialRequest)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// change the sate of ue(SM).
	ue.SetStateSM_PDU_SESSION_PENDING()

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func HandlerDlNasTransportPduaccept(ue *context.UEContext, message *nas.Message) {

	//getting PDU Session establishment accept.
	payloadContainer := nas_control.GetNasPduFromPduAccept(message)
	if payloadContainer.GsmHeader.GetMessageType() == nas.MsgTypePDUSessionEstablishmentAccept {
		log.Info("[UE][NAS] Receiving PDU Session Establishment Accept")

		// update PDU Session information.

		// change the state of ue(SM)(PDU Session Active).
		ue.SetStateSM_PDU_SESSION_ACTIVE()

		// get UE ip
		UeIp := payloadContainer.PDUSessionEstablishmentAccept.GetPDUAddressInformation()
		ue.SetIp(UeIp)
	}
}
