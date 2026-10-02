package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SvcLvlAuthCmd{}

// SvcLvlAuthCmd is detailed in 8.3.17 Service-level authentication command, 24.501
type SvcLvlAuthCmd struct {
	PDUSessId    uint8
	PTI          uint8
	SvcLvlAACntr *ie.SvcLvlAACntr //  LV-E,     5-nB, 9.11.2.10
}

func (m *SvcLvlAuthCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *SvcLvlAuthCmd) MsgType() MsgType {
	return MsgTypeSvcLvlAuthCmd
}

func (m *SvcLvlAuthCmd) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *SvcLvlAuthCmd) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *SvcLvlAuthCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("SvcLvlAuthCmd len(b)=%d, < GsmHdrLen(%d)",
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
		return errors.Wrap(err, "SvcLvlAuthCmd UnmarshalBinary getIeLen of SvcLvlAACntr")
	}
	if err = m.SvcLvlAACntr.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "SvcLvlAuthCmd.SvcLvlAACntr.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *SvcLvlAuthCmd) MarshalBinary() ([]byte, error) {
	if m.SvcLvlAACntr == nil {
		return nil, errors.Errorf("SvcLvlAACntr=%v must present in SvcLvlAuthCmd",
			m.SvcLvlAACntr)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypeSvcLvlAuthCmd),
	})

	// svclvlaacntr, LV-E, 5-nB
	svclvlaacntr, err := m.SvcLvlAACntr.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SvcLvlAuthCmd.SvcLvlAACntr.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(svclvlaacntr))); err != nil {
		return nil, errors.Wrap(err, "SvcLvlAuthCmd) MarshalBinary() binary write SvcLvlAACntr")
	}
	writer.Write(svclvlaacntr)

	return writer.Bytes(), nil
}
