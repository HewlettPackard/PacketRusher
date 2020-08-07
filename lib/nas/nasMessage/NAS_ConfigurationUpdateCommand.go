package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type ConfigurationUpdateCommand struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.SpareHalfOctetAndSecurityHeaderType
	nasType.ConfigurationUpdateCommandMessageIdentity
	*nasType.ConfigurationUpdateIndication
	*nasType.GUTI5G
	*nasType.TAIList
	*nasType.AllowedNSSAI
	*nasType.ServiceAreaList
	*nasType.FullNameForNetwork
	*nasType.ShortNameForNetwork
	*nasType.LocalTimeZone
	*nasType.UniversalTimeAndLocalTimeZone
	*nasType.NetworkDaylightSavingTime
	*nasType.LADNInformation
	*nasType.MICOIndication
	*nasType.NetworkSlicingIndication
	*nasType.ConfiguredNSSAI
	*nasType.RejectedNSSAI
	*nasType.OperatordefinedAccessCategoryDefinitions
	*nasType.SMSIndication
}

func NewConfigurationUpdateCommand(iei uint8) (configurationUpdateCommand *ConfigurationUpdateCommand) {
	configurationUpdateCommand = &ConfigurationUpdateCommand{}
	return configurationUpdateCommand
}

const (
	ConfigurationUpdateCommandConfigurationUpdateIndicationType            uint8 = 0x0D
	ConfigurationUpdateCommandGUTI5GType                                   uint8 = 0x77
	ConfigurationUpdateCommandTAIListType                                  uint8 = 0x54
	ConfigurationUpdateCommandAllowedNSSAIType                             uint8 = 0x15
	ConfigurationUpdateCommandServiceAreaListType                          uint8 = 0x27
	ConfigurationUpdateCommandFullNameForNetworkType                       uint8 = 0x43
	ConfigurationUpdateCommandShortNameForNetworkType                      uint8 = 0x45
	ConfigurationUpdateCommandLocalTimeZoneType                            uint8 = 0x46
	ConfigurationUpdateCommandUniversalTimeAndLocalTimeZoneType            uint8 = 0x47
	ConfigurationUpdateCommandNetworkDaylightSavingTimeType                uint8 = 0x49
	ConfigurationUpdateCommandLADNInformationType                          uint8 = 0x79
	ConfigurationUpdateCommandMICOIndicationType                           uint8 = 0x0B
	ConfigurationUpdateCommandNetworkSlicingIndicationType                 uint8 = 0x09
	ConfigurationUpdateCommandConfiguredNSSAIType                          uint8 = 0x31
	ConfigurationUpdateCommandRejectedNSSAIType                            uint8 = 0x11
	ConfigurationUpdateCommandOperatordefinedAccessCategoryDefinitionsType uint8 = 0x76
	ConfigurationUpdateCommandSMSIndicationType                            uint8 = 0x0F
)

func (a *ConfigurationUpdateCommand) EncodeConfigurationUpdateCommand(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Write(buffer, binary.BigEndian, &a.ConfigurationUpdateCommandMessageIdentity.Octet)
	if a.ConfigurationUpdateIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.ConfigurationUpdateIndication.Octet)
	}
	if a.GUTI5G != nil {
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.GetIei())
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.GetLen())
		binary.Write(buffer, binary.BigEndian, a.GUTI5G.Octet[:a.GUTI5G.GetLen()])
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
	if a.ServiceAreaList != nil {
		binary.Write(buffer, binary.BigEndian, a.ServiceAreaList.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ServiceAreaList.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ServiceAreaList.Buffer)
	}
	if a.FullNameForNetwork != nil {
		binary.Write(buffer, binary.BigEndian, a.FullNameForNetwork.GetIei())
		binary.Write(buffer, binary.BigEndian, a.FullNameForNetwork.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.FullNameForNetwork.Buffer)
	}
	if a.ShortNameForNetwork != nil {
		binary.Write(buffer, binary.BigEndian, a.ShortNameForNetwork.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ShortNameForNetwork.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ShortNameForNetwork.Buffer)
	}
	if a.LocalTimeZone != nil {
		binary.Write(buffer, binary.BigEndian, a.LocalTimeZone.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.LocalTimeZone.Octet)
	}
	if a.UniversalTimeAndLocalTimeZone != nil {
		binary.Write(buffer, binary.BigEndian, a.UniversalTimeAndLocalTimeZone.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.UniversalTimeAndLocalTimeZone.Octet)
	}
	if a.NetworkDaylightSavingTime != nil {
		binary.Write(buffer, binary.BigEndian, a.NetworkDaylightSavingTime.GetIei())
		binary.Write(buffer, binary.BigEndian, a.NetworkDaylightSavingTime.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.NetworkDaylightSavingTime.Octet)
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
	if a.ConfiguredNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.ConfiguredNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ConfiguredNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ConfiguredNSSAI.Buffer)
	}
	if a.RejectedNSSAI != nil {
		binary.Write(buffer, binary.BigEndian, a.RejectedNSSAI.GetIei())
		binary.Write(buffer, binary.BigEndian, a.RejectedNSSAI.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.RejectedNSSAI.Buffer)
	}
	if a.OperatordefinedAccessCategoryDefinitions != nil {
		binary.Write(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.OperatordefinedAccessCategoryDefinitions.Buffer)
	}
	if a.SMSIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.SMSIndication.Octet)
	}
}

func (a *ConfigurationUpdateCommand) DecodeConfigurationUpdateCommand(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.SpareHalfOctetAndSecurityHeaderType.Octet)
	binary.Read(buffer, binary.BigEndian, &a.ConfigurationUpdateCommandMessageIdentity.Octet)
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
		case ConfigurationUpdateCommandConfigurationUpdateIndicationType:
			a.ConfigurationUpdateIndication = nasType.NewConfigurationUpdateIndication(ieiN)
			a.ConfigurationUpdateIndication.Octet = ieiN
		case ConfigurationUpdateCommandGUTI5GType:
			a.GUTI5G = nasType.NewGUTI5G(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.GUTI5G.Len)
			a.GUTI5G.SetLen(a.GUTI5G.GetLen())
			binary.Read(buffer, binary.BigEndian, a.GUTI5G.Octet[:a.GUTI5G.GetLen()])
		case ConfigurationUpdateCommandTAIListType:
			a.TAIList = nasType.NewTAIList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.TAIList.Len)
			a.TAIList.SetLen(a.TAIList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.TAIList.Buffer[:a.TAIList.GetLen()])
		case ConfigurationUpdateCommandAllowedNSSAIType:
			a.AllowedNSSAI = nasType.NewAllowedNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AllowedNSSAI.Len)
			a.AllowedNSSAI.SetLen(a.AllowedNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AllowedNSSAI.Buffer[:a.AllowedNSSAI.GetLen()])
		case ConfigurationUpdateCommandServiceAreaListType:
			a.ServiceAreaList = nasType.NewServiceAreaList(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ServiceAreaList.Len)
			a.ServiceAreaList.SetLen(a.ServiceAreaList.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ServiceAreaList.Buffer[:a.ServiceAreaList.GetLen()])
		case ConfigurationUpdateCommandFullNameForNetworkType:
			a.FullNameForNetwork = nasType.NewFullNameForNetwork(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.FullNameForNetwork.Len)
			a.FullNameForNetwork.SetLen(a.FullNameForNetwork.GetLen())
			binary.Read(buffer, binary.BigEndian, a.FullNameForNetwork.Buffer[:a.FullNameForNetwork.GetLen()])
		case ConfigurationUpdateCommandShortNameForNetworkType:
			a.ShortNameForNetwork = nasType.NewShortNameForNetwork(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ShortNameForNetwork.Len)
			a.ShortNameForNetwork.SetLen(a.ShortNameForNetwork.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ShortNameForNetwork.Buffer[:a.ShortNameForNetwork.GetLen()])
		case ConfigurationUpdateCommandLocalTimeZoneType:
			a.LocalTimeZone = nasType.NewLocalTimeZone(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.LocalTimeZone.Octet)
		case ConfigurationUpdateCommandUniversalTimeAndLocalTimeZoneType:
			a.UniversalTimeAndLocalTimeZone = nasType.NewUniversalTimeAndLocalTimeZone(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.UniversalTimeAndLocalTimeZone.Octet)
		case ConfigurationUpdateCommandNetworkDaylightSavingTimeType:
			a.NetworkDaylightSavingTime = nasType.NewNetworkDaylightSavingTime(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.NetworkDaylightSavingTime.Len)
			a.NetworkDaylightSavingTime.SetLen(a.NetworkDaylightSavingTime.GetLen())
			binary.Read(buffer, binary.BigEndian, &a.NetworkDaylightSavingTime.Octet)
		case ConfigurationUpdateCommandLADNInformationType:
			a.LADNInformation = nasType.NewLADNInformation(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.LADNInformation.Len)
			a.LADNInformation.SetLen(a.LADNInformation.GetLen())
			binary.Read(buffer, binary.BigEndian, a.LADNInformation.Buffer[:a.LADNInformation.GetLen()])
		case ConfigurationUpdateCommandMICOIndicationType:
			a.MICOIndication = nasType.NewMICOIndication(ieiN)
			a.MICOIndication.Octet = ieiN
		case ConfigurationUpdateCommandNetworkSlicingIndicationType:
			a.NetworkSlicingIndication = nasType.NewNetworkSlicingIndication(ieiN)
			a.NetworkSlicingIndication.Octet = ieiN
		case ConfigurationUpdateCommandConfiguredNSSAIType:
			a.ConfiguredNSSAI = nasType.NewConfiguredNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ConfiguredNSSAI.Len)
			a.ConfiguredNSSAI.SetLen(a.ConfiguredNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ConfiguredNSSAI.Buffer[:a.ConfiguredNSSAI.GetLen()])
		case ConfigurationUpdateCommandRejectedNSSAIType:
			a.RejectedNSSAI = nasType.NewRejectedNSSAI(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RejectedNSSAI.Len)
			a.RejectedNSSAI.SetLen(a.RejectedNSSAI.GetLen())
			binary.Read(buffer, binary.BigEndian, a.RejectedNSSAI.Buffer[:a.RejectedNSSAI.GetLen()])
		case ConfigurationUpdateCommandOperatordefinedAccessCategoryDefinitionsType:
			a.OperatordefinedAccessCategoryDefinitions = nasType.NewOperatordefinedAccessCategoryDefinitions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.OperatordefinedAccessCategoryDefinitions.Len)
			a.OperatordefinedAccessCategoryDefinitions.SetLen(a.OperatordefinedAccessCategoryDefinitions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.OperatordefinedAccessCategoryDefinitions.Buffer[:a.OperatordefinedAccessCategoryDefinitions.GetLen()])
		case ConfigurationUpdateCommandSMSIndicationType:
			a.SMSIndication = nasType.NewSMSIndication(ieiN)
			a.SMSIndication.Octet = ieiN
		default:
		}
	}
}
