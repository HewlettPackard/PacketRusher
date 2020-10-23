package nas_control

import (
	"bytes"
	"fmt"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
)

func GetRegistrationComplete(sorTransparentContainer []uint8) (nasPdu []byte) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeRegistrationComplete)

	registrationComplete := nasMessage.NewRegistrationComplete(0)
	registrationComplete.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	registrationComplete.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	registrationComplete.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	registrationComplete.RegistrationCompleteMessageIdentity.SetMessageType(nas.MsgTypeRegistrationComplete)

	if sorTransparentContainer != nil {
		registrationComplete.SORTransparentContainer = nasType.NewSORTransparentContainer(nasMessage.RegistrationCompleteSORTransparentContainerType)
		registrationComplete.SORTransparentContainer.SetLen(uint16(len(sorTransparentContainer)))
		registrationComplete.SORTransparentContainer.SetSORContent(sorTransparentContainer)
	}

	m.GmmMessage.RegistrationComplete = registrationComplete

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
