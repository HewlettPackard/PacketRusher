package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessRelComplete{}

// PDUSessRelComplete is detailed in 8.3.15 PDU session release complete, 24.501
type PDUSessRelComplete struct {
	PDUSessId           uint8
	PTI                 uint8
	Cause5GSM           *ie.Cause5GSM           //    TV,       2B, 9.11.4.2
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
}

const (
	PDUSessRelCompleteIEICause5GSM           uint8 = 0x59
	PDUSessRelCompleteIEIExtendedProtCfgOpts uint8 = 0x7B
)

func (m *PDUSessRelComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessRelComplete) MsgType() MsgType {
	return MsgTypePDUSessRelComplete
}

func (m *PDUSessRelComplete) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessRelComplete) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessRelComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessRelComplete len(b)=%d, < GsmHdrLen(%d)",
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
		case PDUSessRelCompleteIEICause5GSM: // TV, 2B
			ieLen = 1
			if m.Cause5GSM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Cause5GSM = new(ie.Cause5GSM)
			if err = m.Cause5GSM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Cause5GSM = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelComplete.Cause5GSM.UnmarshalBinary")
			}
		case PDUSessRelCompleteIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelComplete UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION RELEASE COMPLETE is sent by the UE to the SMF
			if err = m.ExtendedProtCfgOpts.UnmarshalFromMs(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelComplete.ExtendedProtCfgOpts.UnmarshalFromMs")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessRelComplete unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessRelComplete) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessRelComplete),
	})

	// m.Cause5GSM TV, 2B, IEI=0x59
	if m.Cause5GSM != nil {
		out, err := m.Cause5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelComplete.Cause5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCompleteIEICause5GSM)
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelComplete.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCompleteIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelComplete) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
