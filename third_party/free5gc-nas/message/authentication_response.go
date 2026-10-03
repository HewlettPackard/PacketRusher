package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &AuthRsp{}

// AuthRsp is detailed in 8.2.2 Authentication response, 24.501
type AuthRsp struct {
	AuthRspParam *ie.AuthRspParam //   TLV,      18B, 9.11.3.17
	EAPMsg       *ie.EAPMsg       // TLV-E,  7-1503B, 9.11.2.2
}

const (
	AuthRspIEIAuthRspParam uint8 = 0x2D
	AuthRspIEIEAPMsg       uint8 = 0x78
)

func (m *AuthRsp) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *AuthRsp) MsgType() MsgType {
	return MsgTypeAuthRsp
}

func (m *AuthRsp) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("AuthRsp len(b)=%d, < GmmHdrLen(%d)",
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
		case AuthRspIEIAuthRspParam: // TLV, 18B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "AuthRsp UnmarshalBinary getIeLen of AuthRspParam")
			}
			if m.AuthRspParam != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthRspParam = new(ie.AuthRspParam)
			if err = m.AuthRspParam.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthRspParam = nil
					continue
				}
				return errors.Wrap(err, "AuthRsp.AuthRspParam.UnmarshalBinary")
			}
		case AuthRspIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "AuthRsp UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "AuthRsp.EAPMsg.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("AuthRsp unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *AuthRsp) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeAuthRsp),
	})

	// m.AuthRspParam TLV, 18B, IEI=0x2D
	if m.AuthRspParam != nil {
		out, err := m.AuthRspParam.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthRsp.AuthRspParam.MarshalBinary()")
		}
		writer.WriteByte(AuthRspIEIAuthRspParam)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthRsp.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(AuthRspIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "AuthRsp) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
