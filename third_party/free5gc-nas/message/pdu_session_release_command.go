package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessRelCmd{}

// PDUSessRelCmd is detailed in 8.3.14 PDU session release command, 24.501
type PDUSessRelCmd struct {
	PDUSessId                        uint8
	PTI                              uint8
	Cause5GSM                        *ie.Cause5GSM                        //     V,       1B, 9.11.4.2
	BackoffTimerValue                *ie.GPRSTimer3                       //   TLV,       3B, 9.11.2.5
	EAPMsg                           *ie.EAPMsg                           // TLV-E,  7-1503B, 9.11.2.2
	CongestionReattemptIndicator5GSM *ie.CongestionReattemptIndicator5GSM //   TLV,       3B, 9.11.4.21
	ExtendedProtCfgOpts              *ie.ExtendedProtCfgOpts              // TLV-E, 4-65538B, 9.11.4.6
	AccessType                       *ie.AccessType                       //    TV,       1B, 9.11.2.1A
	SvcLvlAACntr                     *ie.SvcLvlAACntr                     // TLV-E, 4-65538B, 9.11.2.10
}

const (
	PDUSessRelCmdIEIBackoffTimerValue                uint8 = 0x37
	PDUSessRelCmdIEIEAPMsg                           uint8 = 0x78
	PDUSessRelCmdIEICongestionReattemptIndicator5GSM uint8 = 0x61
	PDUSessRelCmdIEIExtendedProtCfgOpts              uint8 = 0x7B
	PDUSessRelCmdIEIAccessType                       uint8 = 0xD0
	PDUSessRelCmdIEISvcLvlAACntr                     uint8 = 0x72
)

func (m *PDUSessRelCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessRelCmd) MsgType() MsgType {
	return MsgTypePDUSessRelCmd
}

func (m *PDUSessRelCmd) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessRelCmd) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessRelCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessRelCmd len(b)=%d, < GsmHdrLen(%d)",
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
		return errors.Wrap(err, "PDUSessRelCmd.Cause5GSM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessRelCmdIEIBackoffTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelCmd UnmarshalBinary getIeLen of BackoffTimerValue")
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
				return errors.Wrap(err, "PDUSessRelCmd.BackoffTimerValue.UnmarshalBinary")
			}
		case PDUSessRelCmdIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelCmd UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "PDUSessRelCmd.EAPMsg.UnmarshalBinary")
			}
		case PDUSessRelCmdIEICongestionReattemptIndicator5GSM: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelCmd UnmarshalBinary getIeLen of CongestionReattemptIndicator5GSM")
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
				return errors.Wrap(err, "PDUSessRelCmd.CongestionReattemptIndicator5GSM.UnmarshalBinary")
			}
		case PDUSessRelCmdIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelCmd UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			// PDU SESSION RELEASE COMMAND is sent by the SMF to the UE
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelCmd.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		case PDUSessRelCmdIEIAccessType: // TV, 1B
			if m.AccessType != nil {
				break
			}
			m.AccessType = new(ie.AccessType)
			if err = m.AccessType.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AccessType = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelCmd.AccessType.UnmarshalBinary")
			}
		case PDUSessRelCmdIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelCmd UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "PDUSessRelCmd.SvcLvlAACntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessRelCmd unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessRelCmd) MarshalBinary() ([]byte, error) {
	if m.Cause5GSM == nil {
		return nil, errors.Errorf("Cause5GSM=%v must present in PDUSessRelCmd",
			m.Cause5GSM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessRelCmd),
	})

	// cause5gsm, V, 1B
	cause5gsm, err := m.Cause5GSM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessRelCmd.Cause5GSM.MarshalBinary()")
	}
	writer.Write(cause5gsm)

	// m.BackoffTimerValue TLV, 3B, IEI=0x37
	if m.BackoffTimerValue != nil {
		out, err := m.BackoffTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.BackoffTimerValue.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCmdIEIBackoffTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCmdIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.CongestionReattemptIndicator5GSM TLV, 3B, IEI=0x61
	if m.CongestionReattemptIndicator5GSM != nil {
		out, err := m.CongestionReattemptIndicator5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.CongestionReattemptIndicator5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCmdIEICongestionReattemptIndicator5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCmdIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.AccessType TV, 1B, IEI=0xD0, >= 0x80 !
	if m.AccessType != nil {
		out, err := m.AccessType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.AccessType.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessRelCmdIEIAccessType)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelCmdIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelCmd) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
