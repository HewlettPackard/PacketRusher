package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SecModeComplete{}

// SecModeComplete is detailed in 8.2.26 Security mode complete, 24.501
type SecModeComplete struct {
	IMEISV       *ie.MobileId5GS // TLV-E,      12B, 9.11.3.4
	NASMsgCntr   *ie.NASMsgCntr  // TLV-E,     4-nB, 9.11.3.33
	NonimeisvPEI *ie.MobileId5GS // TLV-E,     7-nB, 9.11.3.4
}

const (
	SecModeCompleteIEIIMEISV       uint8 = 0x77
	SecModeCompleteIEINASMsgCntr   uint8 = 0x71
	SecModeCompleteIEINonimeisvPEI uint8 = 0x78
)

func (m *SecModeComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SecModeComplete) MsgType() MsgType {
	return MsgTypeSecModeComplete
}

func (m *SecModeComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SecModeComplete len(b)=%d, < GmmHdrLen(%d)",
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
		case SecModeCompleteIEIIMEISV: // TLV-E, 12B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeComplete UnmarshalBinary getIeLen of IMEISV")
			}
			if m.IMEISV != nil {
				reader.Next(int(ieLen))
				break
			}
			m.IMEISV = new(ie.MobileId5GS)
			if err = m.IMEISV.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.IMEISV = nil
					continue
				}
				return errors.Wrap(err, "SecModeComplete.IMEISV.UnmarshalBinary")
			}
		case SecModeCompleteIEINASMsgCntr: // TLV-E, 4-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeComplete UnmarshalBinary getIeLen of NASMsgCntr")
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
				return errors.Wrap(err, "SecModeComplete.NASMsgCntr.UnmarshalBinary")
			}
		case SecModeCompleteIEINonimeisvPEI: // TLV-E, 7-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeComplete UnmarshalBinary getIeLen of NonimeisvPEI")
			}
			if m.NonimeisvPEI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NonimeisvPEI = new(ie.MobileId5GS)
			if err = m.NonimeisvPEI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NonimeisvPEI = nil
					continue
				}
				return errors.Wrap(err, "SecModeComplete.NonimeisvPEI.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("SecModeComplete unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *SecModeComplete) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSecModeComplete),
	})

	// m.IMEISV TLV-E, 12B, IEI=0x77
	if m.IMEISV != nil {
		out, err := m.IMEISV.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeComplete.IMEISV.MarshalBinary()")
		}
		writer.WriteByte(SecModeCompleteIEIIMEISV)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SecModeComplete) MarshalBinary() binary write IMEISV")
		}
		writer.Write(out)
	}

	// m.NASMsgCntr TLV-E, 4-nB, IEI=0x71
	if m.NASMsgCntr != nil {
		out, err := m.NASMsgCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeComplete.NASMsgCntr.MarshalBinary()")
		}
		writer.WriteByte(SecModeCompleteIEINASMsgCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SecModeComplete) MarshalBinary() binary write NASMsgCntr")
		}
		writer.Write(out)
	}

	// m.NonimeisvPEI TLV-E, 7-nB, IEI=0x78
	if m.NonimeisvPEI != nil {
		out, err := m.NonimeisvPEI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeComplete.NonimeisvPEI.MarshalBinary()")
		}
		writer.WriteByte(SecModeCompleteIEINonimeisvPEI)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SecModeComplete) MarshalBinary() binary write NonimeisvPEI")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
