package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type ULNASTransport struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.ULNASTRANSPORTMessageIdentity
	nasType.SpareHalfOctetAndPayloadContainerType
	nasType.PayloadContainer
	*nasType.PduSessionID2Value
	*nasType.OldPDUSessionID
	*nasType.RequestType
	*nasType.SNSSAI
	*nasType.DNN
	*nasType.AdditionalInformation
}

func NewULNASTransport(iei uint8) (uLNASTransport *ULNASTransport) {
	uLNASTransport = &ULNASTransport{}
	return uLNASTransport
}

const (
	ULNASTransportPduSessionID2ValueType    uint8 = 0x12
	ULNASTransportOldPDUSessionIDType       uint8 = 0x59
	ULNASTransportRequestTypeType           uint8 = 0x08
	ULNASTransportSNSSAIType                uint8 = 0x22
	ULNASTransportDNNType                   uint8 = 0x25
	ULNASTransportAdditionalInformationType uint8 = 0x24
)

func (a *ULNASTransport) EncodeULNASTransport(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.ULNASTRANSPORTMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndPayloadContainerType.Octet)
	binary.Write(buffer, binary.BigEndian, a.PayloadContainer.GetLen())
	binary.Write(buffer, binary.BigEndian, &a.PayloadContainer.Buffer)
	if a.PduSessionID2Value != nil {
		binary.Write(buffer, binary.BigEndian, a.PduSessionID2Value.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.PduSessionID2Value.Octet)
	}
	if a.OldPDUSessionID != nil {
		binary.Write(buffer, binary.BigEndian, a.OldPDUSessionID.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.OldPDUSessionID.Octet)
	}
	if a.RequestType != nil {
		binary.Write(buffer, binary.BigEndian, &a.RequestType.Octet)
	}
	if a.SNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.SNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.SNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, a.SNSSAI.Octet[:a.SNSSAI.GetLen()])
	}
	if a.DNN != nil {
		binary.Write(buffer, binary.BigEndian, a.DNN.GetIei())
		binary.Write(buffer, binary.BigEndian, a.DNN.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.DNN.Buffer)
	}
	if a.AdditionalInformation != nil {
		binary.Write(buffer, binary.BigEndian, a.AdditionalInformation.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AdditionalInformation.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.AdditionalInformation.Buffer)
	}
}

func (a *ULNASTransport) DecodeULNASTransport(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.ULNASTRANSPORTMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndPayloadContainerType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PayloadContainer.Len)
	a.PayloadContainer.SetLen(a.PayloadContainer.GetLen())
	binary.Read(buffer, binary.BigEndian, &a.PayloadContainer.Buffer)
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
		case ULNASTransportPduSessionID2ValueType:
			a.PduSessionID2Value = nasType.NewPduSessionID2Value(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PduSessionID2Value.Octet)
		case ULNASTransportOldPDUSessionIDType:
			a.OldPDUSessionID = nasType.NewOldPDUSessionID(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.OldPDUSessionID.Octet)
		case ULNASTransportRequestTypeType:
			a.RequestType = nasType.NewRequestType(ieiN)
			a.RequestType.Octet = ieiN
		case ULNASTransportSNSSAIType:
			a.SNSSAI = nasType.NewSNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.SNSSAI.Len)
			a.SNSSAI.SetLen(a.SNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.SNSSAI.Octet[:a.SNSSAI.GetLen()])
		case ULNASTransportDNNType:
			a.DNN = nasType.NewDNN(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.DNN.Len)
			a.DNN.SetLen(a.DNN.GetLen())
			binary.Read(buffer, binary.BigEndian, a.DNN.Buffer[:a.DNN.GetLen()])
		case ULNASTransportAdditionalInformationType:
			a.AdditionalInformation = nasType.NewAdditionalInformation(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AdditionalInformation.Len)
			a.AdditionalInformation.SetLen(a.AdditionalInformation.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AdditionalInformation.Buffer[:a.AdditionalInformation.GetLen()])
		default:
		}
	}
}
