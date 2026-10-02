package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &DeregReqUEOrig{}

// DeregReqUEOrig is detailed in 8.2.12 De-registration request (UE originating de-registration), 24.501
type DeregReqUEOrig struct {
	DeregType   *ie.DeregType   //     V,     1/2B, 9.11.3.20
	Ngksi       *ie.NASKeySetId //     V,     1/2B, 9.11.3.32
	MobileId5GS *ie.MobileId5GS //  LV-E,     6-nB, 9.11.3.4
}

func (m *DeregReqUEOrig) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *DeregReqUEOrig) MsgType() MsgType {
	return MsgTypeDeregReqUEOrig
}

func (m *DeregReqUEOrig) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("DeregReqUEOrig len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, half85 := ie.GetHalfIEValue(reader.Next(1))
	m.DeregType = new(ie.DeregType) // V, 1/2B
	if err = m.DeregType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "DeregReqUEOrig.DeregType.UnmarshalBinary")
	}
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half85); err != nil {
		return errors.Wrap(err, "DeregReqUEOrig.Ngksi.UnmarshalBinary")
	}

	m.MobileId5GS = new(ie.MobileId5GS) // LV-E, 6-nB
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "DeregReqUEOrig UnmarshalBinary getIeLen of MobileId5GS")
	}
	if err = m.MobileId5GS.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "DeregReqUEOrig.MobileId5GS.UnmarshalBinary")
	}

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *DeregReqUEOrig) MarshalBinary() ([]byte, error) {
	if m.DeregType == nil || m.Ngksi == nil || m.MobileId5GS == nil {
		return nil, errors.Errorf("DeregType=%v Ngksi=%v MobileId5GS=%v must present in DeregReqUEOrig",
			m.DeregType, m.Ngksi, m.MobileId5GS)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeDeregReqUEOrig),
	})

	tmp := [1]byte{}
	// deregtype, V, 1/2B
	deregtype, err := m.DeregType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DeregReqUEOrig.DeregType.MarshalBinary()")
	}

	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DeregReqUEOrig.Ngksi.MarshalBinary()")
	}
	tmp[0] = ie.SetHalfValue(ngksi[0], deregtype[0])
	writer.Write(tmp[:])

	// mobileid5gs, LV-E, 6-nB
	mobileid5gs, err := m.MobileId5GS.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "DeregReqUEOrig.MobileId5GS.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(mobileid5gs))); err != nil {
		return nil, errors.Wrap(err, "DeregReqUEOrig) MarshalBinary() binary write MobileId5GS")
	}
	writer.Write(mobileid5gs)

	return writer.Bytes(), nil
}
