package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RelayAuthRsp{}

// RelayAuthRsp is detailed in 8.2.38 Relay authentication response, 24.501
type RelayAuthRsp struct {
	PRTI   *ie.ProseRelayTransactionId //     V,       1B, 9.11.3.88
	EAPMsg *ie.EAPMsg                  //  LV-E,  6-1502B, 9.11.2.2
}

func (m *RelayAuthRsp) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RelayAuthRsp) MsgType() MsgType {
	return MsgTypeRelayAuthRsp
}

func (m *RelayAuthRsp) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RelayAuthRsp len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.PRTI = new(ie.ProseRelayTransactionId) // V, 1B
	if err = m.PRTI.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "RelayAuthRsp.PRTI.UnmarshalBinary")
	}

	m.EAPMsg = new(ie.EAPMsg) // LV-E, 6-1502B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "RelayAuthRsp UnmarshalBinary getIeLen of EAPMsg")
	}
	if err = m.EAPMsg.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "RelayAuthRsp.EAPMsg.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *RelayAuthRsp) MarshalBinary() ([]byte, error) {
	if m.PRTI == nil || m.EAPMsg == nil {
		return nil, errors.Errorf("PRTI=%v EAPMsg=%v must present in RelayAuthRsp",
			m.PRTI, m.EAPMsg)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRelayAuthRsp),
	})

	// prti, V, 1B
	prti, err := m.PRTI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayAuthRsp.PRTI.MarshalBinary()")
	}
	writer.Write(prti)

	// eapmsg, LV-E, 6-1502B
	eapmsg, err := m.EAPMsg.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayAuthRsp.EAPMsg.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(eapmsg))); err != nil {
		return nil, errors.Wrap(err, "RelayAuthRsp) MarshalBinary() binary write EAPMsg")
	}
	writer.Write(eapmsg)

	return writer.Bytes(), nil
}
