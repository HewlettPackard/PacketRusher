package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type RegistrationAccept struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.RegistrationAcceptMessageIdentity
	nasType.RegistrationResult5GS
	*nasType.GUTI5G
	*nasType.EquivalentPlmns
	*nasType.TAIList
	*nasType.AllowedNSSAI
	*nasType.RejectedNSSAI
	*nasType.ConfiguredNSSAI
	*nasType.NetworkFeatureSupport5GS
	*nasType.PDUSessionStatus
	*nasType.PDUSessionReactivationResult
	*nasType.PDUSessionReactivationResultErrorCause
	*nasType.LADNInformation
	*nasType.MICOIndication
	*nasType.NetworkSlicingIndication
	*nasType.ServiceAreaList
	*nasType.T3512Value
	*nasType.Non3GppDeregistrationTimerValue
	*nasType.T3502Value
	*nasType.EmergencyNumberList
	*nasType.ExtendedEmergencyNumberList
	*nasType.SORTransparentContainer
	*nasType.EAPMessage
	*nasType.NSSAIInclusionMode
	*nasType.OperatordefinedAccessCategoryDefinitions
	*nasType.NegotiatedDRXParameters
}

func NewRegistrationAccept(iei uint8) (registrationAccept *RegistrationAccept) {
	registrationAccept = &RegistrationAccept{}
	return registrationAccept
}

const (
	RegistrationAcceptGUTI5GType                                   uint8 = 0x77
	RegistrationAcceptEquivalentPlmnsType                          uint8 = 0x4A
	RegistrationAcceptTAIListType                                  uint8 = 0x54
	RegistrationAcceptAllowedNSSAIType                             uint8 = 0x15
	RegistrationAcceptRejectedNSSAIType                            uint8 = 0x11
	RegistrationAcceptConfiguredNSSAIType                          uint8 = 0x31
	RegistrationAcceptNetworkFeatureSupport5GSType                 uint8 = 0x21
	RegistrationAcceptPDUSessionStatusType                         uint8 = 0x50
	RegistrationAcceptPDUSessionReactivationResultType             uint8 = 0x26
	RegistrationAcceptPDUSessionReactivationResultErrorCauseType   uint8 = 0x72
	RegistrationAcceptLADNInformationType                          uint8 = 0x79
	RegistrationAcceptMICOIndicationType                           uint8 = 0x0B
	RegistrationAcceptNetworkSlicingIndicationType                 uint8 = 0x09
	RegistrationAcceptServiceAreaListType                          uint8 = 0x27
	RegistrationAcceptT3512ValueType                               uint8 = 0x5E
	RegistrationAcceptNon3GppDeregistrationTimerValueType          uint8 = 0x5D
	RegistrationAcceptT3502ValueType                               uint8 = 0x16
	RegistrationAcceptEmergencyNumberListType                      uint8 = 0x34
	RegistrationAcceptExtendedEmergencyNumberListType              uint8 = 0x7A
	RegistrationAcceptSORTransparentContainerType                  uint8 = 0x73
	RegistrationAcceptEAPMessageType                               uint8 = 0x78
	RegistrationAcceptNSSAIInclusionModeType                       uint8 = 0x0A
	RegistrationAcceptOperatordefinedAccessCategoryDefinitionsType uint8 = 0x76
	RegistrationAcceptNegotiatedDRXParametersType                  uint8 = 0x51
)

func (a *RegistrationAccept) EncodeRegistrationAccept(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.RegistrationAcceptMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, a.RegistrationResult5GS.GetLen())
	binary.Write(buffer, binary.BigEndian, &a.RegistrationResult5GS.Octet)
	if a.GUTI5G != nil {
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.GetIei())
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.GetLen())
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.Octet[:a.GUTI5G.GetLen()])
	}
	if a.EquivalentPlmns != nil {
		binary.Write(buffer, binary.BigEndian, a.EquivalentPlmns.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EquivalentPlmns.GetLen())
		binary.Write(buffer, binary.BigEndian, a.EquivalentPlmns.Octet[:a.EquivalentPlmns.GetLen()])
	}
	if a.TAIList != nil {
		binary.Write(buffer, binary.BigEndian, a.TAIList.GetIei())
		binary.Write(buffer, binary.BigEndian, a.TAIList.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.TAIList.Buffer)
	}
	if a.AllowedNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.AllowedNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AllowedNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.AllowedNSSAI.Buffer)
	}
	if a.RejectedNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.RejectedNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RejectedNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RejectedNSSAI.Buffer)
	}
	if a.ConfiguredNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.ConfiguredNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ConfiguredNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ConfiguredNSSAI.Buffer)
	}
	if a.NetworkFeatureSupport5GS != nil {
		binary.Write(buffer, binary.BigEndian, a.NetworkFeatureSupport5GS.GetIei())
		binary.Write(buffer, binary.BigEndian, a.NetworkFeatureSupport5GS.GetLen())
		binary.Write(buffer, binary.BigEndian, a.NetworkFeatureSupport5GS.Octet[:a.NetworkFeatureSupport5GS.GetLen()])
	}
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
	if a.LADNInformation != nil {
		binary.Write(buffer, binary.BigEndian, a.LADNInformation.GetIei())
		binary.Write(buffer, binary.BigEndian, a.LADNInformation.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.LADNInformation.Buffer)
	}
	if a.MICOIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.MICOIndication.Octet)
	}
	if a.NetworkSlicingIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.NetworkSlicingIndication.Octet)
	}
	if a.ServiceAreaList != nil {
		binary.Write(buffer, binary.BigEndian, a.ServiceAreaList.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ServiceAreaList.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ServiceAreaList.Buffer)
	}
	if a.T3512Value != nil {
		binary.Write(buffer, binary.BigEndian, a.T3512Value.GetIei())
		binary.Write(buffer, binary.BigEndian, a.T3512Value.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.T3512Value.Octet)
	}
	if a.Non3GppDeregistrationTimerValue != nil {
		binary.Write(buffer, binary.BigEndian, a.Non3GppDeregistrationTimerValue.GetIei())
		binary.Write(buffer, binary.BigEndian, a.Non3GppDeregistrationTimerValue.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.Non3GppDeregistrationTimerValue.Octet)
	}
	if a.T3502Value != nil {
		binary.Write(buffer, binary.BigEndian, a.T3502Value.GetIei())
		binary.Write(buffer, binary.BigEndian, a.T3502Value.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.T3502Value.Octet)
	}
	if a.EmergencyNumberList != nil {
		binary.Write(buffer, binary.BigEndian, a.EmergencyNumberList.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EmergencyNumberList.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.EmergencyNumberList.Buffer)
	}
	if a.ExtendedEmergencyNumberList != nil {
		binary.Write(buffer, binary.BigEndian, a.ExtendedEmergencyNumberList.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ExtendedEmergencyNumberList.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ExtendedEmergencyNumberList.Buffer)
	}
	if a.SORTransparentContainer != nil {
		binary.Write(buffer, binary.BigEndian, a.SORTransparentContainer.GetIei())
		binary.Write(buffer, binary.BigEndian, a.SORTransparentContainer.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.SORTransparentContainer.Buffer)
	}
	if a.EAPMessage != nil {
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EAPMessage.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.EAPMessage.Buffer)
	}
	if a.NSSAIInclusionMode != nil {
		binary.Write(buffer, binary.BigEndian, &a.NSSAIInclusionMode.Octet)
	}
	if a.OperatordefinedAccessCategoryDefinitions != nil {
		binary.Write(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.OperatordefinedAccessCategoryDefinitions.Buffer)
	}
	if a.NegotiatedDRXParameters != nil {
		binary.Write(buffer, binary.BigEndian, a.NegotiatedDRXParameters.GetIei())
		binary.Write(buffer, binary.BigEndian, a.NegotiatedDRXParameters.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.NegotiatedDRXParameters.Octet)
	}
}

func (a *RegistrationAccept) DecodeRegistrationAccept(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.RegistrationAcceptMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.RegistrationResult5GS.Len)
	a.RegistrationResult5GS.SetLen(a.RegistrationResult5GS.GetLen())
	binary.Read(buffer, binary.BigEndian, &a.RegistrationResult5GS.Octet)
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
		case RegistrationAcceptGUTI5GType:
			a.GUTI5G = nasType.NewGUTI5G(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.GUTI5G.Len)
			a.GUTI5G.SetLen(a.GUTI5G.GetLen())
			binary.Read(buffer, binary.BigEndian, a.GUTI5G.Octet[:a.GUTI5G.GetLen()])
		case RegistrationAcceptEquivalentPlmnsType:
			a.EquivalentPlmns = nasType.NewEquivalentPlmns(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EquivalentPlmns.Len)
			a.EquivalentPlmns.SetLen(a.EquivalentPlmns.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EquivalentPlmns.Octet[:a.EquivalentPlmns.GetLen()])
		case RegistrationAcceptTAIListType:
			a.TAIList = nasType.NewTAIList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.TAIList.Len)
			a.TAIList.SetLen(a.TAIList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.TAIList.Buffer[:a.TAIList.GetLen()])
		case RegistrationAcceptAllowedNSSAIType:
			a.AllowedNSSAI = nasType.NewAllowedNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AllowedNSSAI.Len)
			a.AllowedNSSAI.SetLen(a.AllowedNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AllowedNSSAI.Buffer[:a.AllowedNSSAI.GetLen()])
		case RegistrationAcceptRejectedNSSAIType:
			a.RejectedNSSAI = nasType.NewRejectedNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RejectedNSSAI.Len)
			a.RejectedNSSAI.SetLen(a.RejectedNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.RejectedNSSAI.Buffer[:a.RejectedNSSAI.GetLen()])
		case RegistrationAcceptConfiguredNSSAIType:
			a.ConfiguredNSSAI = nasType.NewConfiguredNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ConfiguredNSSAI.Len)
			a.ConfiguredNSSAI.SetLen(a.ConfiguredNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ConfiguredNSSAI.Buffer[:a.ConfiguredNSSAI.GetLen()])
		case RegistrationAcceptNetworkFeatureSupport5GSType:
			a.NetworkFeatureSupport5GS = nasType.NewNetworkFeatureSupport5GS(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.NetworkFeatureSupport5GS.Len)
			a.NetworkFeatureSupport5GS.SetLen(a.NetworkFeatureSupport5GS.GetLen())
			binary.Read(buffer, binary.BigEndian, a.NetworkFeatureSupport5GS.Octet[:a.NetworkFeatureSupport5GS.GetLen()])
		case RegistrationAcceptPDUSessionStatusType:
			a.PDUSessionStatus = nasType.NewPDUSessionStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionStatus.Len)
			a.PDUSessionStatus.SetLen(a.PDUSessionStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionStatus.Buffer[:a.PDUSessionStatus.GetLen()])
		case RegistrationAcceptPDUSessionReactivationResultType:
			a.PDUSessionReactivationResult = nasType.NewPDUSessionReactivationResult(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionReactivationResult.Len)
			a.PDUSessionReactivationResult.SetLen(a.PDUSessionReactivationResult.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionReactivationResult.Buffer[:a.PDUSessionReactivationResult.GetLen()])
		case RegistrationAcceptPDUSessionReactivationResultErrorCauseType:
			a.PDUSessionReactivationResultErrorCause = nasType.NewPDUSessionReactivationResultErrorCause(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionReactivationResultErrorCause.Len)
			a.PDUSessionReactivationResultErrorCause.SetLen(a.PDUSessionReactivationResultErrorCause.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionReactivationResultErrorCause.Buffer[:a.PDUSessionReactivationResultErrorCause.GetLen()])
		case RegistrationAcceptLADNInformationType:
			a.LADNInformation = nasType.NewLADNInformation(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.LADNInformation.Len)
			a.LADNInformation.SetLen(a.LADNInformation.GetLen())
			binary.Read(buffer, binary.BigEndian, a.LADNInformation.Buffer[:a.LADNInformation.GetLen()])
		case RegistrationAcceptMICOIndicationType:
			a.MICOIndication = nasType.NewMICOIndication(ieiN)
			a.MICOIndication.Octet = ieiN
		case RegistrationAcceptNetworkSlicingIndicationType:
			a.NetworkSlicingIndication = nasType.NewNetworkSlicingIndication(ieiN)
			a.NetworkSlicingIndication.Octet = ieiN
		case RegistrationAcceptServiceAreaListType:
			a.ServiceAreaList = nasType.NewServiceAreaList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ServiceAreaList.Len)
			a.ServiceAreaList.SetLen(a.ServiceAreaList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ServiceAreaList.Buffer[:a.ServiceAreaList.GetLen()])
		case RegistrationAcceptT3512ValueType:
			a.T3512Value = nasType.NewT3512Value(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.T3512Value.Len)
			a.T3512Value.SetLen(a.T3512Value.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.T3512Value.Octet)
		case RegistrationAcceptNon3GppDeregistrationTimerValueType:
			a.Non3GppDeregistrationTimerValue = nasType.NewNon3GppDeregistrationTimerValue(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Non3GppDeregistrationTimerValue.Len)
			a.Non3GppDeregistrationTimerValue.SetLen(a.Non3GppDeregistrationTimerValue.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.Non3GppDeregistrationTimerValue.Octet)
		case RegistrationAcceptT3502ValueType:
			a.T3502Value = nasType.NewT3502Value(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.T3502Value.Len)
			a.T3502Value.SetLen(a.T3502Value.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.T3502Value.Octet)
		case RegistrationAcceptEmergencyNumberListType:
			a.EmergencyNumberList = nasType.NewEmergencyNumberList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EmergencyNumberList.Len)
			a.EmergencyNumberList.SetLen(a.EmergencyNumberList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EmergencyNumberList.Buffer[:a.EmergencyNumberList.GetLen()])
		case RegistrationAcceptExtendedEmergencyNumberListType:
			a.ExtendedEmergencyNumberList = nasType.NewExtendedEmergencyNumberList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ExtendedEmergencyNumberList.Len)
			a.ExtendedEmergencyNumberList.SetLen(a.ExtendedEmergencyNumberList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ExtendedEmergencyNumberList.Buffer[:a.ExtendedEmergencyNumberList.GetLen()])
		case RegistrationAcceptSORTransparentContainerType:
			a.SORTransparentContainer = nasType.NewSORTransparentContainer(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.SORTransparentContainer.Len)
			a.SORTransparentContainer.SetLen(a.SORTransparentContainer.GetLen())
			binary.Read(buffer, binary.BigEndian, a.SORTransparentContainer.Buffer[:a.SORTransparentContainer.GetLen()])
		case RegistrationAcceptEAPMessageType:
			a.EAPMessage = nasType.NewEAPMessage(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EAPMessage.Len)
			a.EAPMessage.SetLen(a.EAPMessage.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EAPMessage.Buffer[:a.EAPMessage.GetLen()])
		case RegistrationAcceptNSSAIInclusionModeType:
			a.NSSAIInclusionMode = nasType.NewNSSAIInclusionMode(ieiN)
			a.NSSAIInclusionMode.Octet = ieiN
		case RegistrationAcceptOperatordefinedAccessCategoryDefinitionsType:
			a.OperatordefinedAccessCategoryDefinitions = nasType.NewOperatordefinedAccessCategoryDefinitions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.OperatordefinedAccessCategoryDefinitions.Len)
			a.OperatordefinedAccessCategoryDefinitions.SetLen(a.OperatordefinedAccessCategoryDefinitions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.Buffer[:a.OperatordefinedAccessCategoryDefinitions.GetLen()])
		case RegistrationAcceptNegotiatedDRXParametersType:
			a.NegotiatedDRXParameters = nasType.NewNegotiatedDRXParameters(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.NegotiatedDRXParameters.Len)
			a.NegotiatedDRXParameters.SetLen(a.NegotiatedDRXParameters.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.NegotiatedDRXParameters.Octet)
		default:
		}
	}
}
