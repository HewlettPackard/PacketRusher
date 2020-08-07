package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type SecurityProtected5GSNASMessage struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.MessageAuthenticationCode
	nasType.SequenceNumber
	nasType.Plain5GSNASMessage
}

func NewSecurityProtected5GSNASMessage(iei uint8) (securityProtected5GSNASMessage *SecurityProtected5GSNASMessage) {
	securityProtected5GSNASMessage = &SecurityProtected5GSNASMessage{}
	return securityProtected5GSNASMessage
}

func (a *SecurityProtected5GSNASMessage) EncodeSecurityProtected5GSNASMessage(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.MessageAuthenticationCode.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SequenceNumber.Octet)
	binary.Write(buffer, binary.BigEndian, &a.Plain5GSNASMessage)
}

func (a *SecurityProtected5GSNASMessage) DecodeSecurityProtected5GSNASMessage(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.MessageAuthenticationCode.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SequenceNumber.Octet)
	binary.Read(buffer, binary.BigEndian, &a.Plain5GSNASMessage)
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
		default:
		}
	}
}
