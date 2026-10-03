package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &Status5GSM{}

// Status5GSM is detailed in 8.3.16 5GSM status, 24.501
type Status5GSM struct {
	PDUSessId uint8
	PTI       uint8
	Cause5GSM *ie.Cause5GSM //     V,       1B, 9.11.4.2
}

func (m *Status5GSM) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *Status5GSM) MsgType() MsgType {
	return MsgTypeStatus5GSM
}

func (m *Status5GSM) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *Status5GSM) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *Status5GSM) UnmarshalBinary(b []byte) error {
	var err error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("Status5GSM len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	m.Cause5GSM = new(ie.Cause5GSM) // V, 1B
	if err = m.Cause5GSM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "Status5GSM.Cause5GSM.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *Status5GSM) MarshalBinary() ([]byte, error) {
	if m.Cause5GSM == nil {
		return nil, errors.Errorf("Cause5GSM=%v must present in Status5GSM",
			m.Cause5GSM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypeStatus5GSM),
	})

	// cause5gsm, V, 1B
	cause5gsm, err := m.Cause5GSM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "Status5GSM.Cause5GSM.MarshalBinary()")
	}
	writer.Write(cause5gsm)

	return writer.Bytes(), nil
}
