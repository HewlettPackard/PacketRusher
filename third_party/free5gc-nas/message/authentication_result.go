package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &AuthResult{}

// AuthResult is detailed in 8.2.3 Authentication result, 24.501
type AuthResult struct {
	Ngksi          *ie.NASKeySetId //     V,     1/2B, 9.11.3.32
	SpareHalfOctet struct{}        //     V,     1/2B, 9.5
	EAPMsg         *ie.EAPMsg      //  LV-E,  6-1502B, 9.11.2.2
	ABBA           *ie.ABBA        //   TLV,     4-nB, 9.11.3.10
}

const (
	AuthResultIEIABBA uint8 = 0x38
)

func (m *AuthResult) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *AuthResult) MsgType() MsgType {
	return MsgTypeAuthResult
}

func (m *AuthResult) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("AuthResult len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "AuthResult.Ngksi.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	m.EAPMsg = new(ie.EAPMsg) // LV-E, 6-1502B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "AuthResult UnmarshalBinary getIeLen of EAPMsg")
	}
	if err = m.EAPMsg.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "AuthResult.EAPMsg.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case AuthResultIEIABBA: // TLV, 4-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "AuthResult UnmarshalBinary getIeLen of ABBA")
			}
			if m.ABBA != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ABBA = new(ie.ABBA)
			if err = m.ABBA.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ABBA = nil
					continue
				}
				return errors.Wrap(err, "AuthResult.ABBA.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("AuthResult unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *AuthResult) MarshalBinary() ([]byte, error) {
	if m.Ngksi == nil || m.EAPMsg == nil {
		return nil, errors.Errorf("Ngksi=%v SpareHalfOctet=%v EAPMsg=%v must present in AuthResult",
			m.Ngksi, m.SpareHalfOctet, m.EAPMsg)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeAuthResult),
	})

	tmp := [1]byte{}
	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "AuthResult.Ngksi.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, ngksi[0])
	writer.Write(tmp[:])

	// eapmsg, LV-E, 6-1502B
	eapmsg, err := m.EAPMsg.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "AuthResult.EAPMsg.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(eapmsg))); err != nil {
		return nil, errors.Wrap(err, "AuthResult) MarshalBinary() binary write EAPMsg")
	}
	writer.Write(eapmsg)

	// m.ABBA TLV, 4-nB, IEI=0x38
	if m.ABBA != nil {
		out, err := m.ABBA.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthResult.ABBA.MarshalBinary()")
		}
		writer.WriteByte(AuthResultIEIABBA)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
