package nas

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/handler"
	"my5G-RANTester/lib/nas"
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
		// handler Configuration Update Command.

	case nas.MsgTypeDLNASTransport:
		// handler DL NAS Transport.
		log.Info("[UE][NAS] Receive DL NAS Transport")
		handler.HandlerDlNasTransportPduaccept(ue, m)

	case nas.MsgTypeRegistrationReject:
		// handler registration reject
		log.Info("[UE][NAS] Receive Registration Reject")
	}

}
