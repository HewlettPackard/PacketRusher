package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type RegistrationRequest struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.RegistrationRequestMessageIdentity
	nasType.NgksiAndRegistrationType5GS
	nasType.MobileIdentity5GS
	*nasType.NoncurrentNativeNASKeySetIdentifier
	*nasType.Capability5GMM
	*nasType.UESecurityCapability
	*nasType.RequestedNSSAI
	*nasType.LastVisitedRegisteredTAI
	*nasType.S1UENetworkCapability
	*nasType.UplinkDataStatus
	*nasType.PDUSessionStatus
	*nasType.MICOIndication
	*nasType.UEStatus
	*nasType.AdditionalGUTI
	*nasType.AllowedPDUSessionStatus
	*nasType.UesUsageSetting
	*nasType.RequestedDRXParameters
	*nasType.EPSNASMessageContainer
	*nasType.LADNIndication
	*nasType.PayloadContainer
	*nasType.NetworkSlicingIndication
	*nasType.UpdateType5GS
	*nasType.NASMessageContainer
}

func NewRegistrationRequest(iei uint8) (registrationRequest *RegistrationRequest) {
	registrationRequest = &RegistrationRequest{}
	return registrationRequest
}

const (
	RegistrationRequestNoncurrentNativeNASKeySetIdentifierType uint8 = 0x0C
	RegistrationRequestCapability5GMMType                      uint8 = 0x10
	RegistrationRequestUESecurityCapabilityType                uint8 = 0x2E
	RegistrationRequestRequestedNSSAIType                      uint8 = 0x2F
	RegistrationRequestLastVisitedRegisteredTAIType            uint8 = 0x52
	RegistrationRequestS1UENetworkCapabilityType               uint8 = 0x17
	RegistrationRequestUplinkDataStatusType                    uint8 = 0x40
	RegistrationRequestPDUSessionStatusType                    uint8 = 0x50
	RegistrationRequestMICOIndicationType                      uint8 = 0x0B
	RegistrationRequestUEStatusType                            uint8 = 0x2B
	RegistrationRequestAdditionalGUTIType                      uint8 = 0x77
	RegistrationRequestAllowedPDUSessionStatusType             uint8 = 0x25
	RegistrationRequestUesUsageSettingType                     uint8 = 0x18
	RegistrationRequestRequestedDRXParametersType              uint8 = 0x51
	RegistrationRequestEPSNASMessageContainerType              uint8 = 0x70
	RegistrationRequestLADNIndicationType                      uint8 = 0x74
	RegistrationRequestPayloadContainerType                    uint8 = 0x7B
	RegistrationRequestNetworkSlicingIndicationType            uint8 = 0x09
	RegistrationRequestUpdateType5GSType                       uint8 = 0x53
	RegistrationRequestNASMessageContainerType                 uint8 = 0x71
)

func (a *RegistrationRequest) EncodeRegistrationRequest(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.RegistrationRequestMessageIdentity.Octet)
	binary.Write(buffer, binary.BigEndian, &a.NgksiAndRegistrationType5GS.Octet)
	binary.Write(buffer, binary.BigEndian, a.MobileIdentity5GS.GetLen())
	binary.Write(buffer, binary.BigEndian, &a.MobileIdentity5GS.Buffer)
	if a.NoncurrentNativeNASKeySetIdentifier != nil {
		binary.Write(buffer, binary.BigEndian, &a.NoncurrentNativeNASKeySetIdentifier.Octet)
	}
	if a.Capability5GMM != nil {
		binary.Write(buffer, binary.BigEndian, a.Capability5GMM.GetIei())
		binary.Write(buffer, binary.BigEndian, a.Capability5GMM.GetLen())
		binary.Write(buffer, binary.BigEndian, a.Capability5GMM.Octet[:a.Capability5GMM.GetLen()])
	}
	if a.UESecurityCapability != nil {
		binary.Write(buffer, binary.BigEndian, a.UESecurityCapability.GetIei())
		binary.Write(buffer, binary.BigEndian, a.UESecurityCapability.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.UESecurityCapability.Buffer)
	}
	if a.RequestedNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.RequestedNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RequestedNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RequestedNSSAI.Buffer)
	}
	if a.LastVisitedRegisteredTAI != nil {
		binary.Write(buffer, binary.BigEndian, a.LastVisitedRegisteredTAI.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.LastVisitedRegisteredTAI.Octet)
	}
	if a.S1UENetworkCapability != nil {
		binary.Write(buffer, binary.BigEndian, a.S1UENetworkCapability.GetIei())
		binary.Write(buffer, binary.BigEndian, a.S1UENetworkCapability.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.S1UENetworkCapability.Buffer)
	}
	if a.UplinkDataStatus != nil {
		binary.Write(buffer, binary.BigEndian, a.UplinkDataStatus.GetIei())
		binary.Write(buffer, binary.BigEndian, a.UplinkDataStatus.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.UplinkDataStatus.Buffer)
	}
	if a.PDUSessionStatus != nil {
		binary.Write(buffer, binary.BigEndian, a.PDUSessionStatus.GetIei())
		binary.Write(buffer, binary.BigEndian, a.PDUSessionStatus.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.PDUSessionStatus.Buffer)
	}
	if a.MICOIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.MICOIndication.Octet)
	}
	if a.UEStatus != nil {
		binary.Write(buffer, binary.BigEndian, a.UEStatus.GetIei())
		binary.Write(buffer, binary.BigEndian, a.UEStatus.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.UEStatus.Octet)
	}
	if a.AdditionalGUTI != nil {
		binary.Write(buffer, binary.BigEndian, a.AdditionalGUTI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AdditionalGUTI.GetLen())
		binary.Write(buffer, binary.BigEndian, a.AdditionalGUTI.Octet[:a.AdditionalGUTI.GetLen()])
	}
	if a.AllowedPDUSessionStatus != nil {
		binary.Write(buffer, binary.BigEndian, a.AllowedPDUSessionStatus.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AllowedPDUSessionStatus.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.AllowedPDUSessionStatus.Buffer)
	}
	if a.UesUsageSetting != nil {
		binary.Write(buffer, binary.BigEndian, a.UesUsageSetting.GetIei())
		binary.Write(buffer, binary.BigEndian, a.UesUsageSetting.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.UesUsageSetting.Octet)
	}
	if a.RequestedDRXParameters != nil {
		binary.Write(buffer, binary.BigEndian, a.RequestedDRXParameters.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RequestedDRXParameters.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RequestedDRXParameters.Octet)
	}
	if a.EPSNASMessageContainer != nil {
		binary.Write(buffer, binary.BigEndian, a.EPSNASMessageContainer.GetIei())
		binary.Write(buffer, binary.BigEndian, a.EPSNASMessageContainer.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.EPSNASMessageContainer.Buffer)
	}
	if a.LADNIndication != nil {
		binary.Write(buffer, binary.BigEndian, a.LADNIndication.GetIei())
		binary.Write(buffer, binary.BigEndian, a.LADNIndication.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.LADNIndication.Buffer)
	}
	if a.PayloadContainer != nil {
		binary.Write(buffer, binary.BigEndian, a.PayloadContainer.GetIei())
		binary.Write(buffer, binary.BigEndian, a.PayloadContainer.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.PayloadContainer.Buffer)
	}
	if a.NetworkSlicingIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.NetworkSlicingIndication.Octet)
	}
	if a.UpdateType5GS != nil {
		binary.Write(buffer, binary.BigEndian, a.UpdateType5GS.GetIei())
		binary.Write(buffer, binary.BigEndian, a.UpdateType5GS.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.UpdateType5GS.Octet)
	}
	if a.NASMessageContainer != nil {
		binary.Write(buffer, binary.BigEndian, a.NASMessageContainer.GetIei())
		binary.Write(buffer, binary.BigEndian, a.NASMessageContainer.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.NASMessageContainer.Buffer)
	}
}

func (a *RegistrationRequest) DecodeRegistrationRequest(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.RegistrationRequestMessageIdentity.Octet)
	binary.Read(buffer, binary.BigEndian, &a.NgksiAndRegistrationType5GS.Octet)
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
		case RegistrationRequestNoncurrentNativeNASKeySetIdentifierType:
			a.NoncurrentNativeNASKeySetIdentifier = nasType.NewNoncurrentNativeNASKeySetIdentifier(ieiN)
			a.NoncurrentNativeNASKeySetIdentifier.Octet = ieiN
		case RegistrationRequestCapability5GMMType:
			a.Capability5GMM = nasType.NewCapability5GMM(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Capability5GMM.Len)
			a.Capability5GMM.SetLen(a.Capability5GMM.GetLen())
			binary.Read(buffer, binary.BigEndian, a.Capability5GMM.Octet[:a.Capability5GMM.GetLen()])
		case RegistrationRequestUESecurityCapabilityType:
			a.UESecurityCapability = nasType.NewUESecurityCapability(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UESecurityCapability.Len)
			a.UESecurityCapability.SetLen(a.UESecurityCapability.GetLen())
			binary.Read(buffer, binary.BigEndian, a.UESecurityCapability.Buffer[:a.UESecurityCapability.GetLen()])
		case RegistrationRequestRequestedNSSAIType:
			a.RequestedNSSAI = nasType.NewRequestedNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RequestedNSSAI.Len)
			a.RequestedNSSAI.SetLen(a.RequestedNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.RequestedNSSAI.Buffer[:a.RequestedNSSAI.GetLen()])
		case RegistrationRequestLastVisitedRegisteredTAIType:
			a.LastVisitedRegisteredTAI = nasType.NewLastVisitedRegisteredTAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.LastVisitedRegisteredTAI.Octet)
		case RegistrationRequestS1UENetworkCapabilityType:
			a.S1UENetworkCapability = nasType.NewS1UENetworkCapability(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.S1UENetworkCapability.Len)
			a.S1UENetworkCapability.SetLen(a.S1UENetworkCapability.GetLen())
			binary.Read(buffer, binary.BigEndian, a.S1UENetworkCapability.Buffer[:a.S1UENetworkCapability.GetLen()])
		case RegistrationRequestUplinkDataStatusType:
			a.UplinkDataStatus = nasType.NewUplinkDataStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UplinkDataStatus.Len)
			a.UplinkDataStatus.SetLen(a.UplinkDataStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, a.UplinkDataStatus.Buffer[:a.UplinkDataStatus.GetLen()])
		case RegistrationRequestPDUSessionStatusType:
			a.PDUSessionStatus = nasType.NewPDUSessionStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PDUSessionStatus.Len)
			a.PDUSessionStatus.SetLen(a.PDUSessionStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PDUSessionStatus.Buffer[:a.PDUSessionStatus.GetLen()])
		case RegistrationRequestMICOIndicationType:
			a.MICOIndication = nasType.NewMICOIndication(ieiN)
			a.MICOIndication.Octet = ieiN
		case RegistrationRequestUEStatusType:
			a.UEStatus = nasType.NewUEStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UEStatus.Len)
			a.UEStatus.SetLen(a.UEStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.UEStatus.Octet)
		case RegistrationRequestAdditionalGUTIType:
			a.AdditionalGUTI = nasType.NewAdditionalGUTI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AdditionalGUTI.Len)
			a.AdditionalGUTI.SetLen(a.AdditionalGUTI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AdditionalGUTI.Octet[:a.AdditionalGUTI.GetLen()])
		case RegistrationRequestAllowedPDUSessionStatusType:
			a.AllowedPDUSessionStatus = nasType.NewAllowedPDUSessionStatus(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AllowedPDUSessionStatus.Len)
			a.AllowedPDUSessionStatus.SetLen(a.AllowedPDUSessionStatus.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AllowedPDUSessionStatus.Buffer[:a.AllowedPDUSessionStatus.GetLen()])
		case RegistrationRequestUesUsageSettingType:
			a.UesUsageSetting = nasType.NewUesUsageSetting(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UesUsageSetting.Len)
			a.UesUsageSetting.SetLen(a.UesUsageSetting.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.UesUsageSetting.Octet)
		case RegistrationRequestRequestedDRXParametersType:
			a.RequestedDRXParameters = nasType.NewRequestedDRXParameters(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RequestedDRXParameters.Len)
			a.RequestedDRXParameters.SetLen(a.RequestedDRXParameters.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.RequestedDRXParameters.Octet)
		case RegistrationRequestEPSNASMessageContainerType:
			a.EPSNASMessageContainer = nasType.NewEPSNASMessageContainer(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.EPSNASMessageContainer.Len)
			a.EPSNASMessageContainer.SetLen(a.EPSNASMessageContainer.GetLen())
			binary.Read(buffer, binary.BigEndian, a.EPSNASMessageContainer.Buffer[:a.EPSNASMessageContainer.GetLen()])
		case RegistrationRequestLADNIndicationType:
			a.LADNIndication = nasType.NewLADNIndication(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.LADNIndication.Len)
			a.LADNIndication.SetLen(a.LADNIndication.GetLen())
			binary.Read(buffer, binary.BigEndian, a.LADNIndication.Buffer[:a.LADNIndication.GetLen()])
		case RegistrationRequestPayloadContainerType:
			a.PayloadContainer = nasType.NewPayloadContainer(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.PayloadContainer.Len)
			a.PayloadContainer.SetLen(a.PayloadContainer.GetLen())
			binary.Read(buffer, binary.BigEndian, a.PayloadContainer.Buffer[:a.PayloadContainer.GetLen()])
		case RegistrationRequestNetworkSlicingIndicationType:
			a.NetworkSlicingIndication = nasType.NewNetworkSlicingIndication(ieiN)
			a.NetworkSlicingIndication.Octet = ieiN
		case RegistrationRequestUpdateType5GSType:
			a.UpdateType5GS = nasType.NewUpdateType5GS(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UpdateType5GS.Len)
			a.UpdateType5GS.SetLen(a.UpdateType5GS.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.UpdateType5GS.Octet)
		case RegistrationRequestNASMessageContainerType:
			a.NASMessageContainer = nasType.NewNASMessageContainer(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.NASMessageContainer.Len)
			a.NASMessageContainer.SetLen(a.NASMessageContainer.GetLen())
			binary.Read(buffer, binary.BigEndian, a.NASMessageContainer.Buffer[:a.NASMessageContainer.GetLen()])
		default:
		}
	}
}
