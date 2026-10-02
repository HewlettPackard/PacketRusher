package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessModComplete{}

// PDUSessModComplete is detailed in 8.3.10 PDU session modification complete, 24.501
type PDUSessModComplete struct {
	PDUSessId           uint8
	PTI                 uint8
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
	PortMgmtInfoCntr    *ie.PortMgmtInfoCntr    // TLV-E, 4-65538B, 9.11.4.27
}

const (
	PDUSessModCompleteIEIExtendedProtCfgOpts uint8 = 0x7B
	PDUSessModCompleteIEIPortMgmtInfoCntr    uint8 = 0x74
)

func (m *PDUSessModComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessModComplete) MsgType() MsgType {
	return MsgTypePDUSessModComplete
}

func (m *PDUSessModComplete) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessModComplete) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessModComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessModComplete len(b)=%d, < GsmHdrLen(%d)",
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
		case PDUSessModCompleteIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModComplete UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION MODIFICATION COMPLETE is sent by the UE to the SMF
			if err = m.ExtendedProtCfgOpts.UnmarshalFromMs(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModComplete.ExtendedProtCfgOpts.UnmarshalFromMs")
			}
		case PDUSessModCompleteIEIPortMgmtInfoCntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModComplete UnmarshalBinary getIeLen of PortMgmtInfoCntr")
			}
			if m.PortMgmtInfoCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PortMgmtInfoCntr = new(ie.PortMgmtInfoCntr)
			if err = m.PortMgmtInfoCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PortMgmtInfoCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModComplete.PortMgmtInfoCntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessModComplete unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessModComplete) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessModComplete),
	})

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModComplete.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCompleteIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModComplete) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.PortMgmtInfoCntr TLV-E, 4-65538B, IEI=0x74
	if m.PortMgmtInfoCntr != nil {
		out, err := m.PortMgmtInfoCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModComplete.PortMgmtInfoCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCompleteIEIPortMgmtInfoCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModComplete) MarshalBinary() binary write PortMgmtInfoCntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
