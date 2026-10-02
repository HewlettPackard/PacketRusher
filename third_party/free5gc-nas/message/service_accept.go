package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SvcAccept{}

// SvcAccept is detailed in 8.2.17 Service accept, 24.501
type SvcAccept struct {
	PDUSessStatus                     *ie.PDUSessStatus                     //   TLV,    4-34B, 9.11.3.44
	PDUSessReactivationResult         *ie.PDUSessReactivationResult         //   TLV,    4-34B, 9.11.3.42
	PDUSessReactivationResultErrCause *ie.PDUSessReactivationResultErrCause // TLV-E,   5-515B, 9.11.3.43
	EAPMsg                            *ie.EAPMsg                            // TLV-E,  7-1503B, 9.11.2.2
	T3448Value                        *ie.GPRSTimer2                        //   TLV,       3B, 9.11.2.4
	AdditionalReqResult5GS            *ie.AdditionalReqResult5GS            //   TLV,       3B, 9.11.3.81
	ForbiddenTAI_5GSRoaming           *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
	ForbiddenTAI_5GSRegionalProvSvc   *ie.TrackingAreaIdList5GS             //   TLV,   9-114B, 9.11.3.9
}

const (
	SvcAcceptIEIPDUSessStatus                     uint8 = 0x50
	SvcAcceptIEIPDUSessReactivationResult         uint8 = 0x26
	SvcAcceptIEIPDUSessReactivationResultErrCause uint8 = 0x72
	SvcAcceptIEIEAPMsg                            uint8 = 0x78
	SvcAcceptIEIT3448Value                        uint8 = 0x6B
	SvcAcceptIEIAdditionalReqResult5GS            uint8 = 0x34
	SvcAcceptIEIForbiddenTAI_5GSRoaming           uint8 = 0x1D
	SvcAcceptIEIForbiddenTAI_5GSRegionalProvSvc   uint8 = 0x1E
)

func (m *SvcAccept) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SvcAccept) MsgType() MsgType {
	return MsgTypeSvcAccept
}

func (m *SvcAccept) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SvcAccept len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// This message contains 0 Mandatory IE
	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case SvcAcceptIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of PDUSessStatus")
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
				return errors.Wrap(err, "SvcAccept.PDUSessStatus.UnmarshalBinary")
			}
		case SvcAcceptIEIPDUSessReactivationResult: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of PDUSessReactivationResult")
			}
			if m.PDUSessReactivationResult != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessReactivationResult = new(ie.PDUSessReactivationResult)
			if err = m.PDUSessReactivationResult.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessReactivationResult = nil
					continue
				}
				return errors.Wrap(err, "SvcAccept.PDUSessReactivationResult.UnmarshalBinary")
			}
		case SvcAcceptIEIPDUSessReactivationResultErrCause: // TLV-E, 5-515B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of PDUSessReactivationResultErrCause")
			}
			if m.PDUSessReactivationResultErrCause != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessReactivationResultErrCause = new(ie.PDUSessReactivationResultErrCause)
			if err = m.PDUSessReactivationResultErrCause.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessReactivationResultErrCause = nil
					continue
				}
				return errors.Wrap(err, "SvcAccept.PDUSessReactivationResultErrCause.UnmarshalBinary")
			}
		case SvcAcceptIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "SvcAccept.EAPMsg.UnmarshalBinary")
			}
		case SvcAcceptIEIT3448Value: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of T3448Value")
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
				return errors.Wrap(err, "SvcAccept.T3448Value.UnmarshalBinary")
			}
		case SvcAcceptIEIAdditionalReqResult5GS: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of AdditionalReqResult5GS")
			}
			if m.AdditionalReqResult5GS != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalReqResult5GS = new(ie.AdditionalReqResult5GS)
			if err = m.AdditionalReqResult5GS.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalReqResult5GS = nil
					continue
				}
				return errors.Wrap(err, "SvcAccept.AdditionalReqResult5GS.UnmarshalBinary")
			}
		case SvcAcceptIEIForbiddenTAI_5GSRoaming: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of ForbiddenTAI_5GSRoaming")
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
				return errors.Wrap(err, "SvcAccept.ForbiddenTAI_5GSRoaming.UnmarshalBinary")
			}
		case SvcAcceptIEIForbiddenTAI_5GSRegionalProvSvc: // TLV, 9-114B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SvcAccept UnmarshalBinary getIeLen of ForbiddenTAI_5GSRegionalProvSvc")
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
				return errors.Wrap(err, "SvcAccept.ForbiddenTAI_5GSRegionalProvSvc.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("SvcAccept unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *SvcAccept) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSvcAccept),
	})

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessReactivationResult TLV, 4-34B, IEI=0x26
	if m.PDUSessReactivationResult != nil {
		out, err := m.PDUSessReactivationResult.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.PDUSessReactivationResult.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIPDUSessReactivationResult)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PDUSessReactivationResultErrCause TLV-E, 5-515B, IEI=0x72
	if m.PDUSessReactivationResultErrCause != nil {
		out, err := m.PDUSessReactivationResultErrCause.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.PDUSessReactivationResultErrCause.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIPDUSessReactivationResultErrCause)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcAccept) MarshalBinary() binary write PDUSessReactivationResultErrCause")
		}
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SvcAccept) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.T3448Value TLV, 3B, IEI=0x6B
	if m.T3448Value != nil {
		out, err := m.T3448Value.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.T3448Value.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIT3448Value)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AdditionalReqResult5GS TLV, 3B, IEI=0x34
	if m.AdditionalReqResult5GS != nil {
		out, err := m.AdditionalReqResult5GS.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.AdditionalReqResult5GS.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIAdditionalReqResult5GS)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRoaming TLV, 9-114B, IEI=0x1D
	if m.ForbiddenTAI_5GSRoaming != nil {
		out, err := m.ForbiddenTAI_5GSRoaming.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.ForbiddenTAI_5GSRoaming.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIForbiddenTAI_5GSRoaming)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ForbiddenTAI_5GSRegionalProvSvc TLV, 9-114B, IEI=0x1E
	if m.ForbiddenTAI_5GSRegionalProvSvc != nil {
		out, err := m.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAccept.ForbiddenTAI_5GSRegionalProvSvc.MarshalBinary()")
		}
		writer.WriteByte(SvcAcceptIEIForbiddenTAI_5GSRegionalProvSvc)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
