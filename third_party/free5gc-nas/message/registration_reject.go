package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RegRej{}

// RegRej is detailed in 8.2.9 Registration reject, 24.501
type RegRej struct {
	Cause5GMM                       *ie.Cause5GMM             //     V,       1B, 9.11.3.2
	T3346Value                      *ie.GPRSTimer2            //   TLV,       3B, 9.11.2.4
	T3502Value                      *ie.GPRSTimer2            //   TLV,       3B, 9.11.2.4
	EAPMsg                          *ie.EAPMsg                // TLV-E,  7-1503B, 9.11.2.2
	RejectedNSSAI                   *ie.RejectedNSSAI         //   TLV,    4-42B, 9.11.3.46
	CAGInfoList                     *ie.CAGInfoList           // TLV-E,     3-nB, 9.11.3.18A
	ExtendedRejectedNSSAI           *ie.ExtendedRejectedNSSAI //   TLV,    5-90B, 9.11.3.75
	DisasterReturnWaitRange         *ie.RegWaitRange          //   TLV,       4B, 9.11.3.84
	ExtendedCAGInfoList             *ie.ExtendedCAGInfoList   // TLV-E,     3-nB, 9.11.3.86
	LowerBoundTimerValue            *ie.GPRSTimer3            //   TLV,       3B, 9.11.2.5
	ForbiddenTAI_5GSRoaming         *ie.TrackingAreaIdList5GS //   TLV,   9-114B, 9.11.3.9
	ForbiddenTAI_5GSRegionalProvSvc *ie.TrackingAreaIdList5GS //   TLV,   9-114B, 9.11.3.9
}

const (
	RegRejIEIT3346Value                      uint8 = 0x5F
	RegRejIEIT3502Value                      uint8 = 0x16
	RegRejIEIEAPMsg                          uint8 = 0x78
	RegRejIEIRejectedNSSAI                   uint8 = 0x69
	RegRejIEICAGInfoList                     uint8 = 0x75
	RegRejIEIExtendedRejectedNSSAI           uint8 = 0x68
	RegRejIEIDisasterReturnWaitRange         uint8 = 0x2C
	RegRejIEIExtendedCAGInfoList             uint8 = 0x71
	RegRejIEILowerBoundTimerValue            uint8 = 0x3A
	RegRejIEIForbiddenTAI_5GSRoaming         uint8 = 0x1D
	RegRejIEIForbiddenTAI_5GSRegionalProvSvc uint8 = 0x1E
)

func (m *RegRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RegRej) MsgType() MsgType {
	return MsgTypeRegRej
}

func (m *RegRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RegRej len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.Cause5GMM = new(ie.Cause5GMM) // V, 1B
	if err = m.Cause5GMM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "RegRej.Cause5GMM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case RegRejIEIT3346Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of T3346Value")
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
				return errors.Wrap(err, "RegRej.T3346Value.UnmarshalBinary")
			}
		case RegRejIEIT3502Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of T3502Value")
			}
			if m.T3502Value != nil {
				reader.Next(int(ieLen))
				break
			}
			m.T3502Value = new(ie.GPRSTimer2)
			if err = m.T3502Value.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.T3502Value = nil
					continue
				}
				return errors.Wrap(err, "RegRej.T3502Value.UnmarshalBinary")
			}
		case RegRejIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "RegRej.EAPMsg.UnmarshalBinary")
			}
		case RegRejIEIRejectedNSSAI: // TLV, 4-42B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of RejectedNSSAI")
			}
			if m.RejectedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RejectedNSSAI = new(ie.RejectedNSSAI)
			if err = m.RejectedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RejectedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "RegRej.RejectedNSSAI.UnmarshalBinary")
			}
		case RegRejIEICAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of CAGInfoList")
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
				return errors.Wrap(err, "RegRej.CAGInfoList.UnmarshalBinary")
			}
		case RegRejIEIExtendedRejectedNSSAI: // TLV, 5-90B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of ExtendedRejectedNSSAI")
			}
			if m.ExtendedRejectedNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedRejectedNSSAI = new(ie.ExtendedRejectedNSSAI)
			if err = m.ExtendedRejectedNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedRejectedNSSAI = nil
					continue
				}
				return errors.Wrap(err, "RegRej.ExtendedRejectedNSSAI.UnmarshalBinary")
			}
		case RegRejIEIDisasterReturnWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of DisasterReturnWaitRange")
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
				return errors.Wrap(err, "RegRej.DisasterReturnWaitRange.UnmarshalBinary")
			}
		case RegRejIEIExtendedCAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of ExtendedCAGInfoList")
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
				return errors.Wrap(err, "RegRej.ExtendedCAGInfoList.UnmarshalBinary")
			}
		case RegRejIEILowerBoundTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of LowerBoundTimerValue")
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
				return errors.Wrap(err, "RegRej.LowerBoundTimerValue.UnmarshalBinary")
			}
		case RegRejIEIForbiddenTAI_5GSRoaming: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of ForbiddenTAI_5GSRoaming")
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
				return errors.Wrap(err, "RegRej.ForbiddenTAI_5GSRoaming.UnmarshalBinary")
			}
		case RegRejIEIForbiddenTAI_5GSRegionalProvSvc: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "RegRej UnmarshalBinary getIeLen of ForbiddenTAI_5GSRegionalProvSvc")
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
				return errors.Wrap(err, "RegRej.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RegRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RegRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GMM == nil {
		return nil, errors.Errorf("Cause5GMM=%v must present in RegRej",
			m.Cause5GMM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRegRej),
	})

	// cause5gmm, V, 1B
	cause5gmm, err := m.Cause5GMM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RegRej.Cause5GMM.MarshalBinary()")
	}
	writer.Write(cause5gmm)

	// m.T3346Value TLV, 3B, IEI=0x5F
	if m.T3346Value != nil {
		out, err := m.T3346Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.T3346Value.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIT3346Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.T3502Value TLV, 3B, IEI=0x16
	if m.T3502Value != nil {
		out, err := m.T3502Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.T3502Value.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIT3502Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegRej) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.RejectedNSSAI TLV, 4-42B, IEI=0x69
	if m.RejectedNSSAI != nil {
		out, err := m.RejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.RejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CAGInfoList TLV-E, 3-nB, IEI=0x75
	if m.CAGInfoList != nil {
		out, err := m.CAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.CAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEICAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegRej) MarshalBinary() binary write CAGInfoList")
		}
		writer.Write(out)
	}

	// m.ExtendedRejectedNSSAI TLV, 5-90B, IEI=0x68
	if m.ExtendedRejectedNSSAI != nil {
		out, err := m.ExtendedRejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.ExtendedRejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIExtendedRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterReturnWaitRange TLV, 4B, IEI=0x2C
	if m.DisasterReturnWaitRange != nil {
		out, err := m.DisasterReturnWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.DisasterReturnWaitRange.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIDisasterReturnWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedCAGInfoList TLV-E, 3-nB, IEI=0x71
	if m.ExtendedCAGInfoList != nil {
		out, err := m.ExtendedCAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.ExtendedCAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIExtendedCAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RegRej) MarshalBinary() binary write ExtendedCAGInfoList")
		}
		writer.Write(out)
	}

	// m.LowerBoundTimerValue TLV, 3B, IEI=0x3A
	if m.LowerBoundTimerValue != nil {
		out, err := m.LowerBoundTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.LowerBoundTimerValue.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEILowerBoundTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRoaming TLV, 9-114B, IEI=0x1D
	if m.ForbiddenTAI_5GSRoaming != nil {
		out, err := m.ForbiddenTAI_5GSRoaming.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.ForbiddenTAI_5GSRoaming.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIForbiddenTAI_5GSRoaming)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRegionalProvSvc TLV, 9-114B, IEI=0x1E
	if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
		out, err := m.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RegRej.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()")
		}
		writer.WriteByte(RegRejIEIForbiddenTAI_5GSRegionalProvSvc)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
