package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &NwSliceSpecificAuthResult{}

// NwSliceSpecificAuthResult is detailed in 8.2.33 Network slice-specific authentication result, 24.501
type NwSliceSpecificAuthResult struct {
	SNSSAI *ie.SNSSAI //    LV,     2-5B, 9.11.2.8
	EAPMsg *ie.EAPMsg //  LV-E,  6-1502B, 9.11.2.2
}

func (m *NwSliceSpecificAuthResult) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *NwSliceSpecificAuthResult) MsgType() MsgType {
	return MsgTypeNwSliceSpecificAuthResult
}

func (m *NwSliceSpecificAuthResult) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("NwSliceSpecificAuthResult len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.SNSSAI = new(ie.SNSSAI) // LV, 2-5B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthResult UnmarshalBinary getIeLen of SNSSAI")
	}
	if err = m.SNSSAI.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthResult.SNSSAI.UnmarshalBinary")
	}

	m.EAPMsg = new(ie.EAPMsg) // LV-E, 6-1502B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthResult UnmarshalBinary getIeLen of EAPMsg")
	}
	if err = m.EAPMsg.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "NwSliceSpecificAuthResult.EAPMsg.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *NwSliceSpecificAuthResult) MarshalBinary() ([]byte, error) {
	if m.SNSSAI == nil || m.EAPMsg == nil {
		return nil, errors.Errorf("SNSSAI=%v EAPMsg=%v must present in NwSliceSpecificAuthResult",
			m.SNSSAI, m.EAPMsg)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeNwSliceSpecificAuthResult),
	})

	// snssai, LV, 2-5B
	snssai, err := m.SNSSAI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthResult.SNSSAI.MarshalBinary()")
	}
	writer.WriteByte(byte(len(snssai)))
	writer.Write(snssai)

	// eapmsg, LV-E, 6-1502B
	eapmsg, err := m.EAPMsg.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthResult.EAPMsg.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(eapmsg))); err != nil {
		return nil, errors.Wrap(err, "NwSliceSpecificAuthResult) MarshalBinary() binary write EAPMsg")
	}
	writer.Write(eapmsg)

	return writer.Bytes(), nil
}
