package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SecModeRej{}

// SecModeRej is detailed in 8.2.27 Security mode reject, 24.501
type SecModeRej struct {
	Cause5GMM *ie.Cause5GMM //     V,       1B, 9.11.3.2
}

func (m *SecModeRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SecModeRej) MsgType() MsgType {
	return MsgTypeSecModeRej
}

func (m *SecModeRej) UnmarshalBinary(b []byte) error {
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SecModeRej len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.Cause5GMM = new(ie.Cause5GMM) // V, 1B
	if err = m.Cause5GMM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "SecModeRej.Cause5GMM.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *SecModeRej) MarshalBinary() ([]byte, error) {
	if m.Cause5GMM == nil {
		return nil, errors.Errorf("Cause5GMM=%v must present in SecModeRej",
			m.Cause5GMM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSecModeRej),
	})

	// cause5gmm, V, 1B
	cause5gmm, err := m.Cause5GMM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SecModeRej.Cause5GMM.MarshalBinary()")
	}
	writer.Write(cause5gmm)

	return writer.Bytes(), nil
}
