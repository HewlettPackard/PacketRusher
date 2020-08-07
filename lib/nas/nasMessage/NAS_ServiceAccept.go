package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type ServiceAccept struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.ServiceAcceptMessageIdentity
	*nasType.PDUSessionStatus
	*nasType.PDUSessionReactivationResult
	*nasType.PDUSessionReactivationResultErrorCause
	*nasType.EAPMessage
}

func NewServiceAccept(iei uint8) (serviceAccept *ServiceAccept) {
	serviceAccept = &ServiceAccept{}
	return serviceAccept
}

const (
	ServiceAcceptPDUSessionStatusType                       uint8 = 0x50
	ServiceAcceptPDUSessionReactivationResultType           uint8 = 0x26
	ServiceAcceptPDUSessionReactivationResultErrorCauseType uint8 = 0x72
	ServiceAcceptEAPMessageType                             uint8 = 0x78
)

func (a *ServiceAccept) EncodeServiceAccept(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.ServiceAcceptMessageIdentity.Octet)
	if a.PDUSessionStatus != nil {
		binary.Write(buffer, binary.BigEndian, a.PDUSessionStatus.GetIei())
		binary.Write(buffer, binary.BigEndian, a.PDUSessionStatus.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.PDUSessionStatus.Buffer)
	}
	if a.PDUSessionReactivationResult != nil {
		binary.Write(buffer, binary.BigEndian, a.PDUSessionReactivationResult.GetIei())
		binary.Write(buffer, binary.BigEndian, a.PDUSessionReactivationResult.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.PDUSessionReactivationResult.Buffer)
	}
	if a.PDUSessionReactivationResultErrorCause != nil {
		binary.Write(buffer, binary.BigEndian, a.PDUSessionReactivationResultErrorCause.GetIei())
		binary.Write(buffer, binary.BigEndian, a.PDUSessionReactivationResultErrorCause.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.PDUSessionReactivationResultErrorCause.Buffer)
	}
	if a.EAPMessage != nil {
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.EAPMessage.Buffer)
	}
}

func (a *ServiceAccept) DecodeServiceAccept(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.ServiceAcceptMessageIdentity.Octet)
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
		case ServiceAcceptPDUSessionStatusType:
			a.PDUSessionStatus = nasType.NewPDUSessionStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionStatus.Len)
			a.PDUSessionStatus.SetLen(a.PDUSessionStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionStatus.Buffer[:a.PDUSessionStatus.GetLen()])
		case ServiceAcceptPDUSessionReactivationResultType:
			a.PDUSessionReactivationResult = nasType.NewPDUSessionReactivationResult(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionReactivationResult.Len)
			a.PDUSessionReactivationResult.SetLen(a.PDUSessionReactivationResult.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionReactivationResult.Buffer[:a.PDUSessionReactivationResult.GetLen()])
		case ServiceAcceptPDUSessionReactivationResultErrorCauseType:
			a.PDUSessionReactivationResultErrorCause = nasType.NewPDUSessionReactivationResultErrorCause(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionReactivationResultErrorCause.Len)
			a.PDUSessionReactivationResultErrorCause.SetLen(a.PDUSessionReactivationResultErrorCause.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionReactivationResultErrorCause.Buffer[:a.PDUSessionReactivationResultErrorCause.GetLen()])
		case ServiceAcceptEAPMessageType:
			a.EAPMessage = nasType.NewEAPMessage(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EAPMessage.Len)
			a.EAPMessage.SetLen(a.EAPMessage.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EAPMessage.Buffer[:a.EAPMessage.GetLen()])
		default:
		}
	}
}
