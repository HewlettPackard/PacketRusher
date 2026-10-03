package message

import (
	"bytes"

	"github.com/pkg/errors"
)

var _ Message = &DeregAcceptUETerm{}

// DeregAcceptUETerm is detailed in 8.2.15 De-registration accept (UE terminated de-registration), 24.501
type DeregAcceptUETerm struct{}

func (m *DeregAcceptUETerm) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *DeregAcceptUETerm) MsgType() MsgType {
	return MsgTypeDeregAcceptUETerm
}

func (m *DeregAcceptUETerm) UnmarshalBinary(b []byte) error {
	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("DeregAcceptUETerm len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// This message contains 0 Mandatory IE
	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *DeregAcceptUETerm) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeDeregAcceptUETerm),
	})

	return writer.Bytes(), nil
}
