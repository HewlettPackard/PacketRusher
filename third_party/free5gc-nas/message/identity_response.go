package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &IdRsp{}

// IdRsp is detailed in 8.2.22 Identity response, 24.501
type IdRsp struct {
	MobileId *ie.MobileId5GS //  LV-E,     3-nB, 9.11.3.4
}

func (m *IdRsp) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *IdRsp) MsgType() MsgType {
	return MsgTypeIdRsp
}

func (m *IdRsp) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("IdRsp len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.MobileId = new(ie.MobileId5GS) // LV-E, 3-nB
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "IdRsp UnmarshalBinary getIeLen of MobileId")
	}
	if err = m.MobileId.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "IdRsp.MobileId.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *IdRsp) MarshalBinary() ([]byte, error) {
	if m.MobileId == nil {
		return nil, errors.Errorf("MobileId=%v must present in IdRsp",
			m.MobileId)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeIdRsp),
	})

	// mobileid, LV-E, 3-nB
	mobileid, err := m.MobileId.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "IdRsp.MobileId.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(mobileid))); err != nil {
		return nil, errors.Wrap(err, "IdRsp) MarshalBinary() binary write MobileId")
	}
	writer.Write(mobileid)

	return writer.Bytes(), nil
}
