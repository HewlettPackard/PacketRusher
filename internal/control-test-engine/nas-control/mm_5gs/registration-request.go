package mm_5gs

import (
	"bytes"
	"fmt"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
)

func GetRegistrationRequestWith5GMM(registrationType uint8, mobileIdentity nasType.MobileIdentity5GS, requestedNSSAI *nasType.RequestedNSSAI, uplinkDataStatus *nasType.UplinkDataStatus, ueSecurityCapability *nasType.UESecurityCapability) (nasPdu []byte) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeRegistrationRequest)

	registrationRequest := nasMessage.NewRegistrationRequest(0)
	registrationRequest.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	registrationRequest.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	registrationRequest.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0x00)
	registrationRequest.RegistrationRequestMessageIdentity.SetMessageType(nas.MsgTypeRegistrationRequest)
	registrationRequest.NgksiAndRegistrationType5GS.SetTSC(nasMessage.TypeOfSecurityContextFlagNative)
	registrationRequest.NgksiAndRegistrationType5GS.SetNasKeySetIdentifiler(0x01)
	registrationRequest.NgksiAndRegistrationType5GS.SetRegistrationType5GS(registrationType)
	registrationRequest.MobileIdentity5GS = mobileIdentity
	registrationRequest.Capability5GMM = &nasType.Capability5GMM{
		Iei:   nasMessage.RegistrationRequestCapability5GMMType,
		Len:   1,
		Octet: [13]uint8{0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	registrationRequest.UESecurityCapability = ueSecurityCapability
	registrationRequest.RequestedNSSAI = requestedNSSAI
	registrationRequest.UplinkDataStatus = uplinkDataStatus

	registrationRequest.SetFOR(1)

	m.GmmMessage.RegistrationRequest = registrationRequest

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
