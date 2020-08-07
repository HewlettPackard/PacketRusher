package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type PDUSessionModificationRequest struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.PDUSessionID
	nasType.PTI
	nasType.PDUSESSIONMODIFICATIONREQUESTMessageIdentity
	*nasType.Capability5GSM
	*nasType.Cause5GSM
	*nasType.MaximumNumberOfSupportedPacketFilters
	*nasType.AlwaysonPDUSessionRequested
	*nasType.IntegrityProtectionMaximumDataRate
	*nasType.RequestedQosRules
	*nasType.RequestedQosFlowDescriptions
	*nasType.MappedEPSBearerContexts
	*nasType.ExtendedProtocolConfigurationOptions
}

func NewPDUSessionModificationRequest(iei uint8) (pDUSessionModificationRequest *PDUSessionModificationRequest) {
	pDUSessionModificationRequest = &PDUSessionModificationRequest{}
	return pDUSessionModificationRequest
}

const (
	PDUSessionModificationRequestCapability5GSMType                        uint8 = 0x28
	PDUSessionModificationRequestCause5GSMType                             uint8 = 0x59
	PDUSessionModificationRequestMaximumNumberOfSupportedPacketFiltersType uint8 = 0x55
	PDUSessionModificationRequestAlwaysonPDUSessionRequestedType           uint8 = 0x0B
	PDUSessionModificationRequestIntegrityProtectionMaximumDataRateType    uint8 = 0x13
	PDUSessionModificationRequestRequestedQosRulesType                     uint8 = 0x7A
	PDUSessionModificationRequestRequestedQosFlowDescriptionsType          uint8 = 0x79
	PDUSessionModificationRequestMappedEPSBearerContextsType               uint8 = 0x7F
	PDUSessionModificationRequestExtendedProtocolConfigurationOptionsType  uint8 = 0x7B
)

func (a *PDUSessionModificationRequest) EncodePDUSessionModificationRequest(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PDUSessionID.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PTI.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PDUSESSIONMODIFICATIONREQUESTMessageIdentity.Octet)
	if a.Capability5GSM != nil {
		binary.Write(buffer, binary.BigEndian, a.Capability5GSM.GetIei())
		binary.Write(buffer, binary.BigEndian, a.Capability5GSM.GetLen())
		binary.Write(buffer, binary.BigEndian, a.Capability5GSM.Octet[:a.Capability5GSM.GetLen()])
	}
	if a.Cause5GSM != nil {
		binary.Write(buffer, binary.BigEndian, a.Cause5GSM.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.Cause5GSM.Octet)
	}
	if a.MaximumNumberOfSupportedPacketFilters != nil {
		binary.Write(buffer, binary.BigEndian, a.MaximumNumberOfSupportedPacketFilters.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.MaximumNumberOfSupportedPacketFilters.Octet)
	}
	if a.AlwaysonPDUSessionRequested != nil {
		binary.Write(buffer, binary.BigEndian, &a.AlwaysonPDUSessionRequested.Octet)
	}
	if a.IntegrityProtectionMaximumDataRate != nil {
		binary.Write(buffer, binary.BigEndian, a.IntegrityProtectionMaximumDataRate.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.IntegrityProtectionMaximumDataRate.Octet)
	}
	if a.RequestedQosRules != nil {
		binary.Write(buffer, binary.BigEndian, a.RequestedQosRules.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RequestedQosRules.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RequestedQosRules.Buffer)
	}
	if a.RequestedQosFlowDescriptions != nil {
		binary.Write(buffer, binary.BigEndian, a.RequestedQosFlowDescriptions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RequestedQosFlowDescriptions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RequestedQosFlowDescriptions.Buffer)
	}
	if a.MappedEPSBearerContexts != nil {
		binary.Write(buffer, binary.BigEndian, a.MappedEPSBearerContexts.GetIei())
		binary.Write(buffer, binary.BigEndian, a.MappedEPSBearerContexts.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.MappedEPSBearerContexts.Buffer)
	}
	if a.ExtendedProtocolConfigurationOptions != nil {
		binary.Write(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolConfigurationOptions.Buffer)
	}
}

func (a *PDUSessionModificationRequest) DecodePDUSessionModificationRequest(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PDUSessionID.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PTI.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PDUSESSIONMODIFICATIONREQUESTMessageIdentity.Octet)
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
		case PDUSessionModificationRequestCapability5GSMType:
			a.Capability5GSM = nasType.NewCapability5GSM(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Capability5GSM.Len)
			a.Capability5GSM.SetLen(a.Capability5GSM.GetLen())
			binary.Read(buffer, binary.BigEndian, a.Capability5GSM.Octet[:a.Capability5GSM.GetLen()])
		case PDUSessionModificationRequestCause5GSMType:
			a.Cause5GSM = nasType.NewCause5GSM(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Cause5GSM.Octet)
		case PDUSessionModificationRequestMaximumNumberOfSupportedPacketFiltersType:
			a.MaximumNumberOfSupportedPacketFilters = nasType.NewMaximumNumberOfSupportedPacketFilters(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.MaximumNumberOfSupportedPacketFilters.Octet)
		case PDUSessionModificationRequestAlwaysonPDUSessionRequestedType:
			a.AlwaysonPDUSessionRequested = nasType.NewAlwaysonPDUSessionRequested(ieiN)
			a.AlwaysonPDUSessionRequested.Octet = ieiN
		case PDUSessionModificationRequestIntegrityProtectionMaximumDataRateType:
			a.IntegrityProtectionMaximumDataRate = nasType.NewIntegrityProtectionMaximumDataRate(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.IntegrityProtectionMaximumDataRate.Octet)
		case PDUSessionModificationRequestRequestedQosRulesType:
			a.RequestedQosRules = nasType.NewRequestedQosRules(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RequestedQosRules.Len)
			a.RequestedQosRules.SetLen(a.RequestedQosRules.GetLen())
			binary.Read(buffer, binary.BigEndian, a.RequestedQosRules.Buffer[:a.RequestedQosRules.GetLen()])
		case PDUSessionModificationRequestRequestedQosFlowDescriptionsType:
			a.RequestedQosFlowDescriptions = nasType.NewRequestedQosFlowDescriptions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RequestedQosFlowDescriptions.Len)
			a.RequestedQosFlowDescriptions.SetLen(a.RequestedQosFlowDescriptions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.RequestedQosFlowDescriptions.Buffer[:a.RequestedQosFlowDescriptions.GetLen()])
		case PDUSessionModificationRequestMappedEPSBearerContextsType:
			a.MappedEPSBearerContexts = nasType.NewMappedEPSBearerContexts(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.MappedEPSBearerContexts.Len)
			a.MappedEPSBearerContexts.SetLen(a.MappedEPSBearerContexts.GetLen())
			binary.Read(buffer, binary.BigEndian, a.MappedEPSBearerContexts.Buffer[:a.MappedEPSBearerContexts.GetLen()])
		case PDUSessionModificationRequestExtendedProtocolConfigurationOptionsType:
			a.ExtendedProtocolConfigurationOptions = nasType.NewExtendedProtocolConfigurationOptions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolConfigurationOptions.Len)
			a.ExtendedProtocolConfigurationOptions.SetLen(a.ExtendedProtocolConfigurationOptions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.Buffer[:a.ExtendedProtocolConfigurationOptions.GetLen()])
		default:
		}
	}
}
