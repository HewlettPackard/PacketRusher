package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SvcRej{}

// SvcRej is detailed in 8.2.18 Service reject, 24.501
type SvcRej struct {
	Cause5GMM                       *ie.Cause5GMM             //     V,       1B, 9.11.3.2
	PDUSessStatus                   *ie.PDUSessStatus         //   TLV,    4-34B, 9.11.3.44
	T3346Value                      *ie.GPRSTimer2            //   TLV,       3B, 9.11.2.4
	EAPMsg                          *ie.EAPMsg                // TLV-E,  7-1503B, 9.11.2.2
	T3448Value                      *ie.GPRSTimer2            //   TLV,       3B, 9.11.2.4
	CAGInfoList                     *ie.CAGInfoList           // TLV-E,     3-nB, 9.11.3.18A
	DisasterReturnWaitRange         *ie.RegWaitRange          //   TLV,       4B, 9.11.3.84
	ExtendedCAGInfoList             *ie.ExtendedCAGInfoList   // TLV-E,     3-nB, 9.11.3.86
	LowerBoundTimerValue            *ie.GPRSTimer3            //   TLV,       3B, 9.11.2.5
	ForbiddenTAI_5GSRoaming         *ie.TrackingAreaIdList5GS //   TLV,   9-114B, 9.11.3.9
	ForbiddenTAI_5GSRegionalProvSvc *ie.TrackingAreaIdList5GS //   TLV,   9-114B, 9.11.3.9
}

const (
	SvcRejIEIPDUSessStatus                   uint8 = 0x50
	SvcRejIEIT3346Value                      uint8 = 0x5F
	SvcRejIEIEAPMsg                          uint8 = 0x78
	SvcRejIEIT3448Value                      uint8 = 0x6B
	SvcRejIEICAGInfoList                     uint8 = 0x75
	SvcRejIEIDisasterReturnWaitRange         uint8 = 0x2C
	SvcRejIEIExtendedCAGInfoList             uint8 = 0x71
	SvcRejIEILowerBoundTimerValue            uint8 = 0x3A
	SvcRejIEIForbiddenTAI_5GSRoaming         uint8 = 0x1D
	SvcRejIEIForbiddenTAI_5GSRegionalProvSvc uint8 = 0x1E
)

func (m *SvcRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SvcRej) MsgType() MsgType {
	return MsgTypeSvcRej
}

func (m *SvcRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SvcRej len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.Cause5GMM = new(ie.Cause5GMM) // V, 1B
	if err = m.Cause5GMM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "SvcRej.Cause5GMM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case SvcRejIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of PDUSessStatus")
			}
			if m.PDUSessStatus != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessStatus = new(ie.PDUSessStatus)
			if err = m.PDUSessStatus.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessStatus = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.PDUSessStatus.UnmarshalBinary")
			}
		case SvcRejIEIT3346Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of T3346Value")
			}
			if m.T3346Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3346Value = new(ie.GPRSTimer2)
			if err = m.T3346Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3346Value = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.T3346Value.UnmarshalBinary")
			}
		case SvcRejIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "SvcRej.EAPMsg.UnmarshalBinary")
			}
		case SvcRejIEIT3448Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of T3448Value")
			}
			if m.T3448Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3448Value = new(ie.GPRSTimer2)
			if err = m.T3448Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3448Value = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.T3448Value.UnmarshalBinary")
			}
		case SvcRejIEICAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of CAGInfoList")
			}
			if m.CAGInfoList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.CAGInfoList = new(ie.CAGInfoList)
			if err = m.CAGInfoList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CAGInfoList = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.CAGInfoList.UnmarshalBinary")
			}
		case SvcRejIEIDisasterReturnWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of DisasterReturnWaitRange")
			}
			if m.DisasterReturnWaitRange != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DisasterReturnWaitRange = new(ie.RegWaitRange)
			if err = m.DisasterReturnWaitRange.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DisasterReturnWaitRange = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.DisasterReturnWaitRange.UnmarshalBinary")
			}
		case SvcRejIEIExtendedCAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of ExtendedCAGInfoList")
			}
			if m.ExtendedCAGInfoList != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedCAGInfoList = new(ie.ExtendedCAGInfoList)
			if err = m.ExtendedCAGInfoList.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedCAGInfoList = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.ExtendedCAGInfoList.UnmarshalBinary")
			}
		case SvcRejIEILowerBoundTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of LowerBoundTimerValue")
			}
			if m.LowerBoundTimerValue != nil {
				reader.Next(int(ieLen))
				break
			}
			m.LowerBoundTimerValue = new(ie.GPRSTimer3)
			if err = m.LowerBoundTimerValue.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.LowerBoundTimerValue = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.LowerBoundTimerValue.UnmarshalBinary")
			}
		case SvcRejIEIForbiddenTAI_5GSRoaming: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of ForbiddenTAI_5GSRoaming")
			}
			if m.ForbiddenTAI_5GSRoaming != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ForbiddenTAI_5GSRoaming = new(ie.TrackingAreaIdList5GS)
			if err = m.ForbiddenTAI_5GSRoaming.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ForbiddenTAI_5GSRoaming = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.ForbiddenTAI_5GSRoaming.UnmarshalBinary")
			}
		case SvcRejIEIForbiddenTAI_5GSRegionalProvSvc: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcRej UnmarshalBinary getIeLen of ForbiddenTAI_5GSRegionalProvSvc")
			}
			if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ForbiddenTAI_5GSRegionalProvSvc = new(ie.TrackingAreaIdList5GS)
			if err = m.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ForbiddenTAI_5GSRegionalProvSvc = nil
					continue
				}
				return errors.Wrap(err, "SvcRej.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("SvcRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *SvcRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GMM == nil {
		return nil, errors.Errorf("Cause5GMM=%v must present in SvcRej",
			m.Cause5GMM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSvcRej),
	})

	// cause5gmm, V, 1B
	cause5gmm, err := m.Cause5GMM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcRej.Cause5GMM.MarshalBinary()")
	}
	writer.Write(cause5gmm)

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3346Value TLV, 3B, IEI=0x5F
	if m.T3346Value != nil {
		out, err := m.T3346Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.T3346Value.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIT3346Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcRej) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.T3448Value TLV, 3B, IEI=0x6B
	if m.T3448Value != nil {
		out, err := m.T3448Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.T3448Value.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIT3448Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CAGInfoList TLV-E, 3-nB, IEI=0x75
	if m.CAGInfoList != nil {
		out, err := m.CAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.CAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEICAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcRej) MarshalBinary() binary write CAGInfoList")
		}
		writer.Write(out)
	}

	// m.DisasterReturnWaitRange TLV, 4B, IEI=0x2C
	if m.DisasterReturnWaitRange != nil {
		out, err := m.DisasterReturnWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.DisasterReturnWaitRange.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIDisasterReturnWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedCAGInfoList TLV-E, 3-nB, IEI=0x71
	if m.ExtendedCAGInfoList != nil {
		out, err := m.ExtendedCAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.ExtendedCAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIExtendedCAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcRej) MarshalBinary() binary write ExtendedCAGInfoList")
		}
		writer.Write(out)
	}

	// m.LowerBoundTimerValue TLV, 3B, IEI=0x3A
	if m.LowerBoundTimerValue != nil {
		out, err := m.LowerBoundTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.LowerBoundTimerValue.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEILowerBoundTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRoaming TLV, 9-114B, IEI=0x1D
	if m.ForbiddenTAI_5GSRoaming != nil {
		out, err := m.ForbiddenTAI_5GSRoaming.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.ForbiddenTAI_5GSRoaming.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIForbiddenTAI_5GSRoaming)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRegionalProvSvc TLV, 9-114B, IEI=0x1E
	if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
		out, err := m.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcRej.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()")
		}
		writer.WriteByte(SvcRejIEIForbiddenTAI_5GSRegionalProvSvc)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
