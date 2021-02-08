package mm_5gs

import (
	"bytes"
	"fmt"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
)

func AuthenticationFailure(eapMsg string) (nasPdu []byte) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeAuthenticationFailure)

	authenticationFailure := nasMessage.NewAuthenticationFailure(0)
	authenticationFailure.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	authenticationFailure.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	authenticationFailure.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	authenticationFailure.AuthenticationFailureMessageIdentity.SetMessageType(nas.MsgTypeAuthenticationFailure)

	authenticationFailure.Cause5GMM.SetCauseValue(nasMessage.Cause5GMMMACFailure)

	m.GmmMessage.AuthenticationFailure = authenticationFailure

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
