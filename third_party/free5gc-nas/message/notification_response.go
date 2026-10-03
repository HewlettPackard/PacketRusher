package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &NotifRsp{}

// NotifRsp is detailed in 8.2.24 Notification response, 24.501
type NotifRsp struct {
	PDUSessStatus *ie.PDUSessStatus //   TLV,    4-34B, 9.11.3.44
}

const (
	NotifRspIEIPDUSessStatus uint8 = 0x50
)

func (m *NotifRsp) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *NotifRsp) MsgType() MsgType {
	return MsgTypeNotifRsp
}

func (m *NotifRsp) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("NotifRsp len(b)=%d, < GmmHdrLen(%d)",
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
		case NotifRspIEIPDUSessStatus: // TLV, 4-34B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "NotifRsp UnmarshalBinary getIeLen of PDUSessStatus")
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
				return errors.Wrap(err, "NotifRsp.PDUSessStatus.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("NotifRsp unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *NotifRsp) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeNotifRsp),
	})

	// m.PDUSessStatus TLV, 4-34B, IEI=0x50
	if m.PDUSessStatus != nil {
		out, err := m.PDUSessStatus.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "NotifRsp.PDUSessStatus.MarshalBinary()")
		}
		writer.WriteByte(NotifRspIEIPDUSessStatus)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
