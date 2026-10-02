package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &DLNASTransport{}

// DLNASTransport is detailed in 8.2.11 DL NAS transport, 24.501
type DLNASTransport struct {
	PayloadCntrType      *ie.PayloadCntrType //     V,     1/2B, 9.11.3.40
	SpareHalfOctet       struct{}            //     V,     1/2B, 9.5
	PayloadCntr          *ie.PayloadCntr     //  LV-E, 3-65537B, 9.11.3.39
	PDUSessID            *ie.PDUSessId2      //    TV,       2B, 9.11.3.41
	AdditionalInfo       *ie.AdditionalInfo  //   TLV,     3-nB, 9.11.2.1
	Cause5GMM            *ie.Cause5GMM       //    TV,       2B, 9.11.3.2
	BackoffTimerValue    *ie.GPRSTimer3      //   TLV,       3B, 9.11.2.5
	LowerBoundTimerValue *ie.GPRSTimer3      //   TLV,       3B, 9.11.2.5
}

const (
	DLNASTransportIEIPDUSessID            uint8 = 0x12
	DLNASTransportIEIAdditionalInfo       uint8 = 0x24
	DLNASTransportIEICause5GMM            uint8 = 0x58
	DLNASTransportIEIBackoffTimerValue    uint8 = 0x37
	DLNASTransportIEILowerBoundTimerValue uint8 = 0x3A
)

func (m *DLNASTransport) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *DLNASTransport) MsgType() MsgType {
	return MsgTypeDLNASTransport
}

func (m *DLNASTransport) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("DLNASTransport len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.PayloadCntrType = new(ie.PayloadCntrType) // V, 1/2B
	if err = m.PayloadCntrType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "DLNASTransport.PayloadCntrType.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	m.PayloadCntr = new(ie.PayloadCntr) // LV-E, 3-65537B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "DLNASTransport UnmarshalBinary getIeLen of PayloadCntr")
	}
	if err = m.PayloadCntr.UnmarshalBinary(
		reader.Next(int(ieLen)), m.PayloadCntrType.Value); err != nil {
		return errors.Wrap(err, "DLNASTransport.PayloadCntr.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case DLNASTransportIEIPDUSessID: // TV, 2B
			ieLen = 1
			if m.PDUSessID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessID = new(ie.PDUSessId2)
			if err = m.PDUSessID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessID = nil
					continue
				}
				return errors.Wrap(err, "DLNASTransport.PDUSessID.UnmarshalBinary")
			}
		case DLNASTransportIEIAdditionalInfo: // TLV, 3-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DLNASTransport UnmarshalBinary getIeLen of AdditionalInfo")
			}
			if m.AdditionalInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AdditionalInfo = new(ie.AdditionalInfo)
			if err = m.AdditionalInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AdditionalInfo = nil
					continue
				}
				return errors.Wrap(err, "DLNASTransport.AdditionalInfo.UnmarshalBinary")
			}
		case DLNASTransportIEICause5GMM: // TV, 2B
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
				return errors.Wrap(err, "DLNASTransport.Cause5GMM.UnmarshalBinary")
			}
		case DLNASTransportIEIBackoffTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DLNASTransport UnmarshalBinary getIeLen of BackoffTimerValue")
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
				return errors.Wrap(err, "DLNASTransport.BackoffTimerValue.UnmarshalBinary")
			}
		case DLNASTransportIEILowerBoundTimerValue: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "DLNASTransport UnmarshalBinary getIeLen of LowerBoundTimerValue")
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
				return errors.Wrap(err, "DLNASTransport.LowerBoundTimerValue.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("DLNASTransport unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *DLNASTransport) MarshalBinary() ([]byte, error) {
	if m.PayloadCntrType == nil || m.PayloadCntr == nil {
		return nil, errors.Errorf("PayloadCntrType=%v PayloadCntr=%v must present in DLNASTransport",
			m.PayloadCntrType, m.PayloadCntr)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeDLNASTransport),
	})

	tmp := [1]byte{}
	// payloadcntrtype, V, 1/2B
	payloadcntrtype, err := m.PayloadCntrType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DLNASTransport.PayloadCntrType.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B, hardcoded as 0

	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, payloadcntrtype[0])
	writer.Write(tmp[:])

	// payloadcntr, LV-E, 3-65537B
	payloadcntr, err := m.PayloadCntr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DLNASTransport.PayloadCntr.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(payloadcntr))); err != nil {
		return nil, errors.Wrap(err, "DLNASTransport) MarshalBinary() binary write PayloadCntr")
	}
	writer.Write(payloadcntr)

	// m.PDUSessID TV, 2B, IEI=0x12
	if m.PDUSessID != nil {
		out, err := m.PDUSessID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DLNASTransport.PDUSessID.MarshalBinary()")
		}
		writer.WriteByte(DLNASTransportIEIPDUSessID)
		writer.Write(out)
	}

	// m.AdditionalInfo TLV, 3-nB, IEI=0x24
	if m.AdditionalInfo != nil {
		out, err := m.AdditionalInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DLNASTransport.AdditionalInfo.MarshalBinary()")
		}
		writer.WriteByte(DLNASTransportIEIAdditionalInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.Cause5GMM TV, 2B, IEI=0x58
	if m.Cause5GMM != nil {
		out, err := m.Cause5GMM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DLNASTransport.Cause5GMM.MarshalBinary()")
		}
		writer.WriteByte(DLNASTransportIEICause5GMM)
		writer.Write(out)
	}

	// m.BackoffTimerValue TLV, 3B, IEI=0x37
	if m.BackoffTimerValue != nil {
		out, err := m.BackoffTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DLNASTransport.BackoffTimerValue.MarshalBinary()")
		}
		writer.WriteByte(DLNASTransportIEIBackoffTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.LowerBoundTimerValue TLV, 3B, IEI=0x3A
	if m.LowerBoundTimerValue != nil {
		out, err := m.LowerBoundTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "DLNASTransport.LowerBoundTimerValue.MarshalBinary()")
		}
		writer.WriteByte(DLNASTransportIEILowerBoundTimerValue)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
