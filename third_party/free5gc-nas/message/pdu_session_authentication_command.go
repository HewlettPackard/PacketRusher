package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessAuthCmd{}

// PDUSessAuthCmd is detailed in 8.3.4 PDU session authentication command, 24.501
type PDUSessAuthCmd struct {
	PDUSessId           uint8
	PTI                 uint8
	EAPMsg              *ie.EAPMsg              //  LV-E,  6-1502B, 9.11.2.2
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
	// TODO: Add IE
	// RemoteUEHandlingInfoList *ie.RemoteUEHandlingInfoList // TLV-E, 4-65538B, 9.11.4.35
}

const (
	PDUSessAuthCmdIEIExtendedProtCfgOpts uint8 = 0x7B
)

func (m *PDUSessAuthCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessAuthCmd) MsgType() MsgType {
	return MsgTypePDUSessAuthCmd
}

func (m *PDUSessAuthCmd) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessAuthCmd) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessAuthCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessAuthCmd len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	m.EAPMsg = new(ie.EAPMsg) // LV-E, 6-1502B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "PDUSessAuthCmd UnmarshalBinary getIeLen of EAPMsg")
	}
	if err = m.EAPMsg.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "PDUSessAuthCmd.EAPMsg.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessAuthCmdIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessAuthCmd UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// The PDU SESSION AUTHENTICATION COMMAND message is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessAuthCmd.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessAuthCmd unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessAuthCmd) MarshalBinary() ([]byte, error) {
	if m.EAPMsg == nil {
		return nil, errors.Errorf("EAPMsg=%v must present in PDUSessAuthCmd",
			m.EAPMsg)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessAuthCmd),
	})

	// eapmsg, LV-E, 6-1502B
	eapmsg, err := m.EAPMsg.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessAuthCmd.EAPMsg.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(eapmsg))); err != nil {
		return nil, errors.Wrap(err, "PDUSessAuthCmd) MarshalBinary() binary write EAPMsg")
	}
	writer.Write(eapmsg)

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthCmd.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessAuthCmdIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessAuthCmd) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
