package nas

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/handler"
	"my5G-RANTester/lib/nas"
)

func DispatchNas(ue *context.UEContext, message []byte) {

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

		// remove security header.
		payload = payload[7:]

		// TODO check security header
	} else {
		log.Info("[UE][NAS] Message without security header")
	}

	// decode NAS message.
	err := m.PlainNasDecode(&payload)
	if err != nil {
		// TODO return error
		log.Info("[UE][NAS] Decode NAS error", err)
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

		// UE is ready for testing data-plane.
	}

}
