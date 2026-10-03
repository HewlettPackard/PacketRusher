package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RegComplete{}

// RegComplete is detailed in 8.2.8 Registration complete, 24.501
type RegComplete struct {
	SORTransparentCntr *ie.SORTransparentCntr // TLV-E,      20B, 9.11.3.51
}

const (
	RegCompleteIEISORTransparentCntr uint8 = 0x73
)

func (m *RegComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RegComplete) MsgType() MsgType {
	return MsgTypeRegComplete
}

func (m *RegComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RegComplete len(b)=%d, < GmmHdrLen(%d)",
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
		case RegCompleteIEISORTransparentCntr: // TLV-E, 20B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegComplete UnmarshalBinary getIeLen of SORTransparentCntr")
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
				return errors.Wrap(err, "RegComplete.SORTransparentCntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RegComplete unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RegComplete) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRegComplete),
	})

	// m.SORTransparentCntr TLV-E, 20B, IEI=0x73
	if m.SORTransparentCntr != nil {
		out, err := m.SORTransparentCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegComplete.SORTransparentCntr.MarshalBinary()")
		}
		writer.WriteByte(RegCompleteIEISORTransparentCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegComplete) MarshalBinary() binary write SORTransparentCntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
