package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &NwSliceSpecificAuthComplete{}

// NwSliceSpecificAuthComplete is detailed in 8.2.32 Network slice-specific authentication complete, 24.501
type NwSliceSpecificAuthComplete struct {
	SNSSAI *ie.SNSSAI //    LV,     2-5B, 9.11.2.8
	EAPMsg *ie.EAPMsg //  LV-E,  6-1502B, 9.11.2.2
}

func (m *NwSliceSpecificAuthComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *NwSliceSpecificAuthComplete) MsgType() MsgType {
	return MsgTypeNwSliceSpecificAuthComplete
}

func (m *NwSliceSpecificAuthComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("NwSliceSpecificAuthComplete len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.SNSSAI = new(ie.SNSSAI) // LV, 2-5B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthComplete UnmarshalBinary getIeLen of SNSSAI")
	}
	if err = m.SNSSAI.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthComplete.SNSSAI.UnmarshalBinary")
	}

	m.EAPMsg = new(ie.EAPMsg) // LV-E, 6-1502B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthComplete UnmarshalBinary getIeLen of EAPMsg")
	}
	if err = m.EAPMsg.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthComplete.EAPMsg.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *NwSliceSpecificAuthComplete) MarshalBinary() ([]byte, error) {
	if m.SNSSAI == nil || m.EAPMsg == nil {
		return nil, errors.Errorf("SNSSAI=%v EAPMsg=%v must present in NwSliceSpecificAuthComplete",
			m.SNSSAI, m.EAPMsg)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeNwSliceSpecificAuthComplete),
	})

	// snssai, LV, 2-5B
	snssai, err := m.SNSSAI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthComplete.SNSSAI.MarshalBinary()")
	}
	writer.WriteByte(byte(len(snssai)))
	writer.Write(snssai)

	// eapmsg, LV-E, 6-1502B
	eapmsg, err := m.EAPMsg.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthComplete.EAPMsg.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(eapmsg))); err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthComplete) MarshalBinary() binary write EAPMsg")
	}
	writer.Write(eapmsg)

	return writer.Bytes(), nil
}
