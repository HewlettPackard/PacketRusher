package message

import (
	"bytes"

	"github.com/pkg/errors"
)

var _ Message = &CfgUpdateComplete{}

// CfgUpdateComplete is detailed in 8.2.20 Configuration update complete, 24.501
type CfgUpdateComplete struct{}

func (m *CfgUpdateComplete) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *CfgUpdateComplete) MsgType() MsgType {
	return MsgTypeCfgUpdateComplete
}

func (m *CfgUpdateComplete) UnmarshalBinary(b []byte) error {
	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("CfgUpdateComplete len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// This message contains 0 Mandatory IE
	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *CfgUpdateComplete) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeCfgUpdateComplete),
	})

	return writer.Bytes(), nil
}
