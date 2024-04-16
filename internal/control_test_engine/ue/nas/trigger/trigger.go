/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

// Package trigger
// Triggers are the basic building block of test scenarios.
// They allow to trigger NAS procedures by the UE, eg: registration, PDU Session creation...
// Replies from the AMF are handled in the global handler.
package trigger

import (
	context2 "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	log "github.com/sirupsen/logrus"
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

	var err error
	if len(ue.UeSecurity.Kamf) != 0 {
		registrationRequest, err = nas_control.EncodeNasPduWithSecurity(ue, registrationRequest, nas.SecurityHeaderTypeIntegrityProtected, true, false)
		if err != nil {
			log.Fatalf("[UE][NAS] Unable to encode with integrity protection Registration Request: %s", err)
		}
	}
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

	InitPduSessionRequestInner(ue, pduSession)
}

func InitPduSessionRequestInner(ue *context.UEContext, pduSession *context.UEPDUSession) {

	ulNasTransport, err := mm_5gs.Request_UlNasTransport(pduSession, ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// change the state of ue(SM).
	pduSession.SetStateSM_PDU_SESSION_PENDING()

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func InitPduSessionRelease(ue *context.UEContext, pduSession *context.UEPDUSession) {
	log.Info("[UE] Initiating Release of PDU Session ", pduSession.Id)

	if pduSession.GetStateSM() != context.SM5G_PDU_SESSION_ACTIVE {
		log.Warn("[UE][NAS] Skipping releasing the PDU Session ID ", pduSession.Id, " as it's not active")
		return
	}

	ulNasTransport, err := mm_5gs.Release_UlNasTransport(pduSession, ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// change the state of ue(SM).
	pduSession.SetStateSM_PDU_SESSION_INACTIVE()

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func InitPduSessionReleaseComplete(ue *context.UEContext, pduSession *context.UEPDUSession) {
	log.Info("[UE] Initiating PDU Session Release Complete for PDU Session", pduSession.Id)

	if pduSession.GetStateSM() != context.SM5G_PDU_SESSION_INACTIVE {
		log.Warn("[UE][NAS] Unable to send PDU Session Release Complete for a PDU Session which is not inactive")
		return
	}

	ulNasTransport, err := mm_5gs.ReleasComplete_UlNasTransport(pduSession, ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, ulNasTransport)
}

func InitDeregistration(ue *context.UEContext) {
	log.Info("[UE] Initiating Deregistration")

	// registration procedure started.
	deregistrationRequest, err := mm_5gs.DeregistrationRequest(ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending deregistration request: ", err)
	}

	// send to GNB.
	sender.SendToGnb(ue, deregistrationRequest)

	// change the state of ue for deregistered
	ue.SetStateMM_DEREGISTERED()
}

func InitIdentifyResponse(ue *context.UEContext) {
	log.Info("[UE] Initiating Identify Response")

	// trigger identity response.
	identityResponse := mm_5gs.IdentityResponse(ue)

	// send to GNB.
	sender.SendToGnb(ue, identityResponse)
}

func InitConfigurationUpdateComplete(ue *context.UEContext) {
	log.Info("[UE] Initiating Configuration Update Complete")

	// trigger Configuration Update Complete.
	identityResponse, err := mm_5gs.ConfigurationUpdateComplete(ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending Configuration Update Complete: ", err)
	}
	// send to GNB.
	sender.SendToGnb(ue, identityResponse)
}

func InitServiceRequest(ue *context.UEContext) {
	log.Info("[UE] Initiating Service Request")

	// trigger ServiceRequest.
	serviceRequest := mm_5gs.ServiceRequest(ue)
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, serviceRequest, nas.SecurityHeaderTypeIntegrityProtected, true, false)

	if err != nil {
		log.Fatalf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}

	// send to GNB.
	sender.SendToGnb(ue, pdu)
}

func SwitchToIdle(ue *context.UEContext) {
	log.Info("[UE] Switching to 5GMM-IDLE")

	// send to GNB.
	sender.SendToGnbMsg(ue, context2.UEMessage{Idle: true})
}
