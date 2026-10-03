package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RelayKeyReq{}

// RelayKeyReq is detailed in 8.2.34 Relay key request, 24.501
type RelayKeyReq struct {
	PRTI              *ie.ProseRelayTransactionId //     V,       1B, 9.11.3.88
	RelayKeyReqParams *ie.RelayKeyReqParams       //    LV, 22-65537B, 9.11.3.89
}

func (m *RelayKeyReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RelayKeyReq) MsgType() MsgType {
	return MsgTypeRelayKeyReq
}

func (m *RelayKeyReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RelayKeyReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.PRTI = new(ie.ProseRelayTransactionId) // V, 1B
	if err = m.PRTI.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "RelayKeyReq.PRTI.UnmarshalBinary")
	}

	m.RelayKeyReqParams = new(ie.RelayKeyReqParams) // LV, 22-65537B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "RelayKeyReq UnmarshalBinary getIeLen of RelayKeyReqParams")
	}
	if err = m.RelayKeyReqParams.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "RelayKeyReq.RelayKeyReqParams.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *RelayKeyReq) MarshalBinary() ([]byte, error) {
	if m.PRTI == nil || m.RelayKeyReqParams == nil {
		return nil, errors.Errorf("PRTI=%v RelayKeyReqParams=%v must present in RelayKeyReq",
			m.PRTI, m.RelayKeyReqParams)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRelayKeyReq),
	})

	// prti, V, 1B
	prti, err := m.PRTI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayKeyReq.PRTI.MarshalBinary()")
	}
	writer.Write(prti)

	// relaykeyreqparams, LV, 22-65537B
	relaykeyreqparams, err := m.RelayKeyReqParams.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayKeyReq.RelayKeyReqParams.MarshalBinary()")
	}
	writer.WriteByte(byte(len(relaykeyreqparams)))
	writer.Write(relaykeyreqparams)

	return writer.Bytes(), nil
}
