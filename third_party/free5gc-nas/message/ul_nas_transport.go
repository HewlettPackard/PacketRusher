package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &ULNASTransport{}

// ULNASTransport is detailed in 8.2.10 UL NAS transport, 24.501
type ULNASTransport struct {
	PayloadCntrType  *ie.PayloadCntrType  //     V,     1/2B, 9.11.3.40
	SpareHalfOctet   struct{}             //     V,     1/2B, 9.5
	PayloadCntr      *ie.PayloadCntr      //  LV-E, 3-65537B, 9.11.3.39
	PDUSessID        *ie.PDUSessId2       //    TV,       2B, 9.11.3.41
	OldPDUSessID     *ie.PDUSessId2       //    TV,       2B, 9.11.3.41
	ReqType          *ie.ReqType          //    TV,       1B, 9.11.3.47
	SNSSAI           *ie.SNSSAI           //   TLV,    3-10B, 9.11.2.8
	DNN              *ie.DNN              //   TLV,   3-102B, 9.11.2.1B
	AdditionalInfo   *ie.AdditionalInfo   //   TLV,     3-nB, 9.11.2.1
	MAPDUSessInfo    *ie.MAPDUSessInfo    //    TV,       1B, 9.11.3.31A
	RelAssistanceInd *ie.RelAssistanceInd //    TV,       1B, 9.11.3.46A
}

const (
	ULNASTransportIEIPDUSessID        uint8 = 0x12
	ULNASTransportIEIOldPDUSessID     uint8 = 0x59
	ULNASTransportIEIReqType          uint8 = 0x80
	ULNASTransportIEISNSSAI           uint8 = 0x22
	ULNASTransportIEIDNN              uint8 = 0x25
	ULNASTransportIEIAdditionalInfo   uint8 = 0x24
	ULNASTransportIEIMAPDUSessInfo    uint8 = 0xA0
	ULNASTransportIEIRelAssistanceInd uint8 = 0xF0
)

func (m *ULNASTransport) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *ULNASTransport) MsgType() MsgType {
	return MsgTypeULNASTransport
}

func (m *ULNASTransport) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("ULNASTransport len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.PayloadCntrType = new(ie.PayloadCntrType) // V, 1/2B
	if err = m.PayloadCntrType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "ULNASTransport.PayloadCntrType.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	m.PayloadCntr = new(ie.PayloadCntr) // LV-E, 3-65537B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "ULNASTransport UnmarshalBinary getIeLen of PayloadCntr")
	}
	if err = m.PayloadCntr.UnmarshalBinary(
		reader.Next(int(ieLen)), m.PayloadCntrType.Value); err != nil {
		return errors.Wrap(err, "ULNASTransport.PayloadCntr.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case ULNASTransportIEIPDUSessID: // TV, 2B
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
				return errors.Wrap(err, "ULNASTransport.PDUSessID.UnmarshalBinary")
			}
		case ULNASTransportIEIOldPDUSessID: // TV, 2B
			ieLen = 1
			if m.OldPDUSessID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.OldPDUSessID = new(ie.PDUSessId2)
			if err = m.OldPDUSessID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.OldPDUSessID = nil
					continue
				}
				return errors.Wrap(err, "ULNASTransport.OldPDUSessID.UnmarshalBinary")
			}
		case ULNASTransportIEIReqType: // TV, 1B
			if m.ReqType != nil {
				break
			}
			m.ReqType = new(ie.ReqType)
			if err = m.ReqType.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqType = nil
					continue
				}
				return errors.Wrap(err, "ULNASTransport.ReqType.UnmarshalBinary")
			}
		case ULNASTransportIEISNSSAI: // TLV, 3-10B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "ULNASTransport UnmarshalBinary getIeLen of SNSSAI")
			}
			if m.SNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SNSSAI = new(ie.SNSSAI)
			if err = m.SNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SNSSAI = nil
					continue
				}
				return errors.Wrap(err, "ULNASTransport.SNSSAI.UnmarshalBinary")
			}
		case ULNASTransportIEIDNN: // TLV, 3-102B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "ULNASTransport UnmarshalBinary getIeLen of DNN")
			}
			if m.DNN != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DNN = new(ie.DNN)
			if err = m.DNN.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DNN = nil
					continue
				}
				return errors.Wrap(err, "ULNASTransport.DNN.UnmarshalBinary")
			}
		case ULNASTransportIEIAdditionalInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "ULNASTransport UnmarshalBinary getIeLen of AdditionalInfo")
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
				return errors.Wrap(err, "ULNASTransport.AdditionalInfo.UnmarshalBinary")
			}
		case ULNASTransportIEIMAPDUSessInfo: // TV, 1B
			if m.MAPDUSessInfo != nil {
				break
			}
			m.MAPDUSessInfo = new(ie.MAPDUSessInfo)
			if err = m.MAPDUSessInfo.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.MAPDUSessInfo = nil
					continue
				}
				return errors.Wrap(err, "ULNASTransport.MAPDUSessInfo.UnmarshalBinary")
			}
		case ULNASTransportIEIRelAssistanceInd: // TV, 1B
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
				return errors.Wrap(err, "ULNASTransport.RelAssistanceInd.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("ULNASTransport unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *ULNASTransport) MarshalBinary() ([]byte, error) {
	if m.PayloadCntrType == nil || m.PayloadCntr == nil {
		return nil, errors.Errorf("PayloadCntrType=%v PayloadCntr=%v must present in ULNASTransport",
			m.PayloadCntrType, m.PayloadCntr)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeULNASTransport),
	})

	tmp := [1]byte{}
	// payloadcntrtype, V, 1/2B
	payloadcntrtype, err := m.PayloadCntrType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "ULNASTransport.PayloadCntrType.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, payloadcntrtype[0])
	writer.Write(tmp[:])

	// payloadcntr, LV-E, 3-65537B
	payloadcntr, err := m.PayloadCntr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "ULNASTransport.PayloadCntr.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(payloadcntr))); err != nil {
		return nil, errors.Wrap(err, "ULNASTransport) MarshalBinary() binary write PayloadCntr")
	}
	writer.Write(payloadcntr)

	// m.PDUSessID TV, 2B, IEI=0x12
	if m.PDUSessID != nil {
		out, err := m.PDUSessID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.PDUSessID.MarshalBinary()")
		}
		writer.WriteByte(ULNASTransportIEIPDUSessID)
		writer.Write(out)
	}

	// m.OldPDUSessID TV, 2B, IEI=0x59
	if m.OldPDUSessID != nil {
		out, err := m.OldPDUSessID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.OldPDUSessID.MarshalBinary()")
		}
		writer.WriteByte(ULNASTransportIEIOldPDUSessID)
		writer.Write(out)
	}

	// m.ReqType TV, 1B, IEI=0x80, >= 0x80 !
	if m.ReqType != nil {
		out, err := m.ReqType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.ReqType.MarshalBinary()")
		}
		writer.WriteByte(out[0] | ULNASTransportIEIReqType)
	}

	// m.SNSSAI TLV, 3-10B, IEI=0x22
	if m.SNSSAI != nil {
		out, err := m.SNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.SNSSAI.MarshalBinary()")
		}
		writer.WriteByte(ULNASTransportIEISNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DNN TLV, 3-102B, IEI=0x25
	if m.DNN != nil {
		out, err := m.DNN.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.DNN.MarshalBinary()")
		}
		writer.WriteByte(ULNASTransportIEIDNN)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalInfo TLV, 3-nB, IEI=0x24
	if m.AdditionalInfo != nil {
		out, err := m.AdditionalInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.AdditionalInfo.MarshalBinary()")
		}
		writer.WriteByte(ULNASTransportIEIAdditionalInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.MAPDUSessInfo TV, 1B, IEI=0xA0, >= 0x80 !
	if m.MAPDUSessInfo != nil {
		out, err := m.MAPDUSessInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.MAPDUSessInfo.MarshalBinary()")
		}
		writer.WriteByte(out[0] | ULNASTransportIEIMAPDUSessInfo)
	}

	// m.RelAssistanceInd TV, 1B, IEI=0xF0, >= 0x80 !
	if m.RelAssistanceInd != nil {
		out, err := m.RelAssistanceInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "ULNASTransport.RelAssistanceInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | ULNASTransportIEIRelAssistanceInd)
	}
	return writer.Bytes(), nil
}
