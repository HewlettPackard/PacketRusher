package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RemoteUEReport{}

// RemoteUEReport is detailed in 8.3.19 Remote UE report, 24.501
type RemoteUEReport struct {
	PDUSessId               uint8
	PTI                     uint8
	RemoteUECtxConnected    *ie.RemoteUECtxList // TLV-E, 16-65538B, 9.11.4.29
	RemoteUECtxDisconnected *ie.RemoteUECtxList // TLV-E, 16-65538B, 9.11.4.29
}

const (
	RemoteUEReportIEIRemoteUECtxConnected    uint8 = 0x76
	RemoteUEReportIEIRemoteUECtxDisconnected uint8 = 0x70
)

func (m *RemoteUEReport) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *RemoteUEReport) MsgType() MsgType {
	return MsgTypeRemoteUEReport
}

func (m *RemoteUEReport) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *RemoteUEReport) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *RemoteUEReport) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("RemoteUEReport len(b)=%d, < GsmHdrLen(%d)",
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
		case RemoteUEReportIEIRemoteUECtxConnected: // TLV-E, 16-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RemoteUEReport UnmarshalBinary getIeLen of RemoteUECtxConnected")
			}
			if m.RemoteUECtxConnected != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RemoteUECtxConnected = new(ie.RemoteUECtxList)
			if err = m.RemoteUECtxConnected.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RemoteUECtxConnected = nil
					continue
				}
				return errors.Wrap(err, "RemoteUEReport.RemoteUECtxConnected.UnmarshalBinary")
			}
		case RemoteUEReportIEIRemoteUECtxDisconnected: // TLV-E, 16-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RemoteUEReport UnmarshalBinary getIeLen of RemoteUECtxDisconnected")
			}
			if m.RemoteUECtxDisconnected != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RemoteUECtxDisconnected = new(ie.RemoteUECtxList)
			if err = m.RemoteUECtxDisconnected.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RemoteUECtxDisconnected = nil
					continue
				}
				return errors.Wrap(err, "RemoteUEReport.RemoteUECtxDisconnected.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RemoteUEReport unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RemoteUEReport) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypeRemoteUEReport),
	})

	// m.RemoteUECtxConnected TLV-E, 16-65538B, IEI=0x76
	if m.RemoteUECtxConnected != nil {
		out, err := m.RemoteUECtxConnected.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RemoteUEReport.RemoteUECtxConnected.MarshalBinary()")
		}
		writer.WriteByte(RemoteUEReportIEIRemoteUECtxConnected)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RemoteUEReport) MarshalBinary() binary write RemoteUECtxConnected")
		}
		writer.Write(out)
	}

	// m.RemoteUECtxDisconnected TLV-E, 16-65538B, IEI=0x70
	if m.RemoteUECtxDisconnected != nil {
		out, err := m.RemoteUECtxDisconnected.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RemoteUEReport.RemoteUECtxDisconnected.MarshalBinary()")
		}
		writer.WriteByte(RemoteUEReportIEIRemoteUECtxDisconnected)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RemoteUEReport) MarshalBinary() binary write RemoteUECtxDisconnected")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
