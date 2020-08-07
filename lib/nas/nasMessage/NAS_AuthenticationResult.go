package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type AuthenticationResult struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.AuthenticationResultMessageIdentity
	nasType.SpareHalfOctetAndNgksi
	nasType.EAPMessage
	*nasType.ABBA
}

func NewAuthenticationResult(iei uint8) (authenticationResult *AuthenticationResult) {
	authenticationResult = &AuthenticationResult{}
	return authenticationResult
}

const (
	AuthenticationResultABBAType uint8 = 0x38
)

func (a *AuthenticationResult) EncodeAuthenticationResult(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.AuthenticationResultMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndNgksi.Octet)
	binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetLen())
	binary.Write(buffer, binary.BigEndian, &a.EAPMessage.Buffer)
	if a.ABBA != nil {
		binary.Write(buffer, binary.BigEndian, a.ABBA.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ABBA.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ABBA.Buffer)
	}
}

func (a *AuthenticationResult) DecodeAuthenticationResult(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.AuthenticationResultMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndNgksi.Octet)
	binary.Read(buffer, binary.BigEndian, &a.EAPMessage.Len)
	a.EAPMessage.SetLen(a.EAPMessage.GetLen())
	binary.Read(buffer, binary.BigEndian, &a.EAPMessage.Buffer)
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
		case AuthenticationResultABBAType:
			a.ABBA = nasType.NewABBA(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ABBA.Len)
			a.ABBA.SetLen(a.ABBA.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ABBA.Buffer[:a.ABBA.GetLen()])
		default:
		}
	}
}
