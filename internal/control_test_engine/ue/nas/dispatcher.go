package nas

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/handler"
	"my5G-RANTester/lib/nas"
)

func Dispatch(ue *context.UEContext, message []byte) {

	// check if message is null.
	if message == nil {
		// TODO return error
		fmt.Println("NAS message is nill")
	}

	// decode NAS message.
	m := new(nas.Message)
	err := m.PlainNasDecode(&message)
	if err != nil {
		// TODO return error
		fmt.Println("check error")
	}

	// check if NAS is security protected
	if m.SecurityHeader.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {

		// security protected.
		payload := message

		// remove security header.
		payload = payload[7:]

		// decode NAS message again now left security header.
		err := m.PlainNasDecode(&payload)
		if err != nil {
			// TODO return error
			fmt.Println("check error")
		}

		// TODO check security header
	}

	switch m.GmmHeader.GetMessageType() {

	case nas.MsgTypeAuthenticationRequest:
		// handler authentication request.
		handler.HandlerAuthenticationRequest(ue, m)

	case nas.MsgTypeIdentityRequest:
		// handler identity request.

	case nas.MsgTypeSecurityModeCommand:
		// handler security mode command.
		handler.HandlerSecurityModeCommand(ue, m)

	case nas.MsgTypeRegistrationAccept:
		// handler registration accept.
		handler.HandlerRegistrationAccept(ue, m)

	case nas.MsgTypeConfigurationUpdateCommand:
		// handler Configuration Update Command.

	case nas.MsgTypeDLNASTransport:
		// handler DL NAS Transport.
		handler.HandlerDlNasTransportPduaccept(ue, m)

		// UE is ready for testing data-plane.
	}

}
