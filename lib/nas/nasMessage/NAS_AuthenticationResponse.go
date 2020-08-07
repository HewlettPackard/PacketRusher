package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type AuthenticationResponse struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.AuthenticationResponseMessageIdentity
	*nasType.AuthenticationResponseParameter
	*nasType.EAPMessage
}

func NewAuthenticationResponse(iei uint8) (authenticationResponse *AuthenticationResponse) {
	authenticationResponse = &AuthenticationResponse{}
	return authenticationResponse
}

const (
	AuthenticationResponseAuthenticationResponseParameterType uint8 = 0x2D
	AuthenticationResponseEAPMessageType                      uint8 = 0x78
)

func (a *AuthenticationResponse) EncodeAuthenticationResponse(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.AuthenticationResponseMessageIdentity.Octet)
	if a.AuthenticationResponseParameter != nil {
		binary.Write(buffer, binary.BigEndian, a.AuthenticationResponseParameter.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AuthenticationResponseParameter.GetLen())
		binary.Write(buffer, binary.BigEndian, a.AuthenticationResponseParameter.Octet[:a.AuthenticationResponseParameter.GetLen()])
	}
	if a.EAPMessage != nil {
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.EAPMessage.Buffer)
	}
}

func (a *AuthenticationResponse) DecodeAuthenticationResponse(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.AuthenticationResponseMessageIdentity.Octet)
	for buffer.Len() > 0 {
		var ieiN uint8
		var tmpIeiN uint8
		binary.Read(buffer, binary.BigEndian, &ieiN)
		// fmt.Println(ieiN)
		if ieiN >= 0x80 {
			tmpIeiN = (ieiN & 0xf0) >> 4
		} else {
			tmpIeiN = ieiN
		}
		// fmt.Println("type", tmpIeiN)
		switch tmpIeiN {
		case AuthenticationResponseAuthenticationResponseParameterType:
			a.AuthenticationResponseParameter = nasType.NewAuthenticationResponseParameter(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AuthenticationResponseParameter.Len)
			a.AuthenticationResponseParameter.SetLen(a.AuthenticationResponseParameter.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AuthenticationResponseParameter.Octet[:a.AuthenticationResponseParameter.GetLen()])
		case AuthenticationResponseEAPMessageType:
			a.EAPMessage = nasType.NewEAPMessage(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EAPMessage.Len)
			a.EAPMessage.SetLen(a.EAPMessage.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EAPMessage.Buffer[:a.EAPMessage.GetLen()])
		default:
		}
	}
}
