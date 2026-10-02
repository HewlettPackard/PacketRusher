package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RegAccept{}

// RegAccept is detailed in 8.2.7 Registration accept, 24.501
type RegAccept struct {
	RegResult5GS                      *ie.RegResult5GS                      //    LV,       2B, 9.11.3.6
	GUTI5G                            *ie.MobileId5GS                       // TLV-E,      14B, 9.11.3.4
	EquivalentPlmns                   *ie.PLMNList                          //   TLV,    5-47B, 9.11.3.45
	TAIList                           *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
	AllowedNSSAI                      *ie.NSSAI                             //   TLV,    4-74B, 9.11.3.37
	RejectedNSSAI                     *ie.RejectedNSSAI                     //   TLV,    4-42B, 9.11.3.46
	ConfiguredNSSAI                   *ie.NSSAI                             //   TLV,   4-146B, 9.11.3.37
	NwFeatureSupport5GS               *ie.NwFeatureSupport5GS               //   TLV,     3-5B, 9.11.3.5
	PDUSessStatus                     *ie.PDUSessStatus                     //   TLV,    4-34B, 9.11.3.44
	PDUSessReactivationResult         *ie.PDUSessReactivationResult         //   TLV,    4-34B, 9.11.3.42
	PDUSessReactivationResultErrCause *ie.PDUSessReactivationResultErrCause // TLV-E,   5-515B, 9.11.3.43
	LADNInfo                          *ie.LADNInfo                          // TLV-E, 12-1715B, 9.11.3.30
	MICOInd                           *ie.MICOInd                           //    TV,       1B, 9.11.3.31
	NwSlicingInd                      *ie.NwSlicingInd                      //    TV,       1B, 9.11.3.36
	SvcAreaList                       *ie.SvcAreaList                       //   TLV,   6-114B, 9.11.3.49
	T3512Value                        *ie.GPRSTimer3                        //   TLV,       3B, 9.11.2.5
	Non3GppDeregTimerValue            *ie.GPRSTimer2                        //   TLV,       3B, 9.11.2.4
	T3502Value                        *ie.GPRSTimer2                        //   TLV,       3B, 9.11.2.4
	EmergNumList                      *ie.EmergNumList                      //   TLV,    5-50B, 9.11.3.23
	ExtendedEmergNumList              *ie.ExtendedEmergNumList              // TLV-E, 7-65538B, 9.11.3.26
	SORTransparentCntr                *ie.SORTransparentCntr                // TLV-E,    20-nB, 9.11.3.51
	EAPMsg                            *ie.EAPMsg                            // TLV-E,  7-1503B, 9.11.2.2
	NSSAIInclusionMode                *ie.NSSAIInclusionMode                //    TV,       1B, 9.11.3.37A
	OperatorDefinedAccessCategoryDefs *ie.OperatorDefinedAccessCategoryDefs // TLV-E,  3-8323B, 9.11.3.38
	NegotiatedDRXParams               *ie.DRXParams5GS                      //   TLV,       3B, 9.11.3.2A
	Non3GppNWPolicies                 *ie.Non3GppNWProvidedPolicies         //    TV,       1B, 9.11.3.36A
	EPSBearerCtxStatus                *ie.EPSBearerCtxStatus                //   TLV,       4B, 9.11.3.23A
	NegotiatedExtendedDRXParams       *ie.ExtendedDRXParams                 //   TLV,     3-4B, 9.11.3.26A
	T3447Value                        *ie.GPRSTimer3                        //   TLV,       3B, 9.11.2.5
	T3448Value                        *ie.GPRSTimer2                        //   TLV,       3B, 9.11.2.4
	T3324Value                        *ie.GPRSTimer3                        //   TLV,       3B, 9.11.2.5
	UERadioCapabilityID               *ie.UERadioCapabilityID               //   TLV,     3-nB, 9.11.3.68
	UERadioCapabilityIDDelInd         *ie.UERadioCapabilityIDDelInd         //    TV,       1B, 9.11.3.69
	PendingNSSAI                      *ie.NSSAI                             //   TLV,   4-146B, 9.11.3.37
	CipheringKeyData                  *ie.CipheringKeyData                  // TLV-E,    34-nB, 9.11.3.18C
	CAGInfoList                       *ie.CAGInfoList                       // TLV-E,     3-nB, 9.11.3.18A
	Truncated5GSTMSICfg               *ie.Truncated5GSTMSICfg               //   TLV,       3B, 9.11.3.70
	NegotiatedWUSAssistanceInfo       *ie.WUSAssistanceInfo                 //   TLV,     3-nB, 9.11.3.71
	NegotiatedNBN1ModeDRXParams       *ie.NBN1ModeDRXParams                 //   TLV,       3B, 9.11.3.73
	ExtendedRejectedNSSAI             *ie.ExtendedRejectedNSSAI             //   TLV,    5-90B, 9.11.3.75
	SvcLvlAACntr                      *ie.SvcLvlAACntr                      // TLV-E, 4-65538B, 9.11.2.10
	NegotiatedPEIPSAssistanceInfo     *ie.PEIPSAssistanceInfo               //   TLV,     3-nB, 9.11.3.80
	AdditionalReqResult5GS            *ie.AdditionalReqResult5GS            //   TLV,       3B, 9.11.3.81
	NSSRGInfo                         *ie.NSSRGInfo                         // TLV-E,  7-4099B, 9.11.3.82
	DisasterRoamingWaitRange          *ie.RegWaitRange                      //   TLV,       4B, 9.11.3.84
	DisasterReturnWaitRange           *ie.RegWaitRange                      //   TLV,       4B, 9.11.3.84
	DisasterPlmnList                  *ie.DisasterPlmnList                  //   TLV,     2-nB, 9.11.3.83
	ForbiddenTAI_5GSRoaming           *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
	ForbiddenTAI_5GSRegionalProvSvc   *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
	ExtendedCAGInfoList               *ie.ExtendedCAGInfoList               // TLV-E,     3-nB, 9.11.3.86
	NSAGInfo                          *ie.NSAGInfo                          // TLV-E,  9-3143B, 9.11.3.87
}

const (
	RegAcceptIEIGUTI5G                            uint8 = 0x77
	RegAcceptIEIEquivalentPlmns                   uint8 = 0x4A
	RegAcceptIEITAIList                           uint8 = 0x54
	RegAcceptIEIAllowedNSSAI                      uint8 = 0x15
	RegAcceptIEIRejectedNSSAI                     uint8 = 0x11
	RegAcceptIEIConfiguredNSSAI                   uint8 = 0x31
	RegAcceptIEINwFeatureSupport5GS               uint8 = 0x21
	RegAcceptIEIPDUSessStatus                     uint8 = 0x50
	RegAcceptIEIPDUSessReactivationResult         uint8 = 0x26
	RegAcceptIEIPDUSessReactivationResultErrCause uint8 = 0x72
	RegAcceptIEILADNInfo                          uint8 = 0x79
	RegAcceptIEIMICOInd                           uint8 = 0xB0
	RegAcceptIEINwSlicingInd                      uint8 = 0x90
	RegAcceptIEISvcAreaList                       uint8 = 0x27
	RegAcceptIEIT3512Value                        uint8 = 0x5E
	RegAcceptIEINon3GppDeregTimerValue            uint8 = 0x5D
	RegAcceptIEIT3502Value                        uint8 = 0x16
	RegAcceptIEIEmergNumList                      uint8 = 0x34
	RegAcceptIEIExtendedEmergNumList              uint8 = 0x7A
	RegAcceptIEISORTransparentCntr                uint8 = 0x73
	RegAcceptIEIEAPMsg                            uint8 = 0x78
	RegAcceptIEINSSAIInclusionMode                uint8 = 0xA0
	RegAcceptIEIOperatorDefinedAccessCategoryDefs uint8 = 0x76
	RegAcceptIEINegotiatedDRXParams               uint8 = 0x51
	RegAcceptIEINon3GppNWPolicies                 uint8 = 0xD0
	RegAcceptIEIEPSBearerCtxStatus                uint8 = 0x60
	RegAcceptIEINegotiatedExtendedDRXParams       uint8 = 0x6E
	RegAcceptIEIT3447Value                        uint8 = 0x6C
	RegAcceptIEIT3448Value                        uint8 = 0x6B
	RegAcceptIEIT3324Value                        uint8 = 0x6A
	RegAcceptIEIUERadioCapabilityID               uint8 = 0x67
	RegAcceptIEIUERadioCapabilityIDDelInd         uint8 = 0xE0
	RegAcceptIEIPendingNSSAI                      uint8 = 0x39
	RegAcceptIEICipheringKeyData                  uint8 = 0x74
	RegAcceptIEICAGInfoList                       uint8 = 0x75
	RegAcceptIEITruncated5GSTMSICfg               uint8 = 0x1B
	RegAcceptIEINegotiatedWUSAssistanceInfo       uint8 = 0x1C
	RegAcceptIEINegotiatedNBN1ModeDRXParams       uint8 = 0x29
	RegAcceptIEIExtendedRejectedNSSAI             uint8 = 0x68
	RegAcceptIEISvcLvlAACntr                      uint8 = 0x7B
	RegAcceptIEINegotiatedPEIPSAssistanceInfo     uint8 = 0x33
	RegAcceptIEIAdditionalReqResult5GS            uint8 = 0x35
	RegAcceptIEINSSRGInfo                         uint8 = 0x70
	RegAcceptIEIDisasterRoamingWaitRange          uint8 = 0x14
	RegAcceptIEIDisasterReturnWaitRange           uint8 = 0x2C
	RegAcceptIEIDisasterPlmnList                  uint8 = 0x13
	RegAcceptIEIForbiddenTAI_5GSRoaming           uint8 = 0x1D
	RegAcceptIEIForbiddenTAI_5GSRegionalProvSvc   uint8 = 0x1E
	RegAcceptIEIExtendedCAGInfoList               uint8 = 0x71
	RegAcceptIEINSAGInfo                          uint8 = 0x7C
)

func (m *RegAccept) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RegAccept) MsgType() MsgType {
	return MsgTypeRegAccept
}

func (m *RegAccept) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RegAccept len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.RegResult5GS = new(ie.RegResult5GS) // LV, 2B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of RegResult5GS")
	}
	if err = m.RegResult5GS.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "RegAccept.RegResult5GS.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case RegAcceptIEIGUTI5G: // TLV-E, 14B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of GUTI5G")
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
				return errors.Wrap(err, "RegAccept.GUTI5G.UnmarshalBinary")
			}
		case RegAcceptIEIEquivalentPlmns: // TLV, 5-47B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of EquivalentPlmns")
			}
			if m.EquivalentPlmns != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EquivalentPlmns = new(ie.PLMNList)
			if err = m.EquivalentPlmns.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EquivalentPlmns = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.EquivalentPlmns.UnmarshalBinary")
			}
		case RegAcceptIEITAIList: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of TAIList")
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
				return errors.Wrap(err, "RegAccept.TAIList.UnmarshalBinary")
			}
		case RegAcceptIEIAllowedNSSAI: // TLV, 4-74B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of AllowedNSSAI")
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
				return errors.Wrap(err, "RegAccept.AllowedNSSAI.UnmarshalBinary")
			}
		case RegAcceptIEIRejectedNSSAI: // TLV, 4-42B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of RejectedNSSAI")
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
				return errors.Wrap(err, "RegAccept.RejectedNSSAI.UnmarshalBinary")
			}
		case RegAcceptIEIConfiguredNSSAI: // TLV, 4-146B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ConfiguredNSSAI")
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
				return errors.Wrap(err, "RegAccept.ConfiguredNSSAI.UnmarshalBinary")
			}
		case RegAcceptIEINwFeatureSupport5GS: // TLV, 3-5B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NwFeatureSupport5GS")
			}
			if m.NwFeatureSupport5GS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NwFeatureSupport5GS = new(ie.NwFeatureSupport5GS)
			if err = m.NwFeatureSupport5GS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NwFeatureSupport5GS = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NwFeatureSupport5GS.UnmarshalBinary")
			}
		case RegAcceptIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of PDUSessStatus")
			}
			if m.PDUSessStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessStatus = new(ie.PDUSessStatus)
			if err = m.PDUSessStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessStatus = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.PDUSessStatus.UnmarshalBinary")
			}
		case RegAcceptIEIPDUSessReactivationResult: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of PDUSessReactivationResult")
			}
			if m.PDUSessReactivationResult != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessReactivationResult = new(ie.PDUSessReactivationResult)
			if err = m.PDUSessReactivationResult.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessReactivationResult = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.PDUSessReactivationResult.UnmarshalBinary")
			}
		case RegAcceptIEIPDUSessReactivationResultErrCause: // TLV-E, 5-515B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of PDUSessReactivationResultErrCause")
			}
			if m.PDUSessReactivationResultErrCause != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessReactivationResultErrCause = new(ie.PDUSessReactivationResultErrCause)
			if err = m.PDUSessReactivationResultErrCause.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessReactivationResultErrCause = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.PDUSessReactivationResultErrCause.UnmarshalBinary")
			}
		case RegAcceptIEILADNInfo: // TLV-E, 12-1715B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of LADNInfo")
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
				return errors.Wrap(err, "RegAccept.LADNInfo.UnmarshalBinary")
			}
		case RegAcceptIEIMICOInd: // TV, 1B
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
				return errors.Wrap(err, "RegAccept.MICOInd.UnmarshalBinary")
			}
		case RegAcceptIEINwSlicingInd: // TV, 1B
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
				return errors.Wrap(err, "RegAccept.NwSlicingInd.UnmarshalBinary")
			}
		case RegAcceptIEISvcAreaList: // TLV, 6-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of SvcAreaList")
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
				return errors.Wrap(err, "RegAccept.SvcAreaList.UnmarshalBinary")
			}
		case RegAcceptIEIT3512Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of T3512Value")
			}
			if m.T3512Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3512Value = new(ie.GPRSTimer3)
			if err = m.T3512Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3512Value = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.T3512Value.UnmarshalBinary")
			}
		case RegAcceptIEINon3GppDeregTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of Non3GppDeregTimerValue")
			}
			if m.Non3GppDeregTimerValue != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Non3GppDeregTimerValue = new(ie.GPRSTimer2)
			if err = m.Non3GppDeregTimerValue.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Non3GppDeregTimerValue = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.Non3GppDeregTimerValue.UnmarshalBinary")
			}
		case RegAcceptIEIT3502Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of T3502Value")
			}
			if m.T3502Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3502Value = new(ie.GPRSTimer2)
			if err = m.T3502Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3502Value = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.T3502Value.UnmarshalBinary")
			}
		case RegAcceptIEIEmergNumList: // TLV, 5-50B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of EmergNumList")
			}
			if m.EmergNumList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EmergNumList = new(ie.EmergNumList)
			if err = m.EmergNumList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EmergNumList = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.EmergNumList.UnmarshalBinary")
			}
		case RegAcceptIEIExtendedEmergNumList: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ExtendedEmergNumList")
			}
			if m.ExtendedEmergNumList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedEmergNumList = new(ie.ExtendedEmergNumList)
			if err = m.ExtendedEmergNumList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedEmergNumList = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.ExtendedEmergNumList.UnmarshalBinary")
			}
		case RegAcceptIEISORTransparentCntr: // TLV-E, 20-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of SORTransparentCntr")
			}
			if m.SORTransparentCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SORTransparentCntr = new(ie.SORTransparentCntr)
			if err = m.SORTransparentCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SORTransparentCntr = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.SORTransparentCntr.UnmarshalBinary")
			}
		case RegAcceptIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of EAPMsg")
			}
			if m.EAPMsg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EAPMsg = new(ie.EAPMsg)
			if err = m.EAPMsg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EAPMsg = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.EAPMsg.UnmarshalBinary")
			}
		case RegAcceptIEINSSAIInclusionMode: // TV, 1B
			if m.NSSAIInclusionMode != nil {
				break
			}
			m.NSSAIInclusionMode = new(ie.NSSAIInclusionMode)
			if err = m.NSSAIInclusionMode.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NSSAIInclusionMode = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NSSAIInclusionMode.UnmarshalBinary")
			}
		case RegAcceptIEIOperatorDefinedAccessCategoryDefs: // TLV-E, 3-8323B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of OperatorDefinedAccessCategoryDefs")
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
				return errors.Wrap(err, "RegAccept.OperatorDefinedAccessCategoryDefs.UnmarshalBinary")
			}
		case RegAcceptIEINegotiatedDRXParams: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NegotiatedDRXParams")
			}
			if m.NegotiatedDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NegotiatedDRXParams = new(ie.DRXParams5GS)
			if err = m.NegotiatedDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NegotiatedDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NegotiatedDRXParams.UnmarshalBinary")
			}
		case RegAcceptIEINon3GppNWPolicies: // TV, 1B
			if m.Non3GppNWPolicies != nil {
				break
			}
			m.Non3GppNWPolicies = new(ie.Non3GppNWProvidedPolicies)
			if err = m.Non3GppNWPolicies.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Non3GppNWPolicies = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.Non3GppNWPolicies.UnmarshalBinary")
			}
		case RegAcceptIEIEPSBearerCtxStatus: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of EPSBearerCtxStatus")
			}
			if m.EPSBearerCtxStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EPSBearerCtxStatus = new(ie.EPSBearerCtxStatus)
			if err = m.EPSBearerCtxStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EPSBearerCtxStatus = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.EPSBearerCtxStatus.UnmarshalBinary")
			}
		case RegAcceptIEINegotiatedExtendedDRXParams: // TLV, 3-4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NegotiatedExtendedDRXParams")
			}
			if m.NegotiatedExtendedDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NegotiatedExtendedDRXParams = new(ie.ExtendedDRXParams)
			if err = m.NegotiatedExtendedDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NegotiatedExtendedDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NegotiatedExtendedDRXParams.UnmarshalBinary")
			}
		case RegAcceptIEIT3447Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of T3447Value")
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
				return errors.Wrap(err, "RegAccept.T3447Value.UnmarshalBinary")
			}
		case RegAcceptIEIT3448Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of T3448Value")
			}
			if m.T3448Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3448Value = new(ie.GPRSTimer2)
			if err = m.T3448Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3448Value = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.T3448Value.UnmarshalBinary")
			}
		case RegAcceptIEIT3324Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of T3324Value")
			}
			if m.T3324Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3324Value = new(ie.GPRSTimer3)
			if err = m.T3324Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3324Value = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.T3324Value.UnmarshalBinary")
			}
		case RegAcceptIEIUERadioCapabilityID: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of UERadioCapabilityID")
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
				return errors.Wrap(err, "RegAccept.UERadioCapabilityID.UnmarshalBinary")
			}
		case RegAcceptIEIUERadioCapabilityIDDelInd: // TV, 1B
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
				return errors.Wrap(err, "RegAccept.UERadioCapabilityIDDelInd.UnmarshalBinary")
			}
		case RegAcceptIEIPendingNSSAI: // TLV, 4-146B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of PendingNSSAI")
			}
			if m.PendingNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PendingNSSAI = new(ie.NSSAI)
			if err = m.PendingNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PendingNSSAI = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.PendingNSSAI.UnmarshalBinary")
			}
		case RegAcceptIEICipheringKeyData: // TLV-E, 34-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of CipheringKeyData")
			}
			if m.CipheringKeyData != nil {
				reader.Next(int(ieLen))
				break
			}
			m.CipheringKeyData = new(ie.CipheringKeyData)
			if err = m.CipheringKeyData.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CipheringKeyData = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.CipheringKeyData.UnmarshalBinary")
			}
		case RegAcceptIEICAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of CAGInfoList")
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
				return errors.Wrap(err, "RegAccept.CAGInfoList.UnmarshalBinary")
			}
		case RegAcceptIEITruncated5GSTMSICfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of Truncated5GSTMSICfg")
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
				return errors.Wrap(err, "RegAccept.Truncated5GSTMSICfg.UnmarshalBinary")
			}
		case RegAcceptIEINegotiatedWUSAssistanceInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NegotiatedWUSAssistanceInfo")
			}
			if m.NegotiatedWUSAssistanceInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NegotiatedWUSAssistanceInfo = new(ie.WUSAssistanceInfo)
			if err = m.NegotiatedWUSAssistanceInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NegotiatedWUSAssistanceInfo = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NegotiatedWUSAssistanceInfo.UnmarshalBinary")
			}
		case RegAcceptIEINegotiatedNBN1ModeDRXParams: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NegotiatedNBN1ModeDRXParams")
			}
			if m.NegotiatedNBN1ModeDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NegotiatedNBN1ModeDRXParams = new(ie.NBN1ModeDRXParams)
			if err = m.NegotiatedNBN1ModeDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NegotiatedNBN1ModeDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NegotiatedNBN1ModeDRXParams.UnmarshalBinary")
			}
		case RegAcceptIEIExtendedRejectedNSSAI: // TLV, 5-90B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ExtendedRejectedNSSAI")
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
				return errors.Wrap(err, "RegAccept.ExtendedRejectedNSSAI.UnmarshalBinary")
			}
		case RegAcceptIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "RegAccept.SvcLvlAACntr.UnmarshalBinary")
			}
		case RegAcceptIEINegotiatedPEIPSAssistanceInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NegotiatedPEIPSAssistanceInfo")
			}
			if m.NegotiatedPEIPSAssistanceInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NegotiatedPEIPSAssistanceInfo = new(ie.PEIPSAssistanceInfo)
			if err = m.NegotiatedPEIPSAssistanceInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NegotiatedPEIPSAssistanceInfo = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.NegotiatedPEIPSAssistanceInfo.UnmarshalBinary")
			}
		case RegAcceptIEIAdditionalReqResult5GS: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of AdditionalReqResult5GS")
			}
			if m.AdditionalReqResult5GS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalReqResult5GS = new(ie.AdditionalReqResult5GS)
			if err = m.AdditionalReqResult5GS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalReqResult5GS = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.AdditionalReqResult5GS.UnmarshalBinary")
			}
		case RegAcceptIEINSSRGInfo: // TLV-E, 7-4099B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NSSRGInfo")
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
				return errors.Wrap(err, "RegAccept.NSSRGInfo.UnmarshalBinary")
			}
		case RegAcceptIEIDisasterRoamingWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of DisasterRoamingWaitRange")
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
				return errors.Wrap(err, "RegAccept.DisasterRoamingWaitRange.UnmarshalBinary")
			}
		case RegAcceptIEIDisasterReturnWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of DisasterReturnWaitRange")
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
				return errors.Wrap(err, "RegAccept.DisasterReturnWaitRange.UnmarshalBinary")
			}
		case RegAcceptIEIDisasterPlmnList: // TLV, 2-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of DisasterPlmnList")
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
				return errors.Wrap(err, "RegAccept.DisasterPlmnList.UnmarshalBinary")
			}
		case RegAcceptIEIForbiddenTAI_5GSRoaming: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ForbiddenTAI_5GSRoaming")
			}
			if m.ForbiddenTAI_5GSRoaming != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ForbiddenTAI_5GSRoaming = new(ie.TrackingAreaIdList5GS)
			if err = m.ForbiddenTAI_5GSRoaming.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ForbiddenTAI_5GSRoaming = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.ForbiddenTAI_5GSRoaming.UnmarshalBinary")
			}
		case RegAcceptIEIForbiddenTAI_5GSRegionalProvSvc: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ForbiddenTAI_5GSRegionalProvSvc")
			}
			if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ForbiddenTAI_5GSRegionalProvSvc = new(ie.TrackingAreaIdList5GS)
			if err = m.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ForbiddenTAI_5GSRegionalProvSvc = nil
					continue
				}
				return errors.Wrap(err, "RegAccept.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary")
			}
		case RegAcceptIEIExtendedCAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of ExtendedCAGInfoList")
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
				return errors.Wrap(err, "RegAccept.ExtendedCAGInfoList.UnmarshalBinary")
			}
		case RegAcceptIEINSAGInfo: // TLV-E, 9-3143B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegAccept UnmarshalBinary getIeLen of NSAGInfo")
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
				return errors.Wrap(err, "RegAccept.NSAGInfo.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RegAccept unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RegAccept) MarshalBinary() ([]byte, error) {
	if m.RegResult5GS == nil {
		return nil, errors.Errorf("RegResult5GS=%v must present in RegAccept",
			m.RegResult5GS)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRegAccept),
	})

	// regresult5gs, LV, 2B
	regresult5gs, err := m.RegResult5GS.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RegAccept.RegResult5GS.MarshalBinary()")
	}
	writer.WriteByte(byte(len(regresult5gs)))
	writer.Write(regresult5gs)

	// m.GUTI5G TLV-E, 14B, IEI=0x77
	if m.GUTI5G != nil {
		out, err := m.GUTI5G.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.GUTI5G.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIGUTI5G)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write GUTI5G")
		}
		writer.Write(out)
	}

	// m.EquivalentPlmns TLV, 5-47B, IEI=0x4A
	if m.EquivalentPlmns != nil {
		out, err := m.EquivalentPlmns.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.EquivalentPlmns.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIEquivalentPlmns)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.TAIList TLV, 9-114B, IEI=0x54
	if m.TAIList != nil {
		out, err := m.TAIList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.TAIList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEITAIList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AllowedNSSAI TLV, 4-74B, IEI=0x15
	if m.AllowedNSSAI != nil {
		out, err := m.AllowedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.AllowedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIAllowedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RejectedNSSAI TLV, 4-42B, IEI=0x11
	if m.RejectedNSSAI != nil {
		out, err := m.RejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.RejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ConfiguredNSSAI TLV, 4-146B, IEI=0x31
	if m.ConfiguredNSSAI != nil {
		out, err := m.ConfiguredNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ConfiguredNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIConfiguredNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NwFeatureSupport5GS TLV, 3-5B, IEI=0x21
	if m.NwFeatureSupport5GS != nil {
		out, err := m.NwFeatureSupport5GS.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NwFeatureSupport5GS.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINwFeatureSupport5GS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessReactivationResult TLV, 4-34B, IEI=0x26
	if m.PDUSessReactivationResult != nil {
		out, err := m.PDUSessReactivationResult.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.PDUSessReactivationResult.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIPDUSessReactivationResult)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessReactivationResultErrCause TLV-E, 5-515B, IEI=0x72
	if m.PDUSessReactivationResultErrCause != nil {
		out, err := m.PDUSessReactivationResultErrCause.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.PDUSessReactivationResultErrCause.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIPDUSessReactivationResultErrCause)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write PDUSessReactivationResultErrCause")
		}
		writer.Write(out)
	}

	// m.LADNInfo TLV-E, 12-1715B, IEI=0x79
	if m.LADNInfo != nil {
		out, err := m.LADNInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.LADNInfo.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEILADNInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write LADNInfo")
		}
		writer.Write(out)
	}

	// m.MICOInd TV, 1B, IEI=0xB0, >= 0x80 !
	if m.MICOInd != nil {
		out, err := m.MICOInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.MICOInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegAcceptIEIMICOInd)
	}

	// m.NwSlicingInd TV, 1B, IEI=0x90, >= 0x80 !
	if m.NwSlicingInd != nil {
		out, err := m.NwSlicingInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NwSlicingInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegAcceptIEINwSlicingInd)
	}

	// m.SvcAreaList TLV, 6-114B, IEI=0x27
	if m.SvcAreaList != nil {
		out, err := m.SvcAreaList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.SvcAreaList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEISvcAreaList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3512Value TLV, 3B, IEI=0x5E
	if m.T3512Value != nil {
		out, err := m.T3512Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.T3512Value.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIT3512Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.Non3GppDeregTimerValue TLV, 3B, IEI=0x5D
	if m.Non3GppDeregTimerValue != nil {
		out, err := m.Non3GppDeregTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.Non3GppDeregTimerValue.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINon3GppDeregTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3502Value TLV, 3B, IEI=0x16
	if m.T3502Value != nil {
		out, err := m.T3502Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.T3502Value.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIT3502Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EmergNumList TLV, 5-50B, IEI=0x34
	if m.EmergNumList != nil {
		out, err := m.EmergNumList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.EmergNumList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIEmergNumList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedEmergNumList TLV-E, 7-65538B, IEI=0x7A
	if m.ExtendedEmergNumList != nil {
		out, err := m.ExtendedEmergNumList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ExtendedEmergNumList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIExtendedEmergNumList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write ExtendedEmergNumList")
		}
		writer.Write(out)
	}

	// m.SORTransparentCntr TLV-E, 20-nB, IEI=0x73
	if m.SORTransparentCntr != nil {
		out, err := m.SORTransparentCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.SORTransparentCntr.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEISORTransparentCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write SORTransparentCntr")
		}
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.NSSAIInclusionMode TV, 1B, IEI=0xA0, >= 0x80 !
	if m.NSSAIInclusionMode != nil {
		out, err := m.NSSAIInclusionMode.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NSSAIInclusionMode.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegAcceptIEINSSAIInclusionMode)
	}

	// m.OperatorDefinedAccessCategoryDefs TLV-E, 3-8323B, IEI=0x76
	if m.OperatorDefinedAccessCategoryDefs != nil {
		out, err := m.OperatorDefinedAccessCategoryDefs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.OperatorDefinedAccessCategoryDefs.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIOperatorDefinedAccessCategoryDefs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write OperatorDefinedAccessCategoryDefs")
		}
		writer.Write(out)
	}

	// m.NegotiatedDRXParams TLV, 3B, IEI=0x51
	if m.NegotiatedDRXParams != nil {
		out, err := m.NegotiatedDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NegotiatedDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINegotiatedDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.Non3GppNWPolicies TV, 1B, IEI=0xD0, >= 0x80 !
	if m.Non3GppNWPolicies != nil {
		out, err := m.Non3GppNWPolicies.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.Non3GppNWPolicies.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegAcceptIEINon3GppNWPolicies)
	}

	// m.EPSBearerCtxStatus TLV, 4B, IEI=0x60
	if m.EPSBearerCtxStatus != nil {
		out, err := m.EPSBearerCtxStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.EPSBearerCtxStatus.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIEPSBearerCtxStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NegotiatedExtendedDRXParams TLV, 3-4B, IEI=0x6E
	if m.NegotiatedExtendedDRXParams != nil {
		out, err := m.NegotiatedExtendedDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NegotiatedExtendedDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINegotiatedExtendedDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3447Value TLV, 3B, IEI=0x6C
	if m.T3447Value != nil {
		out, err := m.T3447Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.T3447Value.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIT3447Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3448Value TLV, 3B, IEI=0x6B
	if m.T3448Value != nil {
		out, err := m.T3448Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.T3448Value.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIT3448Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3324Value TLV, 3B, IEI=0x6A
	if m.T3324Value != nil {
		out, err := m.T3324Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.T3324Value.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIT3324Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UERadioCapabilityID TLV, 3-nB, IEI=0x67
	if m.UERadioCapabilityID != nil {
		out, err := m.UERadioCapabilityID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.UERadioCapabilityID.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIUERadioCapabilityID)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UERadioCapabilityIDDelInd TV, 1B, IEI=0xE0, >= 0x80 !
	if m.UERadioCapabilityIDDelInd != nil {
		out, err := m.UERadioCapabilityIDDelInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.UERadioCapabilityIDDelInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegAcceptIEIUERadioCapabilityIDDelInd)
	}

	// m.PendingNSSAI TLV, 4-146B, IEI=0x39
	if m.PendingNSSAI != nil {
		out, err := m.PendingNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.PendingNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIPendingNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CipheringKeyData TLV-E, 34-nB, IEI=0x74
	if m.CipheringKeyData != nil {
		out, err := m.CipheringKeyData.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.CipheringKeyData.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEICipheringKeyData)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write CipheringKeyData")
		}
		writer.Write(out)
	}

	// m.CAGInfoList TLV-E, 3-nB, IEI=0x75
	if m.CAGInfoList != nil {
		out, err := m.CAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.CAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEICAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write CAGInfoList")
		}
		writer.Write(out)
	}

	// m.Truncated5GSTMSICfg TLV, 3B, IEI=0x1B
	if m.Truncated5GSTMSICfg != nil {
		out, err := m.Truncated5GSTMSICfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.Truncated5GSTMSICfg.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEITruncated5GSTMSICfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NegotiatedWUSAssistanceInfo TLV, 3-nB, IEI=0x1C
	if m.NegotiatedWUSAssistanceInfo != nil {
		out, err := m.NegotiatedWUSAssistanceInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NegotiatedWUSAssistanceInfo.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINegotiatedWUSAssistanceInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NegotiatedNBN1ModeDRXParams TLV, 3B, IEI=0x29
	if m.NegotiatedNBN1ModeDRXParams != nil {
		out, err := m.NegotiatedNBN1ModeDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NegotiatedNBN1ModeDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINegotiatedNBN1ModeDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedRejectedNSSAI TLV, 5-90B, IEI=0x68
	if m.ExtendedRejectedNSSAI != nil {
		out, err := m.ExtendedRejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ExtendedRejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIExtendedRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x7B
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}

	// m.NegotiatedPEIPSAssistanceInfo TLV, 3-nB, IEI=0x33
	if m.NegotiatedPEIPSAssistanceInfo != nil {
		out, err := m.NegotiatedPEIPSAssistanceInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NegotiatedPEIPSAssistanceInfo.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINegotiatedPEIPSAssistanceInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalReqResult5GS TLV, 3B, IEI=0x35
	if m.AdditionalReqResult5GS != nil {
		out, err := m.AdditionalReqResult5GS.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.AdditionalReqResult5GS.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIAdditionalReqResult5GS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NSSRGInfo TLV-E, 7-4099B, IEI=0x70
	if m.NSSRGInfo != nil {
		out, err := m.NSSRGInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NSSRGInfo.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINSSRGInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write NSSRGInfo")
		}
		writer.Write(out)
	}

	// m.DisasterRoamingWaitRange TLV, 4B, IEI=0x14
	if m.DisasterRoamingWaitRange != nil {
		out, err := m.DisasterRoamingWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.DisasterRoamingWaitRange.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIDisasterRoamingWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterReturnWaitRange TLV, 4B, IEI=0x2C
	if m.DisasterReturnWaitRange != nil {
		out, err := m.DisasterReturnWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.DisasterReturnWaitRange.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIDisasterReturnWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterPlmnList TLV, 2-nB, IEI=0x13
	if m.DisasterPlmnList != nil {
		out, err := m.DisasterPlmnList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.DisasterPlmnList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIDisasterPlmnList)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRoaming TLV, 9-114B, IEI=0x1D
	if m.ForbiddenTAI_5GSRoaming != nil {
		out, err := m.ForbiddenTAI_5GSRoaming.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ForbiddenTAI_5GSRoaming.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIForbiddenTAI_5GSRoaming)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRegionalProvSvc TLV, 9-114B, IEI=0x1E
	if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
		out, err := m.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIForbiddenTAI_5GSRegionalProvSvc)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedCAGInfoList TLV-E, 3-nB, IEI=0x71
	if m.ExtendedCAGInfoList != nil {
		out, err := m.ExtendedCAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.ExtendedCAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEIExtendedCAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write ExtendedCAGInfoList")
		}
		writer.Write(out)
	}

	// m.NSAGInfo TLV-E, 9-3143B, IEI=0x7C
	if m.NSAGInfo != nil {
		out, err := m.NSAGInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegAccept.NSAGInfo.MarshalBinary()")
		}
		writer.WriteByte(RegAcceptIEINSAGInfo)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegAccept) MarshalBinary() binary write NSAGInfo")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
