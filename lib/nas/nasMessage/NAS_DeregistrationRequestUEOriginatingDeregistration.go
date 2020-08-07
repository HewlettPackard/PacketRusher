package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type DeregistrationRequestUEOriginatingDeregistration struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.DeregistrationRequestMessageIdentity
	nasType.NgksiAndDeregistrationType
	nasType.MobileIdentity5GS
}

func NewDeregistrationRequestUEOriginatingDeregistration(iei uint8) (deregistrationRequestUEOriginatingDeregistration *DeregistrationRequestUEOriginatingDeregistration) {
	deregistrationRequestUEOriginatingDeregistration = &DeregistrationRequestUEOriginatingDeregistration{}
	return deregistrationRequestUEOriginatingDeregistration
}

func (a *DeregistrationRequestUEOriginatingDeregistration) EncodeDeregistrationRequestUEOriginatingDeregistration(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.DeregistrationRequestMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, &a.NgksiAndDeregistrationType.Octet)
	binary.Write(buffer, binary.BigEndian, a.MobileIdentity5GS.GetLen())
	binary.Write(buffer, binary.BigEndian, &a.MobileIdentity5GS.Buffer)
}

func (a *DeregistrationRequestUEOriginatingDeregistration) DecodeDeregistrationRequestUEOriginatingDeregistration(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.DeregistrationRequestMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.NgksiAndDeregistrationType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.MobileIdentity5GS.Len)
	a.MobileIdentity5GS.SetLen(a.MobileIdentity5GS.GetLen())
	binary.Read(buffer, binary.BigEndian, &a.MobileIdentity5GS.Buffer)
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
