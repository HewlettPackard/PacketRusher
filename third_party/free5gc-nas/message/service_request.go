package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SvcReq{}

// SvcReq is detailed in 8.2.16 Service request, 24.501
type SvcReq struct {
	Ngksi                *ie.NASKeySetId          //     V,     1/2B, 9.11.3.32
	SvcType              *ie.SvcType              //     V,     1/2B, 9.11.3.50
	TMSI5GS              *ie.MobileId5GS          //  LV-E,       9B, 9.11.3.4
	UplinkDataStatus     *ie.UplinkDataStatus     //   TLV,    4-34B, 9.11.3.57
	PDUSessStatus        *ie.PDUSessStatus        //   TLV,    4-34B, 9.11.3.44
	AllowedPDUSessStatus *ie.AllowedPDUSessStatus //   TLV,    4-34B, 9.11.3.13
	NASMsgCntr           *ie.NASMsgCntr           // TLV-E,     4-nB, 9.11.3.33
	UEReqType            *ie.UEReqType            //   TLV,       3B, 9.11.3.76
	PagingRestriction    *ie.PagingRestriction    //   TLV,    3-35B, 9.11.3.77
}

const (
	SvcReqIEIUplinkDataStatus     uint8 = 0x40
	SvcReqIEIPDUSessStatus        uint8 = 0x50
	SvcReqIEIAllowedPDUSessStatus uint8 = 0x25
	SvcReqIEINASMsgCntr           uint8 = 0x71
	SvcReqIEIUEReqType            uint8 = 0x29
	SvcReqIEIPagingRestriction    uint8 = 0x28
)

func (m *SvcReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SvcReq) MsgType() MsgType {
	return MsgTypeSvcReq
}

func (m *SvcReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SvcReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, half85 := ie.GetHalfIEValue(reader.Next(1))
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "SvcReq.Ngksi.UnmarshalBinary")
	}
	m.SvcType = new(ie.SvcType) // V, 1/2B
	if err = m.SvcType.UnmarshalBinary(half85); err != nil {
		return errors.Wrap(err, "SvcReq.SvcType.UnmarshalBinary")
	}

	m.TMSI5GS = new(ie.MobileId5GS) // LV-E, 9B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of TMSI5GS")
	}
	if err = m.TMSI5GS.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "SvcReq.TMSI5GS.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case SvcReqIEIUplinkDataStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of UplinkDataStatus")
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
				return errors.Wrap(err, "SvcReq.UplinkDataStatus.UnmarshalBinary")
			}
		case SvcReqIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of PDUSessStatus")
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
				return errors.Wrap(err, "SvcReq.PDUSessStatus.UnmarshalBinary")
			}
		case SvcReqIEIAllowedPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of AllowedPDUSessStatus")
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
				return errors.Wrap(err, "SvcReq.AllowedPDUSessStatus.UnmarshalBinary")
			}
		case SvcReqIEINASMsgCntr: // TLV-E, 4-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of NASMsgCntr")
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
				return errors.Wrap(err, "SvcReq.NASMsgCntr.UnmarshalBinary")
			}
		case SvcReqIEIUEReqType: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of UEReqType")
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
				return errors.Wrap(err, "SvcReq.UEReqType.UnmarshalBinary")
			}
		case SvcReqIEIPagingRestriction: // TLV, 3-35B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcReq UnmarshalBinary getIeLen of PagingRestriction")
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
				return errors.Wrap(err, "SvcReq.PagingRestriction.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("SvcReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *SvcReq) MarshalBinary() ([]byte, error) {
	if m.Ngksi == nil || m.SvcType == nil || m.TMSI5GS == nil {
		return nil, errors.Errorf("Ngksi=%v SvcType=%v TMSI5GS=%v must present in SvcReq",
			m.Ngksi, m.SvcType, m.TMSI5GS)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSvcReq),
	})

	// ngksi, V, 1/2B
	tmp := [1]byte{}
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcReq.Ngksi.MarshalBinary()")
	}

	// svctype, V, 1/2B
	svctype, err := m.SvcType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcReq.SvcType.MarshalBinary()")
	}
	tmp[0] = ie.SetHalfValue(svctype[0], ngksi[0])
	writer.Write(tmp[:])

	// tmsi5gs, LV-E, 9B
	tmsi5gs, err := m.TMSI5GS.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcReq.TMSI5GS.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(tmsi5gs))); err != nil {
		return nil, errors.Wrap(err, "SvcReq) MarshalBinary() binary write TMSI5GS")
	}
	writer.Write(tmsi5gs)

	// m.UplinkDataStatus TLV, 4-34B, IEI=0x40
	if m.UplinkDataStatus != nil {
		out, err := m.UplinkDataStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.UplinkDataStatus.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEIUplinkDataStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AllowedPDUSessStatus TLV, 4-34B, IEI=0x25
	if m.AllowedPDUSessStatus != nil {
		out, err := m.AllowedPDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.AllowedPDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEIAllowedPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NASMsgCntr TLV-E, 4-nB, IEI=0x71
	if m.NASMsgCntr != nil {
		out, err := m.NASMsgCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.NASMsgCntr.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEINASMsgCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcReq) MarshalBinary() binary write NASMsgCntr")
		}
		writer.Write(out)
	}

	// m.UEReqType TLV, 3B, IEI=0x29
	if m.UEReqType != nil {
		out, err := m.UEReqType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.UEReqType.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEIUEReqType)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PagingRestriction TLV, 3-35B, IEI=0x28
	if m.PagingRestriction != nil {
		out, err := m.PagingRestriction.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcReq.PagingRestriction.MarshalBinary()")
		}
		writer.WriteByte(SvcReqIEIPagingRestriction)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
