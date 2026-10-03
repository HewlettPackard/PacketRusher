package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessRelRej{}

// PDUSessRelRej is detailed in 8.3.13 PDU session release reject, 24.501
type PDUSessRelRej struct {
	PDUSessId           uint8
	PTI                 uint8
	Cause5GSM           *ie.Cause5GSM           //     V,       1B, 9.11.4.2
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
}

const (
	PDUSessRelRejIEIExtendedProtCfgOpts uint8 = 0x7B
)

func (m *PDUSessRelRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessRelRej) MsgType() MsgType {
	return MsgTypePDUSessRelRej
}

func (m *PDUSessRelRej) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessRelRej) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessRelRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessRelRej len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	m.Cause5GSM = new(ie.Cause5GSM) // V, 1B
	if err = m.Cause5GSM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "PDUSessRelRej.Cause5GSM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessRelRejIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelRej UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION RELEASE REJECT is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelRej.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessRelRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessRelRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GSM == nil {
		return nil, errors.Errorf("Cause5GSM=%v must present in PDUSessRelRej",
			m.Cause5GSM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessRelRej),
	})

	// cause5gsm, V, 1B
	cause5gsm, err := m.Cause5GSM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessRelRej.Cause5GSM.MarshalBinary()")
	}
	writer.Write(cause5gsm)

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelRej.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelRejIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelRej) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
