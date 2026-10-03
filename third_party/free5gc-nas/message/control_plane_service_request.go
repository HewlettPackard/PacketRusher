package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &CtrlPlaneSvcReq{}

// CtrlPlaneSvcReq is detailed in 8.2.30 Control Plane Service request, 24.501
type CtrlPlaneSvcReq struct {
	CtrlPlaneSvcType     *ie.CtrlPlaneSvcType     //     V,     1/2B, 9.11.3.18D
	Ngksi                *ie.NASKeySetId          //     V,     1/2B, 9.11.3.32
	CiotSmallDataCntr    *ie.CiotSmallDataCntr    //   TLV,   4-257B, 9.11.3.18B
	PayloadCntrType      *ie.PayloadCntrType      //    TV,       1B, 9.11.3.40
	PayloadCntr          *ie.PayloadCntr          // TLV-E, 4-65538B, 9.11.3.39
	PDUSessID            *ie.PDUSessId2           //    TV,       2B, 9.11.3.41
	PDUSessStatus        *ie.PDUSessStatus        //   TLV,    4-34B, 9.11.3.44
	RelAssistanceInd     *ie.RelAssistanceInd     //    TV,       1B, 9.11.3.46A
	UplinkDataStatus     *ie.UplinkDataStatus     //   TLV,    4-34B, 9.11.3.57
	NASMsgCntr           *ie.NASMsgCntr           // TLV-E,     4-nB, 9.11.3.33
	AdditionalInfo       *ie.AdditionalInfo       //   TLV,     3-nB, 9.11.2.1
	AllowedPDUSessStatus *ie.AllowedPDUSessStatus //   TLV,    4-34B, 9.11.3.13
	UEReqType            *ie.UEReqType            //   TLV,       3B, 9.11.3.76
	PagingRestriction    *ie.PagingRestriction    //   TLV,    3-35B, 9.11.3.77
}

const (
	CtrlPlaneSvcReqIEICiotSmallDataCntr    uint8 = 0x6F
	CtrlPlaneSvcReqIEIPayloadCntrType      uint8 = 0x80
	CtrlPlaneSvcReqIEIPayloadCntr          uint8 = 0x7B
	CtrlPlaneSvcReqIEIPDUSessID            uint8 = 0x12
	CtrlPlaneSvcReqIEIPDUSessStatus        uint8 = 0x50
	CtrlPlaneSvcReqIEIRelAssistanceInd     uint8 = 0xF0
	CtrlPlaneSvcReqIEIUplinkDataStatus     uint8 = 0x40
	CtrlPlaneSvcReqIEINASMsgCntr           uint8 = 0x71
	CtrlPlaneSvcReqIEIAdditionalInfo       uint8 = 0x24
	CtrlPlaneSvcReqIEIAllowedPDUSessStatus uint8 = 0x25
	CtrlPlaneSvcReqIEIUEReqType            uint8 = 0x29
	CtrlPlaneSvcReqIEIPagingRestriction    uint8 = 0x28
)

func (m *CtrlPlaneSvcReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *CtrlPlaneSvcReq) MsgType() MsgType {
	return MsgTypeCtrlPlaneSvcReq
}

func (m *CtrlPlaneSvcReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("CtrlPlaneSvcReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, half85 := ie.GetHalfIEValue(reader.Next(1))
	m.CtrlPlaneSvcType = new(ie.CtrlPlaneSvcType) // V, 1/2B
	if err = m.CtrlPlaneSvcType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "CtrlPlaneSvcReq.CtrlPlaneSvcType.UnmarshalBinary")
	}
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half85); err != nil {
		return errors.Wrap(err, "CtrlPlaneSvcReq.Ngksi.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case CtrlPlaneSvcReqIEICiotSmallDataCntr: // TLV, 4-257B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of CiotSmallDataCntr")
			}
			if m.CiotSmallDataCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.CiotSmallDataCntr = new(ie.CiotSmallDataCntr)
			if err = m.CiotSmallDataCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CiotSmallDataCntr = nil
					continue
				}
				return errors.Wrap(err, "CtrlPlaneSvcReq.CiotSmallDataCntr.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIPayloadCntrType: // TV, 1B
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.PayloadCntrType.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIPayloadCntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of PayloadCntr")
			}
			if m.PayloadCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			if m.PayloadCntrType == nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary no PayloadCntrType for PayloadCntr")
			}
			m.PayloadCntr = new(ie.PayloadCntr)
			if err = m.PayloadCntr.UnmarshalBinary(
				reader.Next(int(ieLen)), m.PayloadCntrType.Value); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PayloadCntr = nil
					continue
				}
				return errors.Wrap(err, "CtrlPlaneSvcReq.PayloadCntr.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIPDUSessID: // TV, 2B
			ieLen = 1
			if m.PDUSessID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessID = new(ie.PDUSessId2)
			if err = m.PDUSessID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessID = nil
					continue
				}
				return errors.Wrap(err, "CtrlPlaneSvcReq.PDUSessID.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of PDUSessStatus")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.PDUSessStatus.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIRelAssistanceInd: // TV, 1B
			if m.RelAssistanceInd != nil {
				break
			}
			m.RelAssistanceInd = new(ie.RelAssistanceInd)
			if err = m.RelAssistanceInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RelAssistanceInd = nil
					continue
				}
				return errors.Wrap(err, "CtrlPlaneSvcReq.RelAssistanceInd.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIUplinkDataStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of UplinkDataStatus")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.UplinkDataStatus.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEINASMsgCntr: // TLV-E, 4-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of NASMsgCntr")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.NASMsgCntr.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIAdditionalInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of AdditionalInfo")
			}
			if m.AdditionalInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalInfo = new(ie.AdditionalInfo)
			if err = m.AdditionalInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalInfo = nil
					continue
				}
				return errors.Wrap(err, "CtrlPlaneSvcReq.AdditionalInfo.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIAllowedPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of AllowedPDUSessStatus")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.AllowedPDUSessStatus.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIUEReqType: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of UEReqType")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.UEReqType.UnmarshalBinary")
			}
		case CtrlPlaneSvcReqIEIPagingRestriction: // TLV, 3-35B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "CtrlPlaneSvcReq UnmarshalBinary getIeLen of PagingRestriction")
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
				return errors.Wrap(err, "CtrlPlaneSvcReq.PagingRestriction.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("CtrlPlaneSvcReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *CtrlPlaneSvcReq) MarshalBinary() ([]byte, error) {
	if m.CtrlPlaneSvcType == nil || m.Ngksi == nil {
		return nil, errors.Errorf("CtrlPlaneSvcType=%v Ngksi=%v must present in CtrlPlaneSvcReq",
			m.CtrlPlaneSvcType, m.Ngksi)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeCtrlPlaneSvcReq),
	})

	tmp := [1]byte{}
	// ctrlplanesvctype, V, 1/2B
	ctrlplanesvctype, err := m.CtrlPlaneSvcType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "CtrlPlaneSvcReq.CtrlPlaneSvcType.MarshalBinary()")
	}

	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "CtrlPlaneSvcReq.Ngksi.MarshalBinary()")
	}
	tmp[0] = ie.SetHalfValue(ngksi[0], ctrlplanesvctype[0])
	writer.Write(tmp[:])

	// m.CiotSmallDataCntr TLV, 4-257B, IEI=0x6F
	if m.CiotSmallDataCntr != nil {
		out, err := m.CiotSmallDataCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.CiotSmallDataCntr.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEICiotSmallDataCntr)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PayloadCntrType TV, 1B, IEI=0x80, >= 0x80 !
	if m.PayloadCntrType != nil {
		out, err := m.PayloadCntrType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.PayloadCntrType.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CtrlPlaneSvcReqIEIPayloadCntrType)
	}

	// m.PayloadCntr TLV-E, 4-65538B, IEI=0x7B
	if m.PayloadCntr != nil {
		out, err := m.PayloadCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.PayloadCntr.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIPayloadCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq) MarshalBinary() binary write PayloadCntr")
		}
		writer.Write(out)
	}

	// m.PDUSessID TV, 2B, IEI=0x12
	if m.PDUSessID != nil {
		out, err := m.PDUSessID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.PDUSessID.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIPDUSessID)
		writer.Write(out)
	}

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RelAssistanceInd TV, 1B, IEI=0xF0, >= 0x80 !
	if m.RelAssistanceInd != nil {
		out, err := m.RelAssistanceInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.RelAssistanceInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | CtrlPlaneSvcReqIEIRelAssistanceInd)
	}

	// m.UplinkDataStatus TLV, 4-34B, IEI=0x40
	if m.UplinkDataStatus != nil {
		out, err := m.UplinkDataStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.UplinkDataStatus.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIUplinkDataStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NASMsgCntr TLV-E, 4-nB, IEI=0x71
	if m.NASMsgCntr != nil {
		out, err := m.NASMsgCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.NASMsgCntr.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEINASMsgCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq) MarshalBinary() binary write NASMsgCntr")
		}
		writer.Write(out)
	}

	// m.AdditionalInfo TLV, 3-nB, IEI=0x24
	if m.AdditionalInfo != nil {
		out, err := m.AdditionalInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.AdditionalInfo.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIAdditionalInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AllowedPDUSessStatus TLV, 4-34B, IEI=0x25
	if m.AllowedPDUSessStatus != nil {
		out, err := m.AllowedPDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.AllowedPDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIAllowedPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UEReqType TLV, 3B, IEI=0x29
	if m.UEReqType != nil {
		out, err := m.UEReqType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.UEReqType.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIUEReqType)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PagingRestriction TLV, 3-35B, IEI=0x28
	if m.PagingRestriction != nil {
		out, err := m.PagingRestriction.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "CtrlPlaneSvcReq.PagingRestriction.MarshalBinary()")
		}
		writer.WriteByte(CtrlPlaneSvcReqIEIPagingRestriction)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
