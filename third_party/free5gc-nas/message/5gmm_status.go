package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &Status5GMM{}

// Status5GMM is detailed in 8.2.29 5GMM status, 24.501
type Status5GMM struct {
	Cause5GMM *ie.Cause5GMM //     V,       1B, 9.11.3.2
}

func (m *Status5GMM) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *Status5GMM) MsgType() MsgType {
	return MsgTypeStatus5GMM
}

func (m *Status5GMM) UnmarshalBinary(b []byte) error {
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("Status5GMM len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.Cause5GMM = new(ie.Cause5GMM) // V, 1B
	if err = m.Cause5GMM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "Status5GMM.Cause5GMM.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *Status5GMM) MarshalBinary() ([]byte, error) {
	if m.Cause5GMM == nil {
		return nil, errors.Errorf("Cause5GMM=%v must present in Status5GMM",
			m.Cause5GMM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeStatus5GMM),
	})

	// cause5gmm, V, 1B
	cause5gmm, err := m.Cause5GMM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "Status5GMM.Cause5GMM.MarshalBinary()")
	}
	writer.Write(cause5gmm)

	return writer.Bytes(), nil
}
