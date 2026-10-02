package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessAuthResult{}

// PDUSessAuthResult is detailed in 8.3.6 PDU session authentication result, 24.501
type PDUSessAuthResult struct {
	PDUSessId           uint8
	PTI                 uint8
	EAPMsg              *ie.EAPMsg              // TLV-E,  7-1503B, 9.11.2.2
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
}

const (
	PDUSessAuthResultIEIEAPMsg              uint8 = 0x78
	PDUSessAuthResultIEIExtendedProtCfgOpts uint8 = 0x7B
)

func (m *PDUSessAuthResult) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessAuthResult) MsgType() MsgType {
	return MsgTypePDUSessAuthResult
}

func (m *PDUSessAuthResult) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessAuthResult) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessAuthResult) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessAuthResult len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// This message contains 0 Mandatory IE
	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessAuthResultIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessAuthResult UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "PDUSessAuthResult.EAPMsg.UnmarshalBinary")
			}
		case PDUSessAuthResultIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessAuthResult UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// The PDU SESSION AUTHENTICATION RESULT message is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessAuthResult.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessAuthResult unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessAuthResult) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessAuthResult),
	})

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthResult.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessAuthResultIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthResult) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthResult.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessAuthResultIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthResult) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
