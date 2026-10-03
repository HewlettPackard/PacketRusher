package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &IdReq{}

// IdReq is detailed in 8.2.21 Identity request, 24.501
type IdReq struct {
	IdType         *ie.IdType5GS //     V,     1/2B, 9.11.3.3
	SpareHalfOctet struct{}      //     V,     1/2B, 9.5
}

func (m *IdReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *IdReq) MsgType() MsgType {
	return MsgTypeIdReq
}

func (m *IdReq) UnmarshalBinary(b []byte) error {
	var err error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("IdReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.IdType = new(ie.IdType5GS) // V, 1/2B
	if err = m.IdType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "IdReq.IdType.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	// This message contains 0 Optional / Conditional IE
	return nil
}

func (m *IdReq) MarshalBinary() ([]byte, error) {
	if m.IdType == nil {
		return nil, errors.Errorf("IdType=%v SpareHalfOctet=%v must present in IdReq",
			m.IdType, m.SpareHalfOctet)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeIdReq),
	})

	tmp := [1]byte{}
	// idtype, V, 1/2B
	idtype, err := m.IdType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "IdReq.IdType.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, idtype[0])
	writer.Write(tmp[:])

	return writer.Bytes(), nil
}
