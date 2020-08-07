package nasMessage

import (
	"bytes"
	"encoding/binary"
	"my5G-RANTester/lib/nas/nasType"
)

type PDUSessionModificationCommand struct {
	nasType.ExtendedProtocolDiscriminator
	nasType.PDUSessionID
	nasType.PTI
	nasType.PDUSESSIONMODIFICATIONCOMMANDMessageIdentity
	*nasType.Cause5GSM
	*nasType.SessionAMBR
	*nasType.RQTimerValue
	*nasType.AlwaysonPDUSessionIndication
	*nasType.AuthorizedQosRules
	*nasType.MappedEPSBearerContexts
	*nasType.AuthorizedQosFlowDescriptions
	*nasType.ExtendedProtocolConfigurationOptions
}

func NewPDUSessionModificationCommand(iei uint8) (pDUSessionModificationCommand *PDUSessionModificationCommand) {
	pDUSessionModificationCommand = &PDUSessionModificationCommand{}
	return pDUSessionModificationCommand
}

const (
	PDUSessionModificationCommandCause5GSMType                            uint8 = 0x59
	PDUSessionModificationCommandSessionAMBRType                          uint8 = 0x2A
	PDUSessionModificationCommandRQTimerValueType                         uint8 = 0x56
	PDUSessionModificationCommandAlwaysonPDUSessionIndicationType         uint8 = 0x08
	PDUSessionModificationCommandAuthorizedQosRulesType                   uint8 = 0x7A
	PDUSessionModificationCommandMappedEPSBearerContextsType              uint8 = 0x7F
	PDUSessionModificationCommandAuthorizedQosFlowDescriptionsType        uint8 = 0x79
	PDUSessionModificationCommandExtendedProtocolConfigurationOptionsType uint8 = 0x7B
)

func (a *PDUSessionModificationCommand) EncodePDUSessionModificationCommand(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PDUSessionID.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PTI.Octet)
	binary.Write(buffer, binary.BigEndian, &a.PDUSESSIONMODIFICATIONCOMMANDMessageIdentity.Octet)
	if a.Cause5GSM != nil {
		binary.Write(buffer, binary.BigEndian, a.Cause5GSM.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.Cause5GSM.Octet)
	}
	if a.SessionAMBR != nil {
		binary.Write(buffer, binary.BigEndian, a.SessionAMBR.GetIei())
		binary.Write(buffer, binary.BigEndian, a.SessionAMBR.GetLen())
		binary.Write(buffer, binary.BigEndian, a.SessionAMBR.Octet[:a.SessionAMBR.GetLen()])
	}
	if a.RQTimerValue != nil {
		binary.Write(buffer, binary.BigEndian, a.RQTimerValue.GetIei())
		binary.Write(buffer, binary.BigEndian, &a.RQTimerValue.Octet)
	}
	if a.AlwaysonPDUSessionIndication != nil {
		binary.Write(buffer, binary.BigEndian, &a.AlwaysonPDUSessionIndication.Octet)
	}
	if a.AuthorizedQosRules != nil {
		binary.Write(buffer, binary.BigEndian, a.AuthorizedQosRules.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AuthorizedQosRules.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.AuthorizedQosRules.Buffer)
	}
	if a.MappedEPSBearerContexts != nil {
		binary.Write(buffer, binary.BigEndian, a.MappedEPSBearerContexts.GetIei())
		binary.Write(buffer, binary.BigEndian, a.MappedEPSBearerContexts.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.MappedEPSBearerContexts.Buffer)
	}
	if a.AuthorizedQosFlowDescriptions != nil {
		binary.Write(buffer, binary.BigEndian, a.AuthorizedQosFlowDescriptions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.AuthorizedQosFlowDescriptions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.AuthorizedQosFlowDescriptions.Buffer)
	}
	if a.ExtendedProtocolConfigurationOptions != nil {
		binary.Write(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.GetIei())
		binary.Write(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.GetLen())
		binary.Write(buffer, binary.BigEndian, &a.ExtendedProtocolConfigurationOptions.Buffer)
	}
}

func (a *PDUSessionModificationCommand) DecodePDUSessionModificationCommand(byteArray *[]byte) {
	buffer := bytes.NewBuffer(*byteArray)
	binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolDiscriminator.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PDUSessionID.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PTI.Octet)
	binary.Read(buffer, binary.BigEndian, &a.PDUSESSIONMODIFICATIONCOMMANDMessageIdentity.Octet)
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
		case PDUSessionModificationCommandCause5GSMType:
			a.Cause5GSM = nasType.NewCause5GSM(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.Cause5GSM.Octet)
		case PDUSessionModificationCommandSessionAMBRType:
			a.SessionAMBR = nasType.NewSessionAMBR(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.SessionAMBR.Len)
			a.SessionAMBR.SetLen(a.SessionAMBR.GetLen())
			binary.Read(buffer, binary.BigEndian, a.SessionAMBR.Octet[:a.SessionAMBR.GetLen()])
		case PDUSessionModificationCommandRQTimerValueType:
			a.RQTimerValue = nasType.NewRQTimerValue(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.RQTimerValue.Octet)
		case PDUSessionModificationCommandAlwaysonPDUSessionIndicationType:
			a.AlwaysonPDUSessionIndication = nasType.NewAlwaysonPDUSessionIndication(ieiN)
			a.AlwaysonPDUSessionIndication.Octet = ieiN
		case PDUSessionModificationCommandAuthorizedQosRulesType:
			a.AuthorizedQosRules = nasType.NewAuthorizedQosRules(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AuthorizedQosRules.Len)
			a.AuthorizedQosRules.SetLen(a.AuthorizedQosRules.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AuthorizedQosRules.Buffer[:a.AuthorizedQosRules.GetLen()])
		case PDUSessionModificationCommandMappedEPSBearerContextsType:
			a.MappedEPSBearerContexts = nasType.NewMappedEPSBearerContexts(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.MappedEPSBearerContexts.Len)
			a.MappedEPSBearerContexts.SetLen(a.MappedEPSBearerContexts.GetLen())
			binary.Read(buffer, binary.BigEndian, a.MappedEPSBearerContexts.Buffer[:a.MappedEPSBearerContexts.GetLen()])
		case PDUSessionModificationCommandAuthorizedQosFlowDescriptionsType:
			a.AuthorizedQosFlowDescriptions = nasType.NewAuthorizedQosFlowDescriptions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.AuthorizedQosFlowDescriptions.Len)
			a.AuthorizedQosFlowDescriptions.SetLen(a.AuthorizedQosFlowDescriptions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.AuthorizedQosFlowDescriptions.Buffer[:a.AuthorizedQosFlowDescriptions.GetLen()])
		case PDUSessionModificationCommandExtendedProtocolConfigurationOptionsType:
			a.ExtendedProtocolConfigurationOptions = nasType.NewExtendedProtocolConfigurationOptions(ieiN)
			binary.Read(buffer, binary.BigEndian, &a.ExtendedProtocolConfigurationOptions.Len)
			a.ExtendedProtocolConfigurationOptions.SetLen(a.ExtendedProtocolConfigurationOptions.GetLen())
			binary.Read(buffer, binary.BigEndian, a.ExtendedProtocolConfigurationOptions.Buffer[:a.ExtendedProtocolConfigurationOptions.GetLen()])
		default:
		}
	}
}
