package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SvcLvlAuthComplete{}

// SvcLvlAuthComplete is detailed in 8.3.18 Service-level authentication complete, 24.501
type SvcLvlAuthComplete struct {
	PDUSessId    uint8
	PTI          uint8
	SvcLvlAACntr *ie.SvcLvlAACntr //  LV-E,     5-nB, 9.11.2.10
}

func (m *SvcLvlAuthComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *SvcLvlAuthComplete) MsgType() MsgType {
	return MsgTypeSvcLvlAuthComplete
}

func (m *SvcLvlAuthComplete) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *SvcLvlAuthComplete) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *SvcLvlAuthComplete) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("SvcLvlAuthComplete len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	m.SvcLvlAACntr = new(ie.SvcLvlAACntr) // LV-E, 5-nB
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "SvcLvlAuthComplete UnmarshalBinary getIeLen of SvcLvlAACntr")
	}
	if err = m.SvcLvlAACntr.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "SvcLvlAuthComplete.SvcLvlAACntr.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *SvcLvlAuthComplete) MarshalBinary() ([]byte, error) {
	if m.SvcLvlAACntr == nil {
		return nil, errors.Errorf("SvcLvlAACntr=%v must present in SvcLvlAuthComplete",
			m.SvcLvlAACntr)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypeSvcLvlAuthComplete),
	})

	// svclvlaacntr, LV-E, 5-nB
	svclvlaacntr, err := m.SvcLvlAACntr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcLvlAuthComplete.SvcLvlAACntr.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(svclvlaacntr))); err != nil {
		return nil, errors.Wrap(err, "SvcLvlAuthComplete) MarshalBinary() binary write SvcLvlAACntr")
	}
	writer.Write(svclvlaacntr)

	return writer.Bytes(), nil
}
