package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type DeregistrationRequestUETerminatedDeregistration struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.DeregistrationRequestMessageIdentity
	nasType.SpareHalfOctetAndDeregistrationType
	*nasType.Cause5GMM
	*nasType.T3346Value
}

func NewDeregistrationRequestUETerminatedDeregistration(iei uint8) (deregistrationRequestUETerminatedDeregistration *DeregistrationRequestUETerminatedDeregistration) {
	deregistrationRequestUETerminatedDeregistration = &DeregistrationRequestUETerminatedDeregistration{}
	return deregistrationRequestUETerminatedDeregistration
}

const (
	DeregistrationRequestUETerminatedDeregistrationCause5GMMType  uint8 = 0x58
	DeregistrationRequestUETerminatedDeregistrationT3346ValueType uint8 = 0x5F
)

func (a *DeregistrationRequestUETerminatedDeregistration) EncodeDeregistrationRequestUETerminatedDeregistration(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.DeregistrationRequestMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndDeregistrationType.Octet)
	if a.Cause5GMM != nil {
		binary.Write(buffer, binary.BigEndian, a.Cause5GMM.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.Cause5GMM.Octet)
	}
	if a.T3346Value != nil {
		binary.Write(buffer, binary.BigEndian, a.T3346Value.GetIei())
		binary.Write(buffer, binary.BigEndian, a.T3346Value.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.T3346Value.Octet)
	}
}

func (a *DeregistrationRequestUETerminatedDeregistration) DecodeDeregistrationRequestUETerminatedDeregistration(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.DeregistrationRequestMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndDeregistrationType.Octet)
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
		case DeregistrationRequestUETerminatedDeregistrationCause5GMMType:
			a.Cause5GMM = nasType.NewCause5GMM(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Cause5GMM.Octet)
		case DeregistrationRequestUETerminatedDeregistrationT3346ValueType:
			a.T3346Value = nasType.NewT3346Value(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.T3346Value.Len)
			a.T3346Value.SetLen(a.T3346Value.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.T3346Value.Octet)
		default:
		}
	}
}
