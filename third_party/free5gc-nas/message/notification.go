package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &Notif{}

// Notif is detailed in 8.2.23 Notification, 24.501
type Notif struct {
	AccessType     *ie.AccessType //     V,     1/2B, 9.11.2.1A
	SpareHalfOctet struct{}       //     V,     1/2B, 9.5
}

func (m *Notif) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *Notif) MsgType() MsgType {
	return MsgTypeNotif
}

func (m *Notif) UnmarshalBinary(b []byte) error {
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("Notif len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.AccessType = new(ie.AccessType) // V, 1/2B
	if err = m.AccessType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "Notif.AccessType.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *Notif) MarshalBinary() ([]byte, error) {
	if m.AccessType == nil {
		return nil, errors.Errorf("AccessType=%v SpareHalfOctet=%v must present in Notif",
			m.AccessType, m.SpareHalfOctet)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeNotif),
	})

	tmp := [1]byte{}
	// accesstype, V, 1/2B
	accesstype, err := m.AccessType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "Notif.AccessType.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, accesstype[0])
	writer.Write(tmp[:])

	return writer.Bytes(), nil
}
