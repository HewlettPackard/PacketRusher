// Package trigger
// Triggers are the basic building block of test scenarios.
// They allow to trigger NAS procedures by the UE, eg: registration, PDU Session creation...
// Replies from the AMF are handled in the global handler.
// TODO: Maybe it would be better to have a NAS state machine per trigger? -> It would allow more easily for negative testing
package trigger

import (
	log "github.com/sirupsen/logrus"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
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

func InitPduSessionRequest(ue *context.UEContext) {
	log.Info("[UE] Initiating New PDU Session")

	pduSession, err := ue.CreatePDUSession()
	if err != nil {
		log.Fatal("[UE][NAS] ", err)
		return
	}

	ulNasTransport, err := mm_5gs.Request_UlNasTransport(pduSession, ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// change the state of ue(SM).
	ue.SetStateSM_PDU_SESSION_PENDING()

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func InitPduSessionRelease(ue *context.UEContext, pduSession *context.PDUSession) {
	log.Info("[UE] Initiating Release of PDU Session ", pduSession.Id)

	ulNasTransport, err := mm_5gs.Release_UlNasTransport(pduSession, ue)
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

func InitHandover(ue *context.UEContext, gnbChan chan gnbContext.UEMessage) {
	log.Info("[UE] Initiating Handover")

	// Stop connectivity with previous gNb
	close(ue.GetGnbRx())

	newGnbRx := make(chan gnbContext.UEMessage, 1)
	newGnbTx := make(chan gnbContext.UEMessage, 1)
	ue.SetGnbRx(newGnbRx)
	ue.SetGnbTx(newGnbTx)

	// Connect to new gNb
	gnbChan <- gnbContext.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, Msin: ue.GetMsin()}

	// Trigger Handover
	ue.GetGnbRx() <- gnbContext.UEMessage{AmfId: ue.GetAmfUeId()}

	// TODO: Test to remove
	InitDeregistration(ue)
}