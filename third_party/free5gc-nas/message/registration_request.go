package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RegReq{}

// RegReq is detailed in 8.2.6 Registration request, 24.501
type RegReq struct {
	RegType5GS                  *ie.RegType5GS              //     V,     1/2B, 9.11.3.7
	Ngksi                       *ie.NASKeySetId             //     V,     1/2B, 9.11.3.32
	MobileId5GS                 *ie.MobileId5GS             //  LV-E,     6-nB, 9.11.3.4
	NoncurrentNativeNASKeySetId *ie.NASKeySetId             //    TV,       1B, 9.11.3.32
	Capability5GMM              *ie.Capability5GMM          //   TLV,    3-15B, 9.11.3.1
	UESecCapability             *ie.UESecCapability         //   TLV,    4-10B, 9.11.3.54
	ReqNSSAI                    *ie.NSSAI                   //   TLV,    4-74B, 9.11.3.37
	LastVisitedRegisteredTAI    *ie.TrackingAreaId5GS       //    TV,       7B, 9.11.3.8
	S1UENwCapability            *ie.S1UENwCapability        //   TLV,    4-15B, 9.11.3.48
	UplinkDataStatus            *ie.UplinkDataStatus        //   TLV,    4-34B, 9.11.3.57
	PDUSessStatus               *ie.PDUSessStatus           //   TLV,    4-34B, 9.11.3.44
	MICOInd                     *ie.MICOInd                 //    TV,       1B, 9.11.3.31
	UEStatus                    *ie.UEStatus                //   TLV,       3B, 9.11.3.56
	AdditionalGUTI              *ie.MobileId5GS             // TLV-E,      14B, 9.11.3.4
	AllowedPDUSessStatus        *ie.AllowedPDUSessStatus    //   TLV,    4-34B, 9.11.3.13
	UesUsageSetting             *ie.UesUsageSetting         //   TLV,       3B, 9.11.3.55
	ReqDRXParams                *ie.DRXParams5GS            //   TLV,       3B, 9.11.3.2A
	EPSNASMsgCntr               *ie.EPSNASMsgCntr           // TLV-E,     4-nB, 9.11.3.24
	LADNInd                     *ie.LADNInd                 // TLV-E,   3-811B, 9.11.3.29
	PayloadCntrType             *ie.PayloadCntrType         //    TV,       1B, 9.11.3.40
	PayloadCntr                 *ie.PayloadCntr             // TLV-E, 4-65538B, 9.11.3.39
	NwSlicingInd                *ie.NwSlicingInd            //    TV,       1B, 9.11.3.36
	UpdateType5GS               *ie.UpdateType5GS           //   TLV,       3B, 9.11.3.9A
	MobileStationClassmark2     *ie.MobileStationClassmark2 //   TLV,       5B, 9.11.3.31C
	SupportedCodecs             *ie.SupportedCodecList      //   TLV,     5-nB, 9.11.3.51A
	NASMsgCntr                  *ie.NASMsgCntr              // TLV-E,     4-nB, 9.11.3.33
	EPSBearerCtxStatus          *ie.EPSBearerCtxStatus      //   TLV,       4B, 9.11.3.23A
	ReqExtendedDRXParams        *ie.ExtendedDRXParams       //   TLV,     3-4B, 9.11.3.26A
	T3324Value                  *ie.GPRSTimer3              //   TLV,       3B, 9.11.2.5
	UERadioCapabilityID         *ie.UERadioCapabilityID     //   TLV,     3-nB, 9.11.3.68
	ReqMappedNSSAI              *ie.MappedNSSAI             //   TLV,    3-42B, 9.11.3.31B
	AdditionalInfoReq           *ie.AdditionalInfoReq       //   TLV,       3B, 9.11.3.12A
	ReqWUSAssistanceInfo        *ie.WUSAssistanceInfo       //   TLV,     3-nB, 9.11.3.71
	N5GCInd                     *ie.N5GCInd                 //    TV,       1B, 9.11.3.72
	ReqNBN1ModeDRXParams        *ie.NBN1ModeDRXParams       //   TLV,       3B, 9.11.3.73
	UEReqType                   *ie.UEReqType               //   TLV,       3B, 9.11.3.76
	PagingRestriction           *ie.PagingRestriction       //   TLV,    3-35B, 9.11.3.77
	SvcLvlAACntr                *ie.SvcLvlAACntr            // TLV-E, 4-65538B, 9.11.2.10
	NID                         *ie.NID                     //   TLV,       8B, 9.11.3.79
	DisasterPlmnMS              *ie.PlmnId                  //   TLV,       5B, 9.11.3.85
	ReqPEIPSAssistanceInfo      *ie.PEIPSAssistanceInfo     //   TLV,     3-nB, 9.11.3.80
	ReqT3512Value               *ie.GPRSTimer3              //   TLV,       3B, 9.11.2.5
}

const (
	RegReqIEINoncurrentNativeNASKeySetId uint8 = 0xC0
	RegReqIEICapability5GMM              uint8 = 0x10
	RegReqIEIUESecCapability             uint8 = 0x2E
	RegReqIEIReqNSSAI                    uint8 = 0x2F
	RegReqIEILastVisitedRegisteredTAI    uint8 = 0x52
	RegReqIEIS1UENwCapability            uint8 = 0x17
	RegReqIEIUplinkDataStatus            uint8 = 0x40
	RegReqIEIPDUSessStatus               uint8 = 0x50
	RegReqIEIMICOInd                     uint8 = 0xB0
	RegReqIEIUEStatus                    uint8 = 0x2B
	RegReqIEIAdditionalGUTI              uint8 = 0x77
	RegReqIEIAllowedPDUSessStatus        uint8 = 0x25
	RegReqIEIUesUsageSetting             uint8 = 0x18
	RegReqIEIReqDRXParams                uint8 = 0x51
	RegReqIEIEPSNASMsgCntr               uint8 = 0x70
	RegReqIEILADNInd                     uint8 = 0x74
	RegReqIEIPayloadCntrType             uint8 = 0x80
	RegReqIEIPayloadCntr                 uint8 = 0x7B
	RegReqIEINwSlicingInd                uint8 = 0x90
	RegReqIEIUpdateType5GS               uint8 = 0x53
	RegReqIEIMobileStationClassmark2     uint8 = 0x41
	RegReqIEISupportedCodecs             uint8 = 0x42
	RegReqIEINASMsgCntr                  uint8 = 0x71
	RegReqIEIEPSBearerCtxStatus          uint8 = 0x60
	RegReqIEIReqExtendedDRXParams        uint8 = 0x6E
	RegReqIEIT3324Value                  uint8 = 0x6A
	RegReqIEIUERadioCapabilityID         uint8 = 0x67
	RegReqIEIReqMappedNSSAI              uint8 = 0x35
	RegReqIEIAdditionalInfoReq           uint8 = 0x48
	RegReqIEIReqWUSAssistanceInfo        uint8 = 0x1A
	RegReqIEIN5GCInd                     uint8 = 0xA0
	RegReqIEIReqNBN1ModeDRXParams        uint8 = 0x30
	RegReqIEIUEReqType                   uint8 = 0x29
	RegReqIEIPagingRestriction           uint8 = 0x28
	RegReqIEISvcLvlAACntr                uint8 = 0x72
	RegReqIEINID                         uint8 = 0x32
	RegReqIEIdisasterPlmnMS              uint8 = 0x16
	RegReqIEIReqPEIPSAssistanceInfo      uint8 = 0x2A
	RegReqIEIReqT3512Value               uint8 = 0x3B
)

func (m *RegReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RegReq) MsgType() MsgType {
	return MsgTypeRegReq
}

func (m *RegReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RegReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, half85 := ie.GetHalfIEValue(reader.Next(1))
	m.RegType5GS = new(ie.RegType5GS) // V, 1/2B
	if err = m.RegType5GS.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "RegReq.RegType5GS.UnmarshalBinary")
	}
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half85); err != nil {
		return errors.Wrap(err, "RegReq.Ngksi.UnmarshalBinary")
	}

	m.MobileId5GS = new(ie.MobileId5GS) // LV-E, 6-nB
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of MobileId5GS")
	}
	if err = m.MobileId5GS.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "RegReq.MobileId5GS.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case RegReqIEINoncurrentNativeNASKeySetId: // TV, 1B
			if m.NoncurrentNativeNASKeySetId != nil {
				break
			}
			m.NoncurrentNativeNASKeySetId = new(ie.NASKeySetId)
			if err = m.NoncurrentNativeNASKeySetId.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NoncurrentNativeNASKeySetId = nil
					continue
				}
				return errors.Wrap(err, "RegReq.NoncurrentNativeNASKeySetId.UnmarshalBinary")
			}
		case RegReqIEICapability5GMM: // TLV, 3-15B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of Capability5GMM")
			}
			if m.Capability5GMM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Capability5GMM = new(ie.Capability5GMM)
			if err = m.Capability5GMM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Capability5GMM = nil
					continue
				}
				return errors.Wrap(err, "RegReq.Capability5GMM.UnmarshalBinary")
			}
		case RegReqIEIUESecCapability: // TLV, 4-10B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UESecCapability")
			}
			if m.UESecCapability != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UESecCapability = new(ie.UESecCapability)
			if err = m.UESecCapability.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UESecCapability = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UESecCapability.UnmarshalBinary")
			}
		case RegReqIEIReqNSSAI: // TLV, 4-74B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqNSSAI")
			}
			if m.ReqNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqNSSAI = new(ie.NSSAI)
			if err = m.ReqNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqNSSAI = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqNSSAI.UnmarshalBinary")
			}
		case RegReqIEILastVisitedRegisteredTAI: // TV, 7B
			ieLen = 6
			if m.LastVisitedRegisteredTAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.LastVisitedRegisteredTAI = new(ie.TrackingAreaId5GS)
			if err = m.LastVisitedRegisteredTAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.LastVisitedRegisteredTAI = nil
					continue
				}
				return errors.Wrap(err, "RegReq.LastVisitedRegisteredTAI.UnmarshalBinary")
			}
		case RegReqIEIS1UENwCapability: // TLV, 4-15B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of S1UENwCapability")
			}
			if m.S1UENwCapability != nil {
				reader.Next(int(ieLen))
				break
			}
			bs := reader.Next(int(ieLen))
			if len(bs) != int(ieLen) {
				return errors.Errorf(
					"RegReq.S1UENwCapability.UnmarshalBinary(): expected length: %d, but get %d",
					ieLen, len(bs))
			}
			m.S1UENwCapability = new(ie.S1UENwCapability)
			if err = m.S1UENwCapability.UnmarshalBinary(bs); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.S1UENwCapability = nil
					continue
				}
				return errors.Wrap(err, "RegReq.S1UENwCapability.UnmarshalBinary")
			}
		case RegReqIEIUplinkDataStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UplinkDataStatus")
			}
			if m.UplinkDataStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UplinkDataStatus = new(ie.UplinkDataStatus)
			if err = m.UplinkDataStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UplinkDataStatus = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UplinkDataStatus.UnmarshalBinary")
			}
		case RegReqIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of PDUSessStatus")
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
				return errors.Wrap(err, "RegReq.PDUSessStatus.UnmarshalBinary")
			}
		case RegReqIEIMICOInd: // TV, 1B
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
				return errors.Wrap(err, "RegReq.MICOInd.UnmarshalBinary")
			}
		case RegReqIEIUEStatus: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UEStatus")
			}
			if m.UEStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UEStatus = new(ie.UEStatus)
			if err = m.UEStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UEStatus = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UEStatus.UnmarshalBinary")
			}
		case RegReqIEIAdditionalGUTI: // TLV-E, 14B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of AdditionalGUTI")
			}
			if m.AdditionalGUTI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalGUTI = new(ie.MobileId5GS)
			if err = m.AdditionalGUTI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalGUTI = nil
					continue
				}
				return errors.Wrap(err, "RegReq.AdditionalGUTI.UnmarshalBinary")
			}
		case RegReqIEIAllowedPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of AllowedPDUSessStatus")
			}
			if m.AllowedPDUSessStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AllowedPDUSessStatus = new(ie.AllowedPDUSessStatus)
			if err = m.AllowedPDUSessStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AllowedPDUSessStatus = nil
					continue
				}
				return errors.Wrap(err, "RegReq.AllowedPDUSessStatus.UnmarshalBinary")
			}
		case RegReqIEIUesUsageSetting: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UesUsageSetting")
			}
			if m.UesUsageSetting != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UesUsageSetting = new(ie.UesUsageSetting)
			if err = m.UesUsageSetting.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UesUsageSetting = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UesUsageSetting.UnmarshalBinary")
			}
		case RegReqIEIReqDRXParams: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqDRXParams")
			}
			if m.ReqDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqDRXParams = new(ie.DRXParams5GS)
			if err = m.ReqDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqDRXParams.UnmarshalBinary")
			}
		case RegReqIEIEPSNASMsgCntr: // TLV-E, 4-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of EPSNASMsgCntr")
			}
			if m.EPSNASMsgCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EPSNASMsgCntr = new(ie.EPSNASMsgCntr)
			if err = m.EPSNASMsgCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EPSNASMsgCntr = nil
					continue
				}
				return errors.Wrap(err, "RegReq.EPSNASMsgCntr.UnmarshalBinary")
			}
		case RegReqIEILADNInd: // TLV-E, 3-811B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of LADNInd")
			}
			if m.LADNInd != nil {
				reader.Next(int(ieLen))
				break
			}
			m.LADNInd = new(ie.LADNInd)
			if err = m.LADNInd.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.LADNInd = nil
					continue
				}
				return errors.Wrap(err, "RegReq.LADNInd.UnmarshalBinary")
			}
		case RegReqIEIPayloadCntrType: // TV, 1B
			if m.PayloadCntrType != nil {
				break
			}
			m.PayloadCntrType = new(ie.PayloadCntrType)
			if err = m.PayloadCntrType.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PayloadCntrType = nil
					continue
				}
				return errors.Wrap(err, "RegReq.PayloadCntrType.UnmarshalBinary")
			}
		case RegReqIEIPayloadCntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of PayloadCntr")
			}
			if m.PayloadCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			if m.PayloadCntrType == nil {
				return errors.Errorf("RegReq UnmarshalBinary no PayloadCntrType for PayloadCntr")
			}
			m.PayloadCntr = new(ie.PayloadCntr)
			if err = m.PayloadCntr.UnmarshalBinary(
				reader.Next(int(ieLen)), m.PayloadCntrType.Value); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PayloadCntr = nil
					continue
				}
				return errors.Wrap(err, "RegReq.PayloadCntr.UnmarshalBinary")
			}
		case RegReqIEINwSlicingInd: // TV, 1B
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
				return errors.Wrap(err, "RegReq.NwSlicingInd.UnmarshalBinary")
			}
		case RegReqIEIUpdateType5GS: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UpdateType5GS")
			}
			if m.UpdateType5GS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UpdateType5GS = new(ie.UpdateType5GS)
			if err = m.UpdateType5GS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UpdateType5GS = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UpdateType5GS.UnmarshalBinary")
			}
		case RegReqIEIMobileStationClassmark2: // TLV, 5B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of MobileStationClassmark2")
			}
			if m.MobileStationClassmark2 != nil {
				reader.Next(int(ieLen))
				break
			}
			m.MobileStationClassmark2 = new(ie.MobileStationClassmark2)
			if err = m.MobileStationClassmark2.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.MobileStationClassmark2 = nil
					continue
				}
				return errors.Wrap(err, "RegReq.MobileStationClassmark2.UnmarshalBinary")
			}
		case RegReqIEISupportedCodecs: // TLV, 5-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of SupportedCodecs")
			}
			if m.SupportedCodecs != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SupportedCodecs = new(ie.SupportedCodecList)
			if err = m.SupportedCodecs.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SupportedCodecs = nil
					continue
				}
				return errors.Wrap(err, "RegReq.SupportedCodecs.UnmarshalBinary")
			}
		case RegReqIEINASMsgCntr: // TLV-E, 4-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of NASMsgCntr")
			}
			if m.NASMsgCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NASMsgCntr = new(ie.NASMsgCntr)
			if err = m.NASMsgCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NASMsgCntr = nil
					continue
				}
				return errors.Wrap(err, "RegReq.NASMsgCntr.UnmarshalBinary")
			}
		case RegReqIEIEPSBearerCtxStatus: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of EPSBearerCtxStatus")
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
				return errors.Wrap(err, "RegReq.EPSBearerCtxStatus.UnmarshalBinary")
			}
		case RegReqIEIReqExtendedDRXParams: // TLV, 3-4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqExtendedDRXParams")
			}
			if m.ReqExtendedDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqExtendedDRXParams = new(ie.ExtendedDRXParams)
			if err = m.ReqExtendedDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqExtendedDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqExtendedDRXParams.UnmarshalBinary")
			}
		case RegReqIEIT3324Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of T3324Value")
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
				return errors.Wrap(err, "RegReq.T3324Value.UnmarshalBinary")
			}
		case RegReqIEIUERadioCapabilityID: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UERadioCapabilityID")
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
				return errors.Wrap(err, "RegReq.UERadioCapabilityID.UnmarshalBinary")
			}
		case RegReqIEIReqMappedNSSAI: // TLV, 3-42B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqMappedNSSAI")
			}
			if m.ReqMappedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqMappedNSSAI = new(ie.MappedNSSAI)
			if err = m.ReqMappedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqMappedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqMappedNSSAI.UnmarshalBinary")
			}
		case RegReqIEIAdditionalInfoReq: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of AdditionalInfoReq")
			}
			if m.AdditionalInfoReq != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalInfoReq = new(ie.AdditionalInfoReq)
			if err = m.AdditionalInfoReq.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalInfoReq = nil
					continue
				}
				return errors.Wrap(err, "RegReq.AdditionalInfoReq.UnmarshalBinary")
			}
		case RegReqIEIReqWUSAssistanceInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqWUSAssistanceInfo")
			}
			if m.ReqWUSAssistanceInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqWUSAssistanceInfo = new(ie.WUSAssistanceInfo)
			if err = m.ReqWUSAssistanceInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqWUSAssistanceInfo = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqWUSAssistanceInfo.UnmarshalBinary")
			}
		case RegReqIEIN5GCInd: // TV, 1B
			if m.N5GCInd != nil {
				break
			}
			m.N5GCInd = new(ie.N5GCInd)
			if err = m.N5GCInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.N5GCInd = nil
					continue
				}
				return errors.Wrap(err, "RegReq.N5GCInd.UnmarshalBinary")
			}
		case RegReqIEIReqNBN1ModeDRXParams: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqNBN1ModeDRXParams")
			}
			if m.ReqNBN1ModeDRXParams != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqNBN1ModeDRXParams = new(ie.NBN1ModeDRXParams)
			if err = m.ReqNBN1ModeDRXParams.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqNBN1ModeDRXParams = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqNBN1ModeDRXParams.UnmarshalBinary")
			}
		case RegReqIEIUEReqType: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of UEReqType")
			}
			if m.UEReqType != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UEReqType = new(ie.UEReqType)
			if err = m.UEReqType.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UEReqType = nil
					continue
				}
				return errors.Wrap(err, "RegReq.UEReqType.UnmarshalBinary")
			}
		case RegReqIEIPagingRestriction: // TLV, 3-35B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of PagingRestriction")
			}
			if m.PagingRestriction != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PagingRestriction = new(ie.PagingRestriction)
			if err = m.PagingRestriction.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PagingRestriction = nil
					continue
				}
				return errors.Wrap(err, "RegReq.PagingRestriction.UnmarshalBinary")
			}
		case RegReqIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "RegReq.SvcLvlAACntr.UnmarshalBinary")
			}
		case RegReqIEINID: // TLV, 8B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of NID")
			}
			if m.NID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NID = new(ie.NID)
			if err = m.NID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NID = nil
					continue
				}
				return errors.Wrap(err, "RegReq.NID.UnmarshalBinary")
			}
		case RegReqIEIdisasterPlmnMS: // TLV, 5B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of disasterPlmnMS")
			}
			if m.DisasterPlmnMS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DisasterPlmnMS = new(ie.PlmnId)
			if err = m.DisasterPlmnMS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DisasterPlmnMS = nil
					continue
				}
				return errors.Wrap(err, "RegReq.DisasterPlmnMS.UnmarshalBinary")
			}
		case RegReqIEIReqPEIPSAssistanceInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqPEIPSAssistanceInfo")
			}
			if m.ReqPEIPSAssistanceInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqPEIPSAssistanceInfo = new(ie.PEIPSAssistanceInfo)
			if err = m.ReqPEIPSAssistanceInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqPEIPSAssistanceInfo = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqPEIPSAssistanceInfo.UnmarshalBinary")
			}
		case RegReqIEIReqT3512Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegReq UnmarshalBinary getIeLen of ReqT3512Value")
			}
			if m.ReqT3512Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqT3512Value = new(ie.GPRSTimer3)
			if err = m.ReqT3512Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqT3512Value = nil
					continue
				}
				return errors.Wrap(err, "RegReq.ReqT3512Value.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RegReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RegReq) MarshalBinary() ([]byte, error) {
	if m.RegType5GS == nil || m.Ngksi == nil || m.MobileId5GS == nil {
		return nil, errors.Errorf("RegType5GS=%v Ngksi=%v MobileId5GS=%v must present in RegReq",
			m.RegType5GS, m.Ngksi, m.MobileId5GS)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRegReq),
	})

	// regtype5gs, V, 1/2B
	tmp := [1]byte{}
	regtype5gs, err := m.RegType5GS.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RegReq.RegType5GS.MarshalBinary()")
	}

	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RegReq.Ngksi.MarshalBinary()")
	}
	tmp[0] = ie.SetHalfValue(ngksi[0], regtype5gs[0])
	writer.Write(tmp[:])

	// mobileid5gs, LV-E, 6-nB
	mobileid5gs, err := m.MobileId5GS.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RegReq.MobileId5GS.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(mobileid5gs))); err != nil {
		return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write MobileId5GS")
	}
	writer.Write(mobileid5gs)

	// m.NoncurrentNativeNASKeySetId TV, 1B, IEI=0xC0, >= 0x80 !
	if m.NoncurrentNativeNASKeySetId != nil {
		out, err := m.NoncurrentNativeNASKeySetId.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.NoncurrentNativeNASKeySetId.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegReqIEINoncurrentNativeNASKeySetId)
	}

	// m.Capability5GMM TLV, 3-15B, IEI=0x10
	if m.Capability5GMM != nil {
		out, err := m.Capability5GMM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.Capability5GMM.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEICapability5GMM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UESecCapability TLV, 4-10B, IEI=0x2E
	if m.UESecCapability != nil {
		out, err := m.UESecCapability.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UESecCapability.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUESecCapability)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqNSSAI TLV, 4-74B, IEI=0x2F
	if m.ReqNSSAI != nil {
		out, err := m.ReqNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.LastVisitedRegisteredTAI TV, 7B, IEI=0x52
	if m.LastVisitedRegisteredTAI != nil {
		out, err := m.LastVisitedRegisteredTAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.LastVisitedRegisteredTAI.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEILastVisitedRegisteredTAI)
		writer.Write(out)
	}

	// m.S1UENwCapability TLV, 4-15B, IEI=0x17
	if m.S1UENwCapability != nil {
		out, err := m.S1UENwCapability.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.S1UENwCapability.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIS1UENwCapability)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UplinkDataStatus TLV, 4-34B, IEI=0x40
	if m.UplinkDataStatus != nil {
		out, err := m.UplinkDataStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UplinkDataStatus.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUplinkDataStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.MICOInd TV, 1B, IEI=0xB0, >= 0x80 !
	if m.MICOInd != nil {
		out, err := m.MICOInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.MICOInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegReqIEIMICOInd)
	}

	// m.UEStatus TLV, 3B, IEI=0x2B
	if m.UEStatus != nil {
		out, err := m.UEStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UEStatus.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUEStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalGUTI TLV-E, 14B, IEI=0x77
	if m.AdditionalGUTI != nil {
		out, err := m.AdditionalGUTI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.AdditionalGUTI.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIAdditionalGUTI)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write AdditionalGUTI")
		}
		writer.Write(out)
	}

	// m.AllowedPDUSessStatus TLV, 4-34B, IEI=0x25
	if m.AllowedPDUSessStatus != nil {
		out, err := m.AllowedPDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.AllowedPDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIAllowedPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UesUsageSetting TLV, 3B, IEI=0x18
	if m.UesUsageSetting != nil {
		out, err := m.UesUsageSetting.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UesUsageSetting.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUesUsageSetting)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqDRXParams TLV, 3B, IEI=0x51
	if m.ReqDRXParams != nil {
		out, err := m.ReqDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EPSNASMsgCntr TLV-E, 4-nB, IEI=0x70
	if m.EPSNASMsgCntr != nil {
		out, err := m.EPSNASMsgCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.EPSNASMsgCntr.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIEPSNASMsgCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write EPSNASMsgCntr")
		}
		writer.Write(out)
	}

	// m.LADNInd TLV-E, 3-811B, IEI=0x74
	if m.LADNInd != nil {
		out, err := m.LADNInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.LADNInd.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEILADNInd)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write LADNInd")
		}
		writer.Write(out)
	}

	// m.PayloadCntrType TV, 1B, IEI=0x80, >= 0x80 !
	if m.PayloadCntrType != nil {
		out, err := m.PayloadCntrType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.PayloadCntrType.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegReqIEIPayloadCntrType)
	}

	// m.PayloadCntr TLV-E, 4-65538B, IEI=0x7B
	if m.PayloadCntr != nil {
		out, err := m.PayloadCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.PayloadCntr.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIPayloadCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write PayloadCntr")
		}
		writer.Write(out)
	}

	// m.NwSlicingInd TV, 1B, IEI=0x90, >= 0x80 !
	if m.NwSlicingInd != nil {
		out, err := m.NwSlicingInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.NwSlicingInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegReqIEINwSlicingInd)
	}

	// m.UpdateType5GS TLV, 3B, IEI=0x53
	if m.UpdateType5GS != nil {
		out, err := m.UpdateType5GS.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UpdateType5GS.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUpdateType5GS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.MobileStationClassmark2 TLV, 5B, IEI=0x41
	if m.MobileStationClassmark2 != nil {
		out, err := m.MobileStationClassmark2.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.MobileStationClassmark2.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIMobileStationClassmark2)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SupportedCodecs TLV, 5-nB, IEI=0x42
	if m.SupportedCodecs != nil {
		out, err := m.SupportedCodecs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.SupportedCodecs.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEISupportedCodecs)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NASMsgCntr TLV-E, 4-nB, IEI=0x71
	if m.NASMsgCntr != nil {
		out, err := m.NASMsgCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.NASMsgCntr.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEINASMsgCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write NASMsgCntr")
		}
		writer.Write(out)
	}

	// m.EPSBearerCtxStatus TLV, 4B, IEI=0x60
	if m.EPSBearerCtxStatus != nil {
		out, err := m.EPSBearerCtxStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.EPSBearerCtxStatus.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIEPSBearerCtxStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqExtendedDRXParams TLV, 3-4B, IEI=0x6E
	if m.ReqExtendedDRXParams != nil {
		out, err := m.ReqExtendedDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqExtendedDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqExtendedDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3324Value TLV, 3B, IEI=0x6A
	if m.T3324Value != nil {
		out, err := m.T3324Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.T3324Value.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIT3324Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UERadioCapabilityID TLV, 3-nB, IEI=0x67
	if m.UERadioCapabilityID != nil {
		out, err := m.UERadioCapabilityID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UERadioCapabilityID.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUERadioCapabilityID)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqMappedNSSAI TLV, 3-42B, IEI=0x35
	if m.ReqMappedNSSAI != nil {
		out, err := m.ReqMappedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqMappedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqMappedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalInfoReq TLV, 3B, IEI=0x48
	if m.AdditionalInfoReq != nil {
		out, err := m.AdditionalInfoReq.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.AdditionalInfoReq.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIAdditionalInfoReq)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqWUSAssistanceInfo TLV, 3-nB, IEI=0x1A
	if m.ReqWUSAssistanceInfo != nil {
		out, err := m.ReqWUSAssistanceInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqWUSAssistanceInfo.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqWUSAssistanceInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.N5GCInd TV, 1B, IEI=0xA0, >= 0x80 !
	if m.N5GCInd != nil {
		out, err := m.N5GCInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.N5GCInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | RegReqIEIN5GCInd)
	}

	// m.ReqNBN1ModeDRXParams TLV, 3B, IEI=0x30
	if m.ReqNBN1ModeDRXParams != nil {
		out, err := m.ReqNBN1ModeDRXParams.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqNBN1ModeDRXParams.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqNBN1ModeDRXParams)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UEReqType TLV, 3B, IEI=0x29
	if m.UEReqType != nil {
		out, err := m.UEReqType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.UEReqType.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIUEReqType)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PagingRestriction TLV, 3-35B, IEI=0x28
	if m.PagingRestriction != nil {
		out, err := m.PagingRestriction.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.PagingRestriction.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIPagingRestriction)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegReq) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}

	// m.NID TLV, 8B, IEI=0x32
	if m.NID != nil {
		out, err := m.NID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.NID.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEINID)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterPlmnMS TLV, 5B, IEI=0x16
	if m.DisasterPlmnMS != nil {
		out := make([]byte, 3)
		err := m.DisasterPlmnMS.MarshalBinary(out)
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.DisasterPlmnMS.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIdisasterPlmnMS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqPEIPSAssistanceInfo TLV, 3-nB, IEI=0x2A
	if m.ReqPEIPSAssistanceInfo != nil {
		out, err := m.ReqPEIPSAssistanceInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqPEIPSAssistanceInfo.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqPEIPSAssistanceInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqT3512Value TLV, 3B, IEI=0x3B
	if m.ReqT3512Value != nil {
		out, err := m.ReqT3512Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegReq.ReqT3512Value.MarshalBinary()")
		}
		writer.WriteByte(RegReqIEIReqT3512Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
