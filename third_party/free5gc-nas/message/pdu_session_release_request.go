package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessRelReq{}

// PDUSessRelReq is detailed in 8.3.12 PDU session release request, 24.501
type PDUSessRelReq struct {
	PDUSessId           uint8
	PTI                 uint8
	Cause5GSM           *ie.Cause5GSM           //    TV,       2B, 9.11.4.2
	ExtendedProtCfgOpts *ie.ExtendedProtCfgOpts // TLV-E, 4-65538B, 9.11.4.6
}

const (
	PDUSessRelReqIEICause5GSM           uint8 = 0x59
	PDUSessRelReqIEIExtendedProtCfgOpts uint8 = 0x7B
)

func (m *PDUSessRelReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessRelReq) MsgType() MsgType {
	return MsgTypePDUSessRelReq
}

func (m *PDUSessRelReq) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessRelReq) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessRelReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessRelReq len(b)=%d, < GsmHdrLen(%d)",
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
		case PDUSessRelReqIEICause5GSM: // TV, 2B
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
				return errors.Wrap(err, "PDUSessRelReq.Cause5GSM.UnmarshalBinary")
			}
		case PDUSessRelReqIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessRelReq UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION RELEASE REQUEST is sent by the UE to the SMF
			if err = m.ExtendedProtCfgOpts.UnmarshalFromMs(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessRelReq.ExtendedProtCfgOpts.UnmarshalFromMs")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessRelReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessRelReq) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessRelReq),
	})

	// m.Cause5GSM TV, 2B, IEI=0x59
	if m.Cause5GSM != nil {
		out, err := m.Cause5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelReq.Cause5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelReqIEICause5GSM)
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessRelReq.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessRelReqIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessRelReq) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
