package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &CfgUpdateCmd{}

// CfgUpdateCmd is detailed in 8.2.19 Configuration update command, 24.501
type CfgUpdateCmd struct {
	CfgUpdateInd                      *ie.CfgUpdateInd                      //    TV,       1B, 9.11.3.18
	GUTI5G                            *ie.MobileId5GS                       // TLV-E,      14B, 9.11.3.4
	TAIList                           *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
	AllowedNSSAI                      *ie.NSSAI                             //   TLV,    4-74B, 9.11.3.37
	SvcAreaList                       *ie.SvcAreaList                       //   TLV,   6-114B, 9.11.3.49
	FullNameForNw                     *ie.NwName                            //   TLV,     3-nB, 9.11.3.35
	ShortNameForNw                    *ie.NwName                            //   TLV,     3-nB, 9.11.3.35
	LocalTimeZone                     *ie.TimeZone                          //    TV,       2B, 9.11.3.52
	UniversalTimeAndLocalTimeZone     *ie.TimeZoneAndTime                   //    TV,       8B, 9.11.3.53
	NwDST                             *ie.DST                               //   TLV,       3B, 9.11.3.19
	LADNInfo                          *ie.LADNInfo                          // TLV-E,  3-1715B, 9.11.3.30
	MICOInd                           *ie.MICOInd                           //    TV,       1B, 9.11.3.31
	NwSlicingInd                      *ie.NwSlicingInd                      //    TV,       1B, 9.11.3.36
	ConfiguredNSSAI                   *ie.NSSAI                             //   TLV,   4-146B, 9.11.3.37
	RejectedNSSAI                     *ie.RejectedNSSAI                     //   TLV,    4-42B, 9.11.3.46
	OperatorDefinedAccessCategoryDefs *ie.OperatorDefinedAccessCategoryDefs // TLV-E,  3-8323B, 9.11.3.38
	SMSInd                            *ie.SMSInd                            //    TV,       1B, 9.11.3.50A
	T3447Value                        *ie.GPRSTimer3                        //   TLV,       3B, 9.11.2.5
	CAGInfoList                       *ie.CAGInfoList                       // TLV-E,     3-nB, 9.11.3.18A
	UERadioCapabilityID               *ie.UERadioCapabilityID               //   TLV,     3-nB, 9.11.3.68
	UERadioCapabilityIDDelInd         *ie.UERadioCapabilityIDDelInd         //    TV,       1B, 9.11.3.69
	RegResult5GS                      *ie.RegResult5GS                      //   TLV,       3B, 9.11.3.6
	Truncated5GSTMSICfg               *ie.Truncated5GSTMSICfg               //   TLV,       3B, 9.11.3.70
	AdditionalCfgInd                  *ie.AdditionalCfgInd                  //    TV,       1B, 9.11.3.74
	ExtendedRejectedNSSAI             *ie.ExtendedRejectedNSSAI             //   TLV,    5-90B, 9.11.3.75
	SvcLvlAACntr                      *ie.SvcLvlAACntr                      // TLV-E, 4-65538B, 9.11.2.10
	NSSRGInfo                         *ie.NSSRGInfo                         // TLV-E,  7-4099B, 9.11.3.82
	DisasterRoamingWaitRange          *ie.RegWaitRange                      //   TLV,       4B, 9.11.3.84
	DisasterReturnWaitRange           *ie.RegWaitRange                      //   TLV,       4B, 9.11.3.84
	DisasterPlmnList                  *ie.DisasterPlmnList                  //   TLV,     2-nB, 9.11.3.83
	ExtendedCAGInfoList               *ie.ExtendedCAGInfoList               // TLV-E,     3-nB, 9.11.3.86
	UpdatedPEIPSAssistanceInfo        *ie.PEIPSAssistanceInfo               //   TLV,     3-nB, 9.11.3.80
	NSAGInfo                          *ie.NSAGInfo                          // TLV-E,  9-3143B, 9.11.3.87
	PriorityIndicator                 *ie.PriorityIndicator                 //    TV,       1B, 9.11.3.91
}

const (
	CfgUpdateCmdIEICfgUpdateInd                      uint8 = 0xD0
	CfgUpdateCmdIEIGUTI5G                            uint8 = 0x77
	CfgUpdateCmdIEITAIList                           uint8 = 0x54
	CfgUpdateCmdIEIAllowedNSSAI                      uint8 = 0x15
	CfgUpdateCmdIEISvcAreaList                       uint8 = 0x27
	CfgUpdateCmdIEIFullNameForNw                     uint8 = 0x43
	CfgUpdateCmdIEIShortNameForNw                    uint8 = 0x45
	CfgUpdateCmdIEILocalTimeZone                     uint8 = 0x46
	CfgUpdateCmdIEIUniversalTimeAndLocalTimeZone     uint8 = 0x47
	CfgUpdateCmdIEINwDST                             uint8 = 0x49
	CfgUpdateCmdIEILADNInfo                          uint8 = 0x79
	CfgUpdateCmdIEIMICOInd                           uint8 = 0xB0
	CfgUpdateCmdIEINwSlicingInd                      uint8 = 0x90
	CfgUpdateCmdIEIConfiguredNSSAI                   uint8 = 0x31
	CfgUpdateCmdIEIRejectedNSSAI                     uint8 = 0x11
	CfgUpdateCmdIEIOperatorDefinedAccessCategoryDefs uint8 = 0x76
	CfgUpdateCmdIEISMSInd                            uint8 = 0xF0
	CfgUpdateCmdIEIT3447Value                        uint8 = 0x6C
	CfgUpdateCmdIEICAGInfoList                       uint8 = 0x75
	CfgUpdateCmdIEIUERadioCapabilityID               uint8 = 0x67
	CfgUpdateCmdIEIUERadioCapabilityIDDelInd         uint8 = 0xA0
	CfgUpdateCmdIEIRegResult5GS                      uint8 = 0x44
	CfgUpdateCmdIEITruncated5GSTMSICfg               uint8 = 0x1B
	CfgUpdateCmdIEIAdditionalCfgInd                  uint8 = 0xC0
	CfgUpdateCmdIEIExtendedRejectedNSSAI             uint8 = 0x68
	CfgUpdateCmdIEISvcLvlAACntr                      uint8 = 0x72
	CfgUpdateCmdIEINSSRGInfo                         uint8 = 0x70
	CfgUpdateCmdIEIDisasterRoamingWaitRange          uint8 = 0x14
	CfgUpdateCmdIEIDisasterReturnWaitRange           uint8 = 0x2C
	CfgUpdateCmdIEIDisasterPlmnList                  uint8 = 0x13
	CfgUpdateCmdIEIExtendedCAGInfoList               uint8 = 0x71
	CfgUpdateCmdIEIUpdatedPEIPSAssistanceInfo        uint8 = 0x1F
	CfgUpdateCmdIEINSAGInfo                          uint8 = 0x73
	CfgUpdateCmdIEIPriorityIndicator                 uint8 = 0xE0
)

func (m *CfgUpdateCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *CfgUpdateCmd) MsgType() MsgType {
	return MsgTypeCfgUpdateCmd
}

func (m *CfgUpdateCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("CfgUpdateCmd len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// This message contains 0 Mandatory IE
	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case CfgUpdateCmdIEICfgUpdateInd: // TV, 1B
			if m.CfgUpdateInd != nil {
				break
			}
			m.CfgUpdateInd = new(ie.CfgUpdateInd)
			if err = m.CfgUpdateInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CfgUpdateInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.CfgUpdateInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIGUTI5G: // TLV-E, 14B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of GUTI5G")
			}
			if m.GUTI5G != nil {
				reader.Next(int(ieLen))
				break
			}
			m.GUTI5G = new(ie.MobileId5GS)
			if err = m.GUTI5G.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.GUTI5G = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.GUTI5G.UnmarshalBinary")
			}
		case CfgUpdateCmdIEITAIList: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of TAIList")
			}
			if m.TAIList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.TAIList = new(ie.TrackingAreaIdList5GS)
			if err = m.TAIList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.TAIList = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.TAIList.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIAllowedNSSAI: // TLV, 4-74B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of AllowedNSSAI")
			}
			if m.AllowedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AllowedNSSAI = new(ie.NSSAI)
			if err = m.AllowedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AllowedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.AllowedNSSAI.UnmarshalBinary")
			}
		case CfgUpdateCmdIEISvcAreaList: // TLV, 6-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of SvcAreaList")
			}
			if m.SvcAreaList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SvcAreaList = new(ie.SvcAreaList)
			if err = m.SvcAreaList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SvcAreaList = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.SvcAreaList.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIFullNameForNw: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of FullNameForNw")
			}
			if m.FullNameForNw != nil {
				reader.Next(int(ieLen))
				break
			}
			m.FullNameForNw = new(ie.NwName)
			if err = m.FullNameForNw.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.FullNameForNw = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.FullNameForNw.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIShortNameForNw: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of ShortNameForNw")
			}
			if m.ShortNameForNw != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ShortNameForNw = new(ie.NwName)
			if err = m.ShortNameForNw.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ShortNameForNw = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.ShortNameForNw.UnmarshalBinary")
			}
		case CfgUpdateCmdIEILocalTimeZone: // TV, 2B
			ieLen = 1
			if m.LocalTimeZone != nil {
				reader.Next(int(ieLen))
				break
			}
			m.LocalTimeZone = new(ie.TimeZone)
			if err = m.LocalTimeZone.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.LocalTimeZone = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.LocalTimeZone.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIUniversalTimeAndLocalTimeZone: // TV, 8B
			ieLen = 7
			if m.UniversalTimeAndLocalTimeZone != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UniversalTimeAndLocalTimeZone = new(ie.TimeZoneAndTime)
			if err = m.UniversalTimeAndLocalTimeZone.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UniversalTimeAndLocalTimeZone = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.UniversalTimeAndLocalTimeZone.UnmarshalBinary")
			}
		case CfgUpdateCmdIEINwDST: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of NwDST")
			}
			if m.NwDST != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NwDST = new(ie.DST)
			if err = m.NwDST.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NwDST = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.NwDST.UnmarshalBinary")
			}
		case CfgUpdateCmdIEILADNInfo: // TLV-E, 3-1715B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of LADNInfo")
			}
			if m.LADNInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.LADNInfo = new(ie.LADNInfo)
			if err = m.LADNInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.LADNInfo = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.LADNInfo.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIMICOInd: // TV, 1B
			if m.MICOInd != nil {
				break
			}
			m.MICOInd = new(ie.MICOInd)
			if err = m.MICOInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.MICOInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.MICOInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEINwSlicingInd: // TV, 1B
			if m.NwSlicingInd != nil {
				break
			}
			m.NwSlicingInd = new(ie.NwSlicingInd)
			if err = m.NwSlicingInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NwSlicingInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.NwSlicingInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIConfiguredNSSAI: // TLV, 4-146B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of ConfiguredNSSAI")
			}
			if m.ConfiguredNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ConfiguredNSSAI = new(ie.NSSAI)
			if err = m.ConfiguredNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ConfiguredNSSAI = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.ConfiguredNSSAI.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIRejectedNSSAI: // TLV, 4-42B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of RejectedNSSAI")
			}
			if m.RejectedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RejectedNSSAI = new(ie.RejectedNSSAI)
			if err = m.RejectedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RejectedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.RejectedNSSAI.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIOperatorDefinedAccessCategoryDefs: // TLV-E, 3-8323B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of OperatorDefinedAccessCategoryDefs")
			}
			if m.OperatorDefinedAccessCategoryDefs != nil {
				reader.Next(int(ieLen))
				break
			}
			m.OperatorDefinedAccessCategoryDefs = new(ie.OperatorDefinedAccessCategoryDefs)
			if err = m.OperatorDefinedAccessCategoryDefs.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.OperatorDefinedAccessCategoryDefs = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.OperatorDefinedAccessCategoryDefs.UnmarshalBinary")
			}
		case CfgUpdateCmdIEISMSInd: // TV, 1B
			if m.SMSInd != nil {
				break
			}
			m.SMSInd = new(ie.SMSInd)
			if err = m.SMSInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SMSInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.SMSInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIT3447Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of T3447Value")
			}
			if m.T3447Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3447Value = new(ie.GPRSTimer3)
			if err = m.T3447Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3447Value = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.T3447Value.UnmarshalBinary")
			}
		case CfgUpdateCmdIEICAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of CAGInfoList")
			}
			if m.CAGInfoList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.CAGInfoList = new(ie.CAGInfoList)
			if err = m.CAGInfoList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CAGInfoList = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.CAGInfoList.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIUERadioCapabilityID: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of UERadioCapabilityID")
			}
			if m.UERadioCapabilityID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UERadioCapabilityID = new(ie.UERadioCapabilityID)
			if err = m.UERadioCapabilityID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UERadioCapabilityID = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.UERadioCapabilityID.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIUERadioCapabilityIDDelInd: // TV, 1B
			if m.UERadioCapabilityIDDelInd != nil {
				break
			}
			m.UERadioCapabilityIDDelInd = new(ie.UERadioCapabilityIDDelInd)
			if err = m.UERadioCapabilityIDDelInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UERadioCapabilityIDDelInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.UERadioCapabilityIDDelInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIRegResult5GS: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of RegResult5GS")
			}
			if m.RegResult5GS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RegResult5GS = new(ie.RegResult5GS)
			if err = m.RegResult5GS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RegResult5GS = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.RegResult5GS.UnmarshalBinary")
			}
		case CfgUpdateCmdIEITruncated5GSTMSICfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of Truncated5GSTMSICfg")
			}
			if m.Truncated5GSTMSICfg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Truncated5GSTMSICfg = new(ie.Truncated5GSTMSICfg)
			if err = m.Truncated5GSTMSICfg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Truncated5GSTMSICfg = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.Truncated5GSTMSICfg.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIAdditionalCfgInd: // TV, 1B
			if m.AdditionalCfgInd != nil {
				break
			}
			m.AdditionalCfgInd = new(ie.AdditionalCfgInd)
			if err = m.AdditionalCfgInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalCfgInd = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.AdditionalCfgInd.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIExtendedRejectedNSSAI: // TLV, 5-90B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of ExtendedRejectedNSSAI")
			}
			if m.ExtendedRejectedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedRejectedNSSAI = new(ie.ExtendedRejectedNSSAI)
			if err = m.ExtendedRejectedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedRejectedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.ExtendedRejectedNSSAI.UnmarshalBinary")
			}
		case CfgUpdateCmdIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of SvcLvlAACntr")
			}
			if m.SvcLvlAACntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SvcLvlAACntr = new(ie.SvcLvlAACntr)
			if err = m.SvcLvlAACntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SvcLvlAACntr = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.SvcLvlAACntr.UnmarshalBinary")
			}
		case CfgUpdateCmdIEINSSRGInfo: // TLV-E, 7-4099B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of NSSRGInfo")
			}
			if m.NSSRGInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NSSRGInfo = new(ie.NSSRGInfo)
			if err = m.NSSRGInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NSSRGInfo = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.NSSRGInfo.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIDisasterRoamingWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of DisasterRoamingWaitRange")
			}
			if m.DisasterRoamingWaitRange != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DisasterRoamingWaitRange = new(ie.RegWaitRange)
			if err = m.DisasterRoamingWaitRange.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DisasterRoamingWaitRange = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.DisasterRoamingWaitRange.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIDisasterReturnWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of DisasterReturnWaitRange")
			}
			if m.DisasterReturnWaitRange != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DisasterReturnWaitRange = new(ie.RegWaitRange)
			if err = m.DisasterReturnWaitRange.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DisasterReturnWaitRange = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.DisasterReturnWaitRange.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIDisasterPlmnList: // TLV, 2-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of DisasterPlmnList")
			}
			if m.DisasterPlmnList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DisasterPlmnList = new(ie.DisasterPlmnList)
			if err = m.DisasterPlmnList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DisasterPlmnList = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.DisasterPlmnList.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIExtendedCAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of ExtendedCAGInfoList")
			}
			if m.ExtendedCAGInfoList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedCAGInfoList = new(ie.ExtendedCAGInfoList)
			if err = m.ExtendedCAGInfoList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedCAGInfoList = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.ExtendedCAGInfoList.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIUpdatedPEIPSAssistanceInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of UpdatedPEIPSAssistanceInfo")
			}
			if m.UpdatedPEIPSAssistanceInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UpdatedPEIPSAssistanceInfo = new(ie.PEIPSAssistanceInfo)
			if err = m.UpdatedPEIPSAssistanceInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UpdatedPEIPSAssistanceInfo = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.UpdatedPEIPSAssistanceInfo.UnmarshalBinary")
			}
		case CfgUpdateCmdIEINSAGInfo: // TLV-E, 9-3143B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CfgUpdateCmd UnmarshalBinary getIeLen of NSAGInfo")
			}
			if m.NSAGInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NSAGInfo = new(ie.NSAGInfo)
			if err = m.NSAGInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NSAGInfo = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.NSAGInfo.UnmarshalBinary")
			}
		case CfgUpdateCmdIEIPriorityIndicator: // TV, 1B
			if m.PriorityIndicator != nil {
				break
			}
			m.PriorityIndicator = new(ie.PriorityIndicator)
			if err = m.PriorityIndicator.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PriorityIndicator = nil
					continue
				}
				return errors.Wrap(err, "CfgUpdateCmd.PriorityIndicator.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("CfgUpdateCmd unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *CfgUpdateCmd) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeCfgUpdateCmd),
	})

	// m.CfgUpdateInd TV, 1B, IEI=0xD0, >= 0x80 !
	if m.CfgUpdateInd != nil {
		out, err := m.CfgUpdateInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.CfgUpdateInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEICfgUpdateInd)
	}

	// m.GUTI5G TLV-E, 14B, IEI=0x77
	if m.GUTI5G != nil {
		out, err := m.GUTI5G.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.GUTI5G.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIGUTI5G)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write GUTI5G")
		}
		writer.Write(out)
	}

	// m.TAIList TLV, 9-114B, IEI=0x54
	if m.TAIList != nil {
		out, err := m.TAIList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.TAIList.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEITAIList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AllowedNSSAI TLV, 4-74B, IEI=0x15
	if m.AllowedNSSAI != nil {
		out, err := m.AllowedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.AllowedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIAllowedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcAreaList TLV, 6-114B, IEI=0x27
	if m.SvcAreaList != nil {
		out, err := m.SvcAreaList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.SvcAreaList.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEISvcAreaList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.FullNameForNw TLV, 3-nB, IEI=0x43
	if m.FullNameForNw != nil {
		out, err := m.FullNameForNw.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.FullNameForNw.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIFullNameForNw)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ShortNameForNw TLV, 3-nB, IEI=0x45
	if m.ShortNameForNw != nil {
		out, err := m.ShortNameForNw.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.ShortNameForNw.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIShortNameForNw)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.LocalTimeZone TV, 2B, IEI=0x46
	if m.LocalTimeZone != nil {
		out, err := m.LocalTimeZone.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.LocalTimeZone.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEILocalTimeZone)
		writer.Write(out)
	}

	// m.UniversalTimeAndLocalTimeZone TV, 8B, IEI=0x47
	if m.UniversalTimeAndLocalTimeZone != nil {
		out, err := m.UniversalTimeAndLocalTimeZone.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.UniversalTimeAndLocalTimeZone.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIUniversalTimeAndLocalTimeZone)
		writer.Write(out)
	}

	// m.NwDST TLV, 3B, IEI=0x49
	if m.NwDST != nil {
		out, err := m.NwDST.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.NwDST.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEINwDST)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.LADNInfo TLV-E, 3-1715B, IEI=0x79
	if m.LADNInfo != nil {
		out, err := m.LADNInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.LADNInfo.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEILADNInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write LADNInfo")
		}
		writer.Write(out)
	}

	// m.MICOInd TV, 1B, IEI=0xB0, >= 0x80 !
	if m.MICOInd != nil {
		out, err := m.MICOInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.MICOInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEIMICOInd)
	}

	// m.NwSlicingInd TV, 1B, IEI=0x90, >= 0x80 !
	if m.NwSlicingInd != nil {
		out, err := m.NwSlicingInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.NwSlicingInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEINwSlicingInd)
	}

	// m.ConfiguredNSSAI TLV, 4-146B, IEI=0x31
	if m.ConfiguredNSSAI != nil {
		out, err := m.ConfiguredNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.ConfiguredNSSAI.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIConfiguredNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RejectedNSSAI TLV, 4-42B, IEI=0x11
	if m.RejectedNSSAI != nil {
		out, err := m.RejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.RejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.OperatorDefinedAccessCategoryDefs TLV-E, 3-8323B, IEI=0x76
	if m.OperatorDefinedAccessCategoryDefs != nil {
		out, err := m.OperatorDefinedAccessCategoryDefs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.OperatorDefinedAccessCategoryDefs.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIOperatorDefinedAccessCategoryDefs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write OperatorDefinedAccessCategoryDefs")
		}
		writer.Write(out)
	}

	// m.SMSInd TV, 1B, IEI=0xF0, >= 0x80 !
	if m.SMSInd != nil {
		out, err := m.SMSInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.SMSInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEISMSInd)
	}

	// m.T3447Value TLV, 3B, IEI=0x6C
	if m.T3447Value != nil {
		out, err := m.T3447Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.T3447Value.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIT3447Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CAGInfoList TLV-E, 3-nB, IEI=0x75
	if m.CAGInfoList != nil {
		out, err := m.CAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.CAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEICAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write CAGInfoList")
		}
		writer.Write(out)
	}

	// m.UERadioCapabilityID TLV, 3-nB, IEI=0x67
	if m.UERadioCapabilityID != nil {
		out, err := m.UERadioCapabilityID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.UERadioCapabilityID.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIUERadioCapabilityID)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UERadioCapabilityIDDelInd TV, 1B, IEI=0xA0, >= 0x80 !
	if m.UERadioCapabilityIDDelInd != nil {
		out, err := m.UERadioCapabilityIDDelInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.UERadioCapabilityIDDelInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEIUERadioCapabilityIDDelInd)
	}

	// m.RegResult5GS TLV, 3B, IEI=0x44
	if m.RegResult5GS != nil {
		out, err := m.RegResult5GS.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.RegResult5GS.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIRegResult5GS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.Truncated5GSTMSICfg TLV, 3B, IEI=0x1B
	if m.Truncated5GSTMSICfg != nil {
		out, err := m.Truncated5GSTMSICfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.Truncated5GSTMSICfg.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEITruncated5GSTMSICfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalCfgInd TV, 1B, IEI=0xC0, >= 0x80 !
	if m.AdditionalCfgInd != nil {
		out, err := m.AdditionalCfgInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.AdditionalCfgInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEIAdditionalCfgInd)
	}

	// m.ExtendedRejectedNSSAI TLV, 5-90B, IEI=0x68
	if m.ExtendedRejectedNSSAI != nil {
		out, err := m.ExtendedRejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.ExtendedRejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIExtendedRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}

	// m.NSSRGInfo TLV-E, 7-4099B, IEI=0x70
	if m.NSSRGInfo != nil {
		out, err := m.NSSRGInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.NSSRGInfo.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEINSSRGInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write NSSRGInfo")
		}
		writer.Write(out)
	}

	// m.DisasterRoamingWaitRange TLV, 4B, IEI=0x14
	if m.DisasterRoamingWaitRange != nil {
		out, err := m.DisasterRoamingWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.DisasterRoamingWaitRange.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIDisasterRoamingWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterReturnWaitRange TLV, 4B, IEI=0x2C
	if m.DisasterReturnWaitRange != nil {
		out, err := m.DisasterReturnWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.DisasterReturnWaitRange.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIDisasterReturnWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterPlmnList TLV, 2-nB, IEI=0x13
	if m.DisasterPlmnList != nil {
		out, err := m.DisasterPlmnList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.DisasterPlmnList.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIDisasterPlmnList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedCAGInfoList TLV-E, 3-nB, IEI=0x71
	if m.ExtendedCAGInfoList != nil {
		out, err := m.ExtendedCAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.ExtendedCAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIExtendedCAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write ExtendedCAGInfoList")
		}
		writer.Write(out)
	}

	// m.UpdatedPEIPSAssistanceInfo TLV, 3-nB, IEI=0x1F
	if m.UpdatedPEIPSAssistanceInfo != nil {
		out, err := m.UpdatedPEIPSAssistanceInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.UpdatedPEIPSAssistanceInfo.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEIUpdatedPEIPSAssistanceInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NSAGInfo TLV-E, 9-3143B, IEI=0x73
	if m.NSAGInfo != nil {
		out, err := m.NSAGInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.NSAGInfo.MarshalBinary()")
		}
		writer.WriteByte(CfgUpdateCmdIEINSAGInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd) MarshalBinary() binary write NSAGInfo")
		}
		writer.Write(out)
	}

	// m.PriorityIndicator TV, 1B, IEI=0xE0, >= 0x80 !
	if m.PriorityIndicator != nil {
		out, err := m.PriorityIndicator.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CfgUpdateCmd.PriorityIndicator.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CfgUpdateCmdIEIPriorityIndicator)
	}
	return writer.Bytes(), nil
}
