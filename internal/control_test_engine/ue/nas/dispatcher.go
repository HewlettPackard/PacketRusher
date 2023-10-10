package nas

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/handler"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"reflect"
)

func DispatchNas(ue *context.UEContext, message []byte) {

	var cph bool

	// check if message is null.
	if message == nil {
		// TODO return error
		log.Fatal("[UE][NAS] NAS message is nil")
	}

	// decode NAS message.
	m := new(nas.Message)
	m.SecurityHeaderType = nas.GetSecurityHeaderType(message) & 0x0f

	payload := message

	// check if NAS is security protected
	if m.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {

		log.Info("[UE][NAS] Message with security header")

		// information to check integrity and ciphered.

		// sequence number.
		sequenceNumber := payload[6]

		// mac verification.
		macReceived := payload[2:6]

		// remove security Header except for sequence Number
		payload := payload[6:]

		// check security header type.
		cph = false
		switch m.SecurityHeaderType {

		case nas.SecurityHeaderTypeIntegrityProtected:
			log.Info("[UE][NAS] Message with integrity")

		case nas.SecurityHeaderTypeIntegrityProtectedAndCiphered:
			log.Info("[UE][NAS] Message with integrity and ciphered")
			cph = true

		case nas.SecurityHeaderTypeIntegrityProtectedWithNew5gNasSecurityContext:
			log.Info("[UE][NAS] Message with integrity and with NEW 5G NAS SECURITY CONTEXT")
			ue.UeSecurity.DLCount.Set(0, 0)

		case nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext:
			log.Info("[UE][NAS] Message with integrity, ciphered and with NEW 5G NAS SECURITY CONTEXT")
			cph = true
			ue.UeSecurity.DLCount.Set(0, 0)

		}

		// check security header(Downlink data).
		if ue.UeSecurity.DLCount.SQN() > sequenceNumber {
			ue.UeSecurity.DLCount.SetOverflow(ue.UeSecurity.DLCount.Overflow() + 1)
		}
		ue.UeSecurity.DLCount.SetSQN(sequenceNumber)

		mac32, err := security.NASMacCalculate(ue.UeSecurity.IntegrityAlg,
			ue.UeSecurity.KnasInt,
			ue.UeSecurity.DLCount.Get(),
			security.Bearer3GPP,
			security.DirectionDownlink, payload)
		if err != nil {
			log.Info("NAS MAC calculate error")
			return
		}

		// check integrity
		if !reflect.DeepEqual(mac32, macReceived) {
			log.Info("[UE][NAS] NAS MAC verification failed(received:", macReceived, "expected:", mac32)
			return
		} else {
			log.Info("[UE][NAS] successful NAS MAC verification")
		}

		// check ciphering.
		if cph {
			if err = security.NASEncrypt(ue.UeSecurity.CipheringAlg, ue.UeSecurity.KnasEnc, ue.UeSecurity.DLCount.Get(), security.Bearer3GPP,
				security.DirectionDownlink, payload[1:]); err != nil {
				log.Info("error in encrypt algorithm")
				return
			} else {
				log.Info("[UE][NAS] successful NAS CIPHERING")
			}
		}

		// remove security header.
		payload = message[7:]

		// decode NAS message.
		err = m.PlainNasDecode(&payload)
		if err != nil {
			// TODO return error
			log.Info("[UE][NAS] Decode NAS error", err)
		}

	} else {

		log.Info("[UE][NAS] Message without security header")

		// decode NAS message.
		err := m.PlainNasDecode(&payload)
		if err != nil {
			// TODO return error
			log.Info("[UE][NAS] Decode NAS error", err)
		}
	}

	switch m.GmmHeader.GetMessageType() {

	case nas.MsgTypeAuthenticationRequest:
		// handler authentication request.
		log.Info("[UE][NAS] Receive Authentication Request")
		handler.HandlerAuthenticationRequest(ue, m)

	case nas.MsgTypeAuthenticationReject:
		// handler authentication reject.
		log.Info("[UE][NAS] Receive Authentication Reject")
		handler.HandlerAuthenticationReject(ue, m)

	case nas.MsgTypeIdentityRequest:
		log.Info("[UE][NAS] Receive Identify Request")
		// handler identity request.

	case nas.MsgTypeSecurityModeCommand:
		// handler security mode command.
		log.Info("[UE][NAS] Receive Security Mode Command")
		handler.HandlerSecurityModeCommand(ue, m)

	case nas.MsgTypeRegistrationAccept:
		// handler registration accept.
		log.Info("[UE][NAS] Receive Registration Accept")
		handler.HandlerRegistrationAccept(ue, m)

	case nas.MsgTypeConfigurationUpdateCommand:
		log.Info("[UE][NAS] Receive Configuration Update Command")
		log.Info("[UE][NAS] ", ue.GetSupi(), " was succesfully registered")

	case nas.MsgTypeDLNASTransport:
		// handler DL NAS Transport.
		log.Info("[UE][NAS] Receive DL NAS Transport")
		handleCause5GMM(m.DLNASTransport.Cause5GMM)
		handler.HandlerDlNasTransportPduaccept(ue, m)

	case nas.MsgTypeRegistrationReject:
		// handler registration reject
		log.Error("[UE][NAS] Receive Registration Reject")
		handleCause5GMM(&m.RegistrationReject.Cause5GMM)
	}

}


func handleCause5GMM(cause5GMM *nasType.Cause5GMM) {
	if cause5GMM != nil {
		log.Error("[UE][NAS] UE received a 5GMM Failure, cause: ", cause5GMMToString(cause5GMM.Octet))
	}
}

func cause5GMMToString(cause5GMM uint8) string {
	switch cause5GMM {
	case nasMessage.Cause5GMMIllegalUE:
		return "Illegal UE"
	case nasMessage.Cause5GMMPEINotAccepted:
		return "PEI not accepted"
	case nasMessage.Cause5GMMIllegalME:
		return "5GS services not allowed"
	case nasMessage.Cause5GMM5GSServicesNotAllowed:
		return "5GS services not allowed"
	case nasMessage.Cause5GMMUEIdentityCannotBeDerivedByTheNetwork:
		return "UE identity cannot be derived by the network"
	case nasMessage.Cause5GMMImplicitlyDeregistered:
		return "Implicitly de-registered"
	case nasMessage.Cause5GMMPLMNNotAllowed:
		return "PLMN not allowed"
	case nasMessage.Cause5GMMTrackingAreaNotAllowed:
		return "Tracking area not allowed"
	case nasMessage.Cause5GMMRoamingNotAllowedInThisTrackingArea:
		return "Roaming not allowed in this tracking area"
	case nasMessage.Cause5GMMNoSuitableCellsInTrackingArea:
		return "No suitable cells in tracking area"
	case nasMessage.Cause5GMMMACFailure:
		return "MAC failure"
	case nasMessage.Cause5GMMSynchFailure:
		return "Synch failure"
	case nasMessage.Cause5GMMCongestion:
		return "Congestion"
	case nasMessage.Cause5GMMUESecurityCapabilitiesMismatch:
		return "UE security capabilities mismatch"
	case nasMessage.Cause5GMMSecurityModeRejectedUnspecified:
		return "Security mode rejected, unspecified"
	case nasMessage.Cause5GMMNon5GAuthenticationUnacceptable:
		return "Non-5G authentication unacceptable"
	case nasMessage.Cause5GMMN1ModeNotAllowed:
		return "N1 mode not allowed"
	case nasMessage.Cause5GMMRestrictedServiceArea:
		return "Restricted service area"
	case nasMessage.Cause5GMMLADNNotAvailable:
		return "LADN not available"
	case nasMessage.Cause5GMMMaximumNumberOfPDUSessionsReached:
		return "Maximum number of PDU sessions reached"
	case nasMessage.Cause5GMMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nasMessage.Cause5GMMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nasMessage.Cause5GMMngKSIAlreadyInUse:
		return "ngKSI already in use"
	case nasMessage.Cause5GMMNon3GPPAccessTo5GCNNotAllowed:
		return "Non-3GPP access to 5GCN not allowed"
	case nasMessage.Cause5GMMServingNetworkNotAuthorized:
		return "Serving network not authorized"
	case nasMessage.Cause5GMMPayloadWasNotForwarded:
		return "Payload was not forwarded"
	case nasMessage.Cause5GMMDNNNotSupportedOrNotSubscribedInTheSlice:
		return "DNN not supported or not subscribed in the slice"
	case nasMessage.Cause5GMMInsufficientUserPlaneResourcesForThePDUSession:
		return "Insufficient user-plane resources for the PDU session"
	case nasMessage.Cause5GMMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nasMessage.Cause5GMMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nasMessage.Cause5GMMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implementedE"
	case nasMessage.Cause5GMMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nasMessage.Cause5GMMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nasMessage.Cause5GMMConditionalIEError:
		return "Conditional IE error"
	case nasMessage.Cause5GMMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nasMessage.Cause5GMMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please share the pcap with packetrusher@hpe.com."
	default:
		return "Protocol error, unspecified. Please share the pcap with packetrusher@hpe.com."
	}
}