package message

import (
	"bytes"

	"github.com/pkg/errors"
)

var _ Message = &RemoteUEReportRsp{}

// RemoteUEReportRsp is detailed in 8.3.20 Remote UE report response, 24.501
type RemoteUEReportRsp struct {
	PDUSessId uint8
	PTI       uint8
	// TODO: add IE
	// EapMsg               *ie.EAPMsg                   // TLV-E, 6-1502B, 9.11.2.2
	// RemoteUeHandlingInfo *ie.RemoteUeHandlingInfoList // TLV-E, 16-4-65538B, 9.11.4.f
	// AuthoQosFlowDescs    *ie.QosFlowDescs             // TLV-E, 6-65538B, 9.11.4.12
}

func (m *RemoteUEReportRsp) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *RemoteUEReportRsp) MsgType() MsgType {
	return MsgTypeRemoteUEReportRsp
}

func (m *RemoteUEReportRsp) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *RemoteUEReportRsp) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *RemoteUEReportRsp) UnmarshalBinary(b []byte) error {
	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("RemoteUEReportRsp len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// This message contains 0 Mandatory IE
	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *RemoteUEReportRsp) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypeRemoteUEReportRsp),
	})

	return writer.Bytes(), nil
}
