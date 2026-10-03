package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessModRej{}

// PDUSessModRej is detailed in 8.3.8 PDU session modification reject, 24.501
type PDUSessModRej struct {
	PDUSessId                        uint8
	PTI                              uint8
	Cause5GSM                        *ie.Cause5GSM                        //     V,       1B, 9.11.4.2
	BackoffTimerValue                *ie.GPRSTimer3                       //   TLV,       3B, 9.11.2.5
	CongestionReattemptIndicator5GSM *ie.CongestionReattemptIndicator5GSM //   TLV,       3B, 9.11.4.21
	ExtendedProtCfgOpts              *ie.ExtendedProtCfgOpts              // TLV-E, 4-65538B, 9.11.4.6
	ReattemptIndicator               *ie.ReattemptIndicator               //   TLV,       3B, 9.11.4.17
}

const (
	PDUSessModRejIEIBackoffTimerValue                uint8 = 0x37
	PDUSessModRejIEICongestionReattemptIndicator5GSM uint8 = 0x61
	PDUSessModRejIEIExtendedProtCfgOpts              uint8 = 0x7B
	PDUSessModRejIEIReattemptIndicator               uint8 = 0x1D
)

func (m *PDUSessModRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessModRej) MsgType() MsgType {
	return MsgTypePDUSessModRej
}

func (m *PDUSessModRej) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessModRej) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessModRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessModRej len(b)=%d, < GsmHdrLen(%d)",
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
		return errors.Wrap(err, "PDUSessModRej.Cause5GSM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessModRejIEIBackoffTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModRej UnmarshalBinary getIeLen of BackoffTimerValue")
			}
			if m.BackoffTimerValue != nil {
				reader.Next(int(ieLen))
				break
			}
			m.BackoffTimerValue = new(ie.GPRSTimer3)
			if err = m.BackoffTimerValue.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.BackoffTimerValue = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModRej.BackoffTimerValue.UnmarshalBinary")
			}
		case PDUSessModRejIEICongestionReattemptIndicator5GSM: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModRej UnmarshalBinary getIeLen of CongestionReattemptIndicator5GSM")
			}
			if m.CongestionReattemptIndicator5GSM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.CongestionReattemptIndicator5GSM = new(ie.CongestionReattemptIndicator5GSM)
			if err = m.CongestionReattemptIndicator5GSM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CongestionReattemptIndicator5GSM = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModRej.CongestionReattemptIndicator5GSM.UnmarshalBinary")
			}
		case PDUSessModRejIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModRej UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION MODIFICATION REJECT is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModRej.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		case PDUSessModRejIEIReattemptIndicator: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModRej UnmarshalBinary getIeLen of ReattemptIndicator")
			}
			if m.ReattemptIndicator != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReattemptIndicator = new(ie.ReattemptIndicator)
			if err = m.ReattemptIndicator.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReattemptIndicator = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModRej.ReattemptIndicator.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessModRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessModRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GSM == nil {
		return nil, errors.Errorf("Cause5GSM=%v must present in PDUSessModRej",
			m.Cause5GSM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessModRej),
	})

	// cause5gsm, V, 1B
	cause5gsm, err := m.Cause5GSM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessModRej.Cause5GSM.MarshalBinary()")
	}
	writer.Write(cause5gsm)

	// m.BackoffTimerValue TLV, 3B, IEI=0x37
	if m.BackoffTimerValue != nil {
		out, err := m.BackoffTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModRej.BackoffTimerValue.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModRejIEIBackoffTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CongestionReattemptIndicator5GSM TLV, 3B, IEI=0x61
	if m.CongestionReattemptIndicator5GSM != nil {
		out, err := m.CongestionReattemptIndicator5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModRej.CongestionReattemptIndicator5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModRejIEICongestionReattemptIndicator5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModRej.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModRejIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModRej) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.ReattemptIndicator TLV, 3B, IEI=0x1D
	if m.ReattemptIndicator != nil {
		out, err := m.ReattemptIndicator.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModRej.ReattemptIndicator.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModRejIEIReattemptIndicator)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
