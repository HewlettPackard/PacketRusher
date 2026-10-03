package message

import (
	"bytes"

	"github.com/pkg/errors"
)

var _ Message = &DeregAcceptUEOrig{}

// DeregAcceptUEOrig is detailed in 8.2.13 De-registration accept (UE originating de-registration), 24.501
type DeregAcceptUEOrig struct{}

func (m *DeregAcceptUEOrig) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *DeregAcceptUEOrig) MsgType() MsgType {
	return MsgTypeDeregAcceptUEOrig
}

func (m *DeregAcceptUEOrig) UnmarshalBinary(b []byte) error {
	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("DeregAcceptUEOrig len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// This message contains 0 Mandatory IE
	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *DeregAcceptUEOrig) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeDeregAcceptUEOrig),
	})

	return writer.Bytes(), nil
}
