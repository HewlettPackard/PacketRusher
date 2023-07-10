package trigger

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"
	"my5G-RANTester/lib/nas/nasMessage"
)

func InitRegistration(ue *context.UEContext) {
	log.Info("[UE] Initiating Registration")

	// registration procedure started.
	registrationRequest := mm_5gs.GetRegistrationRequest(
		nasMessage.RegistrationType5GSInitialRegistration,
		nil,
		nil,
		false,
		ue)

	// send to GNB.
	sender.SendToGnb(ue, registrationRequest)

	// change the state of ue for deregistered
	ue.SetStateMM_DEREGISTERED()
}

func InitNewPduSession(ue *context.UEContext) {
	log.Info("[UE] Initiating New PDU Session")

	pduSession, err := ue.CreatePDUSession()
	if err != nil {
		log.Fatal("[UE][NAS] ", err)
		return
	}

	ulNasTransport, err := mm_5gs.UlNasTransport(pduSession, ue, nasMessage.ULNASTransportRequestTypeInitialRequest)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// change the state of ue(SM).
	ue.SetStateSM_PDU_SESSION_PENDING()

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func InitDeregistration(ue *context.UEContext) {
	log.Info("[UE] Initiating Deregistration")

	// registration procedure started.
	deregistrationRequest := mm_5gs.GetDeregistrationRequest(ue)

	// send to GNB.
	sender.SendToGnb(ue, deregistrationRequest)

	// change the state of ue for deregistered
	ue.SetStateMM_DEREGISTERED()
}