package handler

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/ue/nas/trigger"
	"time"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	log "github.com/sirupsen/logrus"
)

func HandlerAuthenticationReject(ue *context.UEContext, message *nas.Message) {

	log.Info("[UE][NAS] Authentication of UE ", ue.GetUeId(), " failed")

	ue.SetStateMM_DEREGISTERED()
}

func HandlerAuthenticationRequest(ue *context.UEContext, message *nas.Message) {
	var authenticationResponse []byte

	// getting RAND and AUTN from the message.
	rand := message.AuthenticationRequest.GetRANDValue()
	autn := message.AuthenticationRequest.GetAUTN()

	// getting resStar
	paramAutn, check := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, rand[:], ue.UeSecurity.Snn, autn[:])

	switch check {

	case "MAC failure":
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: FAILED")
		log.Info("[UE][NAS] Send authentication failure with MAC failure")
		authenticationResponse = mm_5gs.AuthenticationFailure("MAC failure", "", paramAutn)
		// not change the state of UE.

	case "SQN failure":
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message: INVALID")
		log.Info("[UE][NAS] Send authentication failure with Synch failure")
		authenticationResponse = mm_5gs.AuthenticationFailure("SQN failure", "", paramAutn)
		// not change the state of UE.

	case "successful":
		// getting NAS Authentication Response.
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message: VALID")
		log.Info("[UE][NAS] Send authentication response")
		authenticationResponse = mm_5gs.AuthenticationResponse(paramAutn, "")

		// change state of UE for registered-initiated
		ue.SetStateMM_REGISTERED_INITIATED()
	}

	// sending to GNB
	sender.SendToGnb(ue, authenticationResponse)
}

func HandlerSecurityModeCommand(ue *context.UEContext, message *nas.Message) {

	ue.UeSecurity.CipheringAlg = message.SecurityModeCommand.SelectedNASSecurityAlgorithms.GetTypeOfCipheringAlgorithm()
	switch ue.UeSecurity.CipheringAlg {
	case 0:
		log.Info("[UE][NAS] Type of ciphering algorithm is 5G-EA0")
	case 1:
		log.Info("[UE][NAS] Type of ciphering algorithm is 128-5G-EA1")
	case 2:
		log.Info("[UE][NAS] Type of ciphering algorithm is 128-5G-EA2")
	}

	ue.UeSecurity.IntegrityAlg = message.SecurityModeCommand.SelectedNASSecurityAlgorithms.GetTypeOfIntegrityProtectionAlgorithm()
	switch ue.UeSecurity.IntegrityAlg {
	case 0:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 5G-IA0")
	case 1:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 128-5G-IA1")
	case 2:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 128-5G-IA2")
	}

	rinmr := uint8(0)
	if message.SecurityModeCommand.Additional5GSecurityInformation != nil {
		// checking BIT RINMR that triggered registration request in security mode complete.
		rinmr = message.SecurityModeCommand.Additional5GSecurityInformation.GetRINMR()
	}
	// getting NAS Security Mode Complete.
	securityModeComplete, err := mm_5gs.SecurityModeComplete(ue, rinmr)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending Security Mode Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, securityModeComplete)
}

func HandlerRegistrationAccept(ue *context.UEContext, message *nas.Message) {

	// change the state of ue for registered
	ue.SetStateMM_REGISTERED()

	// saved 5g GUTI and others information.
	ue.SetAmfRegionId(message.RegistrationAccept.GetAMFRegionID())
	ue.SetAmfPointer(message.RegistrationAccept.GetAMFPointer())
	ue.SetAmfSetId(message.RegistrationAccept.GetAMFSetID())
	ue.Set5gGuti(message.RegistrationAccept.GetTMSI5G())

	// use the slice allowed by the network
	// in PDU session request
	if ue.Snssai.Sst == 0 {

		// check the allowed NSSAI received from the 5GC
		snssai := message.RegistrationAccept.AllowedNSSAI.GetSNSSAIValue()

		// update UE slice selected for PDU Session
		ue.Snssai.Sst = int32(snssai[1])
		ue.Snssai.Sd = fmt.Sprintf("0%x0%x0%x", snssai[2], snssai[3], snssai[4])

		log.Warn("[UE][NAS] ALLOWED NSSAI: SST: ", ue.Snssai.Sst, " SD: ", ue.Snssai.Sd)
	}

	log.Info("[UE][NAS] UE 5G GUTI: ", ue.Get5gGuti())

	// getting NAS registration complete.
	registrationComplete, err := mm_5gs.RegistrationComplete(ue)
	if err != nil {
		log.Fatal("[UE][NAS] Error sending Registration Complete: ", err)
	}

	// sending to GNB
	sender.SendToGnb(ue, registrationComplete)

	// waiting receive Configuration Update Command.
	// TODO: Wait more properly for Configuration Update Command
	time.Sleep(50 * time.Millisecond)
}

func HandlerDlNasTransportPduaccept(ue *context.UEContext, message *nas.Message) {

	//getting PDU Session establishment accept.
	payloadContainer := nas_control.GetNasPduFromPduAccept(message)

	switch payloadContainer.GsmHeader.GetMessageType() {
	case nas.MsgTypePDUSessionEstablishmentAccept:
		log.Info("[UE][NAS] Receiving PDU Session Establishment Accept")

		// get UE ip
		pduSessionEstablishmentAccept := payloadContainer.PDUSessionEstablishmentAccept

		// update PDU Session information.
		pduSessionId := pduSessionEstablishmentAccept.GetPDUSessionID()
		pduSession, err := ue.GetPduSession(pduSessionId)
		// change the state of ue(SM)(PDU Session Active).
		pduSession.SetStateSM_PDU_SESSION_ACTIVE()
		if err != nil {
			log.Error("[UE][NAS] Receiving PDU Session Establishment Accept about an unknown PDU Session, id: ", pduSessionId)
			return
		}

		UeIp := pduSessionEstablishmentAccept.GetPDUAddressInformation()
		pduSession.SetIp(UeIp)

	case nas.MsgTypePDUSessionReleaseCommand:
		log.Info("[UE][NAS] Receiving PDU Session Release Command")

		pduSessionReleaseCommand := payloadContainer.PDUSessionReleaseCommand
		pduSessionId := pduSessionReleaseCommand.GetPDUSessionID()
		pduSession, err := ue.GetPduSession(pduSessionId)
		if pduSession == nil || err != nil {
			log.Error("[UE][NAS] Unable to delete PDU Session ", pduSessionId, " from UE ", ue.GetMsin(), " as the PDU Session was not found. Ignoring.")
			break
		}
		ue.DeletePduSession(pduSessionId)
		log.Info("[UE][NAS] Successfully released PDU Session ", pduSessionId, " from UE Context")

	case nas.MsgTypePDUSessionEstablishmentReject:
		log.Error("[UE][NAS] Receiving PDU Session Establishment Reject")

		pduSessionEstablishmentReject := payloadContainer.PDUSessionEstablishmentReject
		pduSessionId := pduSessionEstablishmentReject.GetPDUSessionID()

		log.Error("[UE][NAS] PDU Session Establishment Reject for PDU Session ID ", pduSessionId, ", 5GSM Cause: ", cause5GSMToString(pduSessionEstablishmentReject.GetCauseValue()))

		// Per 5GSM state machine in TS 24.501 - 6.1.3.2.1., we re-try the setup until it's successful
		pduSession, err := ue.GetPduSession(pduSessionId)
		if err != nil {
			log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSession, " after Reject as ", err)
			break
		}
		trigger.InitPduSessionRequestInner(ue, pduSession)

	default:
		log.Error("[UE][NAS] Receiving Unknown Dl NAS Transport message!! ", payloadContainer.GsmHeader.GetMessageType())
	}
}

func cause5GSMToString(causeValue uint8) string {
	switch causeValue {
	case nasMessage.Cause5GSMInsufficientResources:
		return "Insufficient Ressources"
	case nasMessage.Cause5GSMMissingOrUnknownDNN:
		return "Missing or Unknown DNN"
	case nasMessage.Cause5GSMUnknownPDUSessionType:
		return "Unknown PDU Session Type"
	case nasMessage.Cause5GSMUserAuthenticationOrAuthorizationFailed:
		return "User authentification or authorization failed"
	case nasMessage.Cause5GSMRequestRejectedUnspecified:
		return "Request rejected, unspecified"
	case nasMessage.Cause5GSMServiceOptionTemporarilyOutOfOrder:
		return "Service option temporarily out of order. Please share pcap with packetrusher@hpe.com."
	case nasMessage.Cause5GSMPTIAlreadyInUse:
		return "PTI already in use"
	case nasMessage.Cause5GSMRegularDeactivation:
		return "Regular deactivation"
	case nasMessage.Cause5GSMReactivationRequested:
		return "Reactivation requested"
	case nasMessage.Cause5GSMInvalidPDUSessionIdentity:
		return "Invalid PDU session identity"
	case nasMessage.Cause5GSMSemanticErrorsInPacketFilter:
		return "Semantic errors in packet filter(s)"
	case nasMessage.Cause5GSMSyntacticalErrorInPacketFilter:
		return "Syntactical error in packet filter(s)"
	case nasMessage.Cause5GSMOutOfLADNServiceArea:
		return "Out of LADN service area"
	case nasMessage.Cause5GSMPTIMismatch:
		return "PTI mismatch"
	case nasMessage.Cause5GSMPDUSessionTypeIPv4OnlyAllowed:
		return "PDU session type IPv4 only allowed"
	case nasMessage.Cause5GSMPDUSessionTypeIPv6OnlyAllowed:
		return "PDU session type IPv6 only allowed"
	case nasMessage.Cause5GSMPDUSessionDoesNotExist:
		return "PDU session does not exist"
	case nasMessage.Cause5GSMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nasMessage.Cause5GSMNotSupportedSSCMode:
		return "Not supported SSC mode"
	case nasMessage.Cause5GSMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nasMessage.Cause5GSMMissingOrUnknownDNNInASlice:
		return "Missing or unknown DNN in a slice"
	case nasMessage.Cause5GSMInvalidPTIValue:
		return "Invalid PTI value"
	case nasMessage.Cause5GSMMaximumDataRatePerUEForUserPlaneIntegrityProtectionIsTooLow:
		return "Maximum data rate per UE for user-plane integrity protection is too low"
	case nasMessage.Cause5GSMSemanticErrorInTheQoSOperation:
		return "Semantic error in the QoS operation"
	case nasMessage.Cause5GSMSyntacticalErrorInTheQoSOperation:
		return "Syntactical error in the QoS operation"
	case nasMessage.Cause5GSMInvalidMappedEPSBearerIdentity:
		return "Invalid mapped EPS bearer identity"
	case nasMessage.Cause5GSMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nasMessage.Cause5GSMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nasMessage.Cause5GSMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implemented"
	case nasMessage.Cause5GSMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nasMessage.Cause5GSMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nasMessage.Cause5GSMConditionalIEError:
		return "Conditional IE error"
	case nasMessage.Cause5GSMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nasMessage.Cause5GSMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please share pcap with packetrusher@hpe.com."
	default:
		return "Service option temporarily out of order. Please share pcap with packetrusher@hpe.com."
	}
}
