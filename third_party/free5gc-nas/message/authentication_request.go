package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &AuthReq{}

// AuthReq is detailed in 8.2.1 Authentication request, 24.501
type AuthReq struct {
	Ngksi                   *ie.NASKeySetId   //     V,     1/2B, 9.11.3.32
	SpareHalfOctet          struct{}          //     V,     1/2B, 9.5
	ABBA                    *ie.ABBA          //    LV,     3-nB, 9.11.3.10
	AuthParamRAND5GAuthChlg *ie.AuthParamRAND //    TV,      17B, 9.11.3.16
	AuthParamAUTN5GAuthChlg *ie.AuthParamAUTN //   TLV,      18B, 9.11.3.15
	EAPMsg                  *ie.EAPMsg        // TLV-E,  7-1503B, 9.11.2.2
}

const (
	AuthReqIEIAuthParamRAND5GAuthChlg uint8 = 0x21
	AuthReqIEIAuthParamAUTN5GAuthChlg uint8 = 0x20
	AuthReqIEIEAPMsg                  uint8 = 0x78
)

func (m *AuthReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *AuthReq) MsgType() MsgType {
	return MsgTypeAuthReq
}

func (m *AuthReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("AuthReq len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "AuthReq.Ngksi.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	m.ABBA = new(ie.ABBA) // LV, 3-nB
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "AuthReq UnmarshalBinary getIeLen of ABBA")
	}
	if err = m.ABBA.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "AuthReq.ABBA.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case AuthReqIEIAuthParamRAND5GAuthChlg: // TV, 17B
			ieLen = 16
			if m.AuthParamRAND5GAuthChlg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthParamRAND5GAuthChlg = new(ie.AuthParamRAND)
			if err = m.AuthParamRAND5GAuthChlg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthParamRAND5GAuthChlg = nil
					continue
				}
				return errors.Wrap(err, "AuthReq.AuthParamRAND5GAuthChlg.UnmarshalBinary")
			}
		case AuthReqIEIAuthParamAUTN5GAuthChlg: // TLV, 18B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "AuthReq UnmarshalBinary getIeLen of AuthParamAUTN5GAuthChlg")
			}
			if m.AuthParamAUTN5GAuthChlg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthParamAUTN5GAuthChlg = new(ie.AuthParamAUTN)
			if err = m.AuthParamAUTN5GAuthChlg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthParamAUTN5GAuthChlg = nil
					continue
				}
				return errors.Wrap(err, "AuthReq.AuthParamAUTN5GAuthChlg.UnmarshalBinary")
			}
		case AuthReqIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "AuthReq UnmarshalBinary getIeLen of EAPMsg")
			}
			if m.EAPMsg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EAPMsg = new(ie.EAPMsg)
			if err = m.EAPMsg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EAPMsg = nil
					continue
				}
				return errors.Wrap(err, "AuthReq.EAPMsg.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("AuthReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *AuthReq) MarshalBinary() ([]byte, error) {
	if m.Ngksi == nil || m.ABBA == nil {
		return nil, errors.Errorf("Ngksi=%v SpareHalfOctet=%v ABBA=%v must present in AuthReq",
			m.Ngksi, m.SpareHalfOctet, m.ABBA)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeAuthReq),
	})

	tmp := [1]byte{}
	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "AuthReq.Ngksi.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, ngksi[0])
	writer.Write(tmp[:])

	// abba, LV, 3-nB
	abba, err := m.ABBA.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "AuthReq.ABBA.MarshalBinary()")
	}
	writer.WriteByte(byte(len(abba)))
	writer.Write(abba)

	// m.AuthParamRAND5GAuthChlg TV, 17B, IEI=0x21
	if m.AuthParamRAND5GAuthChlg != nil {
		out, err := m.AuthParamRAND5GAuthChlg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthReq.AuthParamRAND5GAuthChlg.MarshalBinary()")
		}
		writer.WriteByte(AuthReqIEIAuthParamRAND5GAuthChlg)
		writer.Write(out)
	}

	// m.AuthParamAUTN5GAuthChlg TLV, 18B, IEI=0x20
	if m.AuthParamAUTN5GAuthChlg != nil {
		out, err := m.AuthParamAUTN5GAuthChlg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthReq.AuthParamAUTN5GAuthChlg.MarshalBinary()")
		}
		writer.WriteByte(AuthReqIEIAuthParamAUTN5GAuthChlg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthReq.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(AuthReqIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "AuthReq) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
