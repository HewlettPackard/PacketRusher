package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &DeregReqUETerm{}

// DeregReqUETerm is detailed in 8.2.14 De-registration request (UE terminated de-registration), 24.501
type DeregReqUETerm struct {
	DeregType                       *ie.DeregType             //     V,     1/2B, 9.11.3.20
	SpareHalfOctet                  struct{}                  //     V,     1/2B, 9.5
	Cause5GMM                       *ie.Cause5GMM             //    TV,       2B, 9.11.3.2
	T3346Value                      *ie.GPRSTimer2            //   TLV,       3B, 9.11.2.4
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
	DeregReqUETermIEICause5GMM                       uint8 = 0x58
	DeregReqUETermIEIT3346Value                      uint8 = 0x5F
	DeregReqUETermIEIRejectedNSSAI                   uint8 = 0x6D
	DeregReqUETermIEICAGInfoList                     uint8 = 0x75
	DeregReqUETermIEIExtendedRejectedNSSAI           uint8 = 0x68
	DeregReqUETermIEIDisasterReturnWaitRange         uint8 = 0x2C
	DeregReqUETermIEIExtendedCAGInfoList             uint8 = 0x71
	DeregReqUETermIEILowerBoundTimerValue            uint8 = 0x3A
	DeregReqUETermIEIForbiddenTAI_5GSRoaming         uint8 = 0x1D
	DeregReqUETermIEIForbiddenTAI_5GSRegionalProvSvc uint8 = 0x1E
)

func (m *DeregReqUETerm) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *DeregReqUETerm) MsgType() MsgType {
	return MsgTypeDeregReqUETerm
}

func (m *DeregReqUETerm) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("DeregReqUETerm len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.DeregType = new(ie.DeregType) // V, 1/2B
	if err = m.DeregType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "DeregReqUETerm.DeregType.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case DeregReqUETermIEICause5GMM: // TV, 2B
			ieLen = 1
			if m.Cause5GMM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Cause5GMM = new(ie.Cause5GMM)
			if err = m.Cause5GMM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Cause5GMM = nil
					continue
				}
				return errors.Wrap(err, "DeregReqUETerm.Cause5GMM.UnmarshalBinary")
			}
		case DeregReqUETermIEIT3346Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of T3346Value")
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
				return errors.Wrap(err, "DeregReqUETerm.T3346Value.UnmarshalBinary")
			}
		case DeregReqUETermIEIRejectedNSSAI: // TLV, 4-42B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of RejectedNSSAI")
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
				return errors.Wrap(err, "DeregReqUETerm.RejectedNSSAI.UnmarshalBinary")
			}
		case DeregReqUETermIEICAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of CAGInfoList")
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
				return errors.Wrap(err, "DeregReqUETerm.CAGInfoList.UnmarshalBinary")
			}
		case DeregReqUETermIEIExtendedRejectedNSSAI: // TLV, 5-90B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of ExtendedRejectedNSSAI")
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
				return errors.Wrap(err, "DeregReqUETerm.ExtendedRejectedNSSAI.UnmarshalBinary")
			}
		case DeregReqUETermIEIDisasterReturnWaitRange: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of DisasterReturnWaitRange")
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
				return errors.Wrap(err, "DeregReqUETerm.DisasterReturnWaitRange.UnmarshalBinary")
			}
		case DeregReqUETermIEIExtendedCAGInfoList: // TLV-E, 3-nB
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of ExtendedCAGInfoList")
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
				return errors.Wrap(err, "DeregReqUETerm.ExtendedCAGInfoList.UnmarshalBinary")
			}
		case DeregReqUETermIEILowerBoundTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of LowerBoundTimerValue")
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
				return errors.Wrap(err, "DeregReqUETerm.LowerBoundTimerValue.UnmarshalBinary")
			}
		case DeregReqUETermIEIForbiddenTAI_5GSRoaming: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of ForbiddenTAI_5GSRoaming")
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
				return errors.Wrap(err, "DeregReqUETerm.ForbiddenTAI_5GSRoaming.UnmarshalBinary")
			}
		case DeregReqUETermIEIForbiddenTAI_5GSRegionalProvSvc: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DeregReqUETerm UnmarshalBinary getIeLen of ForbiddenTAI_5GSRegionalProvSvc")
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
				return errors.Wrap(err, "DeregReqUETerm.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("DeregReqUETerm unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *DeregReqUETerm) MarshalBinary() ([]byte, error) {
	if m.DeregType == nil {
		return nil, errors.Errorf("DeregType=%v SpareHalfOctet=%v must present in DeregReqUETerm",
			m.DeregType, m.SpareHalfOctet)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeDeregReqUETerm),
	})

	tmp := [1]byte{}
	// deregtype, V, 1/2B
	deregtype, err := m.DeregType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DeregReqUETerm.DeregType.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, deregtype[0])
	writer.Write(tmp[:])

	// m.Cause5GMM TV, 2B, IEI=0x58
	if m.Cause5GMM != nil {
		out, err := m.Cause5GMM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.Cause5GMM.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEICause5GMM)
		writer.Write(out)
	}

	// m.T3346Value TLV, 3B, IEI=0x5F
	if m.T3346Value != nil {
		out, err := m.T3346Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.T3346Value.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIT3346Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RejectedNSSAI TLV, 4-42B, IEI=0x6D
	if m.RejectedNSSAI != nil {
		out, err := m.RejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.RejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.CAGInfoList TLV-E, 3-nB, IEI=0x75
	if m.CAGInfoList != nil {
		out, err := m.CAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.CAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEICAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm) MarshalBinary() binary write CAGInfoList")
		}
		writer.Write(out)
	}

	// m.ExtendedRejectedNSSAI TLV, 5-90B, IEI=0x68
	if m.ExtendedRejectedNSSAI != nil {
		out, err := m.ExtendedRejectedNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.ExtendedRejectedNSSAI.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIExtendedRejectedNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DisasterReturnWaitRange TLV, 4B, IEI=0x2C
	if m.DisasterReturnWaitRange != nil {
		out, err := m.DisasterReturnWaitRange.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.DisasterReturnWaitRange.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIDisasterReturnWaitRange)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedCAGInfoList TLV-E, 3-nB, IEI=0x71
	if m.ExtendedCAGInfoList != nil {
		out, err := m.ExtendedCAGInfoList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.ExtendedCAGInfoList.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIExtendedCAGInfoList)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm) MarshalBinary() binary write ExtendedCAGInfoList")
		}
		writer.Write(out)
	}

	// m.LowerBoundTimerValue TLV, 3B, IEI=0x3A
	if m.LowerBoundTimerValue != nil {
		out, err := m.LowerBoundTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.LowerBoundTimerValue.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEILowerBoundTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRoaming TLV, 9-114B, IEI=0x1D
	if m.ForbiddenTAI_5GSRoaming != nil {
		out, err := m.ForbiddenTAI_5GSRoaming.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.ForbiddenTAI_5GSRoaming.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIForbiddenTAI_5GSRoaming)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRegionalProvSvc TLV, 9-114B, IEI=0x1E
	if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
		out, err := m.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DeregReqUETerm.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()")
		}
		writer.WriteByte(DeregReqUETermIEIForbiddenTAI_5GSRegionalProvSvc)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
