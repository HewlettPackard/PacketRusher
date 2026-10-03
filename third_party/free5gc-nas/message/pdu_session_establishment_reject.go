package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessEstRej{}

// PDUSessEstRej is detailed in 8.3.3 PDU session establishment reject, 24.501
type PDUSessEstRej struct {
	PDUSessId                        uint8
	PTI                              uint8
	Cause5GSM                        *ie.Cause5GSM                        //     V,       1B, 9.11.4.2
	BackoffTimerValue                *ie.GPRSTimer3                       //   TLV,       3B, 9.11.2.5
	AllowedSSCMode                   *ie.AllowedSSCMode                   //    TV,       1B, 9.11.4.5
	EAPMsg                           *ie.EAPMsg                           // TLV-E,  7-1503B, 9.11.2.2
	CongestionReattemptIndicator5GSM *ie.CongestionReattemptIndicator5GSM //   TLV,       3B, 9.11.4.21
	ExtendedProtCfgOpts              *ie.ExtendedProtCfgOpts              // TLV-E, 4-65538B, 9.11.4.6
	ReattemptIndicator               *ie.ReattemptIndicator               //   TLV,       3B, 9.11.4.17
	SvcLvlAACntr                     *ie.SvcLvlAACntr                     // TLV-E, 4-65538B, 9.11.2.10
}

const (
	PDUSessEstRejIEIBackoffTimerValue                uint8 = 0x37
	PDUSessEstRejIEIAllowedSSCMode                   uint8 = 0xF0
	PDUSessEstRejIEIEAPMsg                           uint8 = 0x78
	PDUSessEstRejIEICongestionReattemptIndicator5GSM uint8 = 0x61
	PDUSessEstRejIEIExtendedProtCfgOpts              uint8 = 0x7B
	PDUSessEstRejIEIReattemptIndicator               uint8 = 0x1D
	PDUSessEstRejIEISvcLvlAACntr                     uint8 = 0x72
)

func (m *PDUSessEstRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessEstRej) MsgType() MsgType {
	return MsgTypePDUSessEstRej
}

func (m *PDUSessEstRej) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessEstRej) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessEstRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessEstRej len(b)=%d, < GsmHdrLen(%d)",
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
		return errors.Wrap(err, "PDUSessEstRej.Cause5GSM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessEstRejIEIBackoffTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of BackoffTimerValue")
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
				return errors.Wrap(err, "PDUSessEstRej.BackoffTimerValue.UnmarshalBinary")
			}
		case PDUSessEstRejIEIAllowedSSCMode: // TV, 1B
			if m.AllowedSSCMode != nil {
				break
			}
			m.AllowedSSCMode = new(ie.AllowedSSCMode)
			if err = m.AllowedSSCMode.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AllowedSSCMode = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstRej.AllowedSSCMode.UnmarshalBinary")
			}
		case PDUSessEstRejIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "PDUSessEstRej.EAPMsg.UnmarshalBinary")
			}
		case PDUSessEstRejIEICongestionReattemptIndicator5GSM: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of CongestionReattemptIndicator5GSM")
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
				return errors.Wrap(err, "PDUSessEstRej.CongestionReattemptIndicator5GSM.UnmarshalBinary")
			}
		case PDUSessEstRejIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION ESTABLISHMENT REJECT is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstRej.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		case PDUSessEstRejIEIReattemptIndicator: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of ReattemptIndicator")
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
				return errors.Wrap(err, "PDUSessEstRej.ReattemptIndicator.UnmarshalBinary")
			}
		case PDUSessEstRejIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstRej UnmarshalBinary getIeLen of SvcLvlAACntr")
			}
			if m.SvcLvlAACntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SvcLvlAACntr = new(ie.SvcLvlAACntr)
			if err = m.SvcLvlAACntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SvcLvlAACntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstRej.SvcLvlAACntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessEstRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessEstRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GSM == nil {
		return nil, errors.Errorf("Cause5GSM=%v must present in PDUSessEstRej",
			m.Cause5GSM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessEstRej),
	})

	// cause5gsm, V, 1B
	cause5gsm, err := m.Cause5GSM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstRej.Cause5GSM.MarshalBinary()")
	}
	writer.Write(cause5gsm)

	// m.BackoffTimerValue TLV, 3B, IEI=0x37
	if m.BackoffTimerValue != nil {
		out, err := m.BackoffTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.BackoffTimerValue.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEIBackoffTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AllowedSSCMode TV, 1B, IEI=0xF0, >= 0x80 !
	if m.AllowedSSCMode != nil {
		out, err := m.AllowedSSCMode.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.AllowedSSCMode.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstRejIEIAllowedSSCMode)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.CongestionReattemptIndicator5GSM TLV, 3B, IEI=0x61
	if m.CongestionReattemptIndicator5GSM != nil {
		out, err := m.CongestionReattemptIndicator5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.CongestionReattemptIndicator5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEICongestionReattemptIndicator5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.ReattemptIndicator TLV, 3B, IEI=0x1D
	if m.ReattemptIndicator != nil {
		out, err := m.ReattemptIndicator.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.ReattemptIndicator.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEIReattemptIndicator)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstRejIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstRej) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
