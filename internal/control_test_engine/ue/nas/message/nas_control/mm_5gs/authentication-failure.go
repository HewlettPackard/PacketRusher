package mm_5gs

import (
	"bytes"
	"fmt"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
)

func AuthenticationFailure(cause, eapMsg string, paramAutn []byte) (nasPdu []byte) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeAuthenticationFailure)

	authenticationFailure := nasMessage.NewAuthenticationFailure(0)
	authenticationFailure.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	authenticationFailure.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	authenticationFailure.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	authenticationFailure.AuthenticationFailureMessageIdentity.SetMessageType(nas.MsgTypeAuthenticationFailure)

	switch cause {

	case "MAC failure":
		authenticationFailure.Cause5GMM.SetCauseValue(nasMessage.Cause5GMMMACFailure)
	case "SQN failure":
		authenticationFailure.Cause5GMM.SetCauseValue(nasMessage.Cause5GMMSynchFailure)
		authenticationFailure.AuthenticationFailureParameter = nasType.NewAuthenticationFailureParameter(nasMessage.AuthenticationFailureAuthenticationFailureParameterType)
		authenticationFailure.AuthenticationFailureParameter.SetLen(uint8(len(paramAutn)))
		copy(authenticationFailure.AuthenticationFailureParameter.Octet[:], paramAutn)
	}

	m.GmmMessage.AuthenticationFailure = authenticationFailure

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
