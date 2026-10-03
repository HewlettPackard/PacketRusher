package message

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &AuthFailure{}

// AuthFailure is detailed in 8.2.4 Authentication failure, 24.501
type AuthFailure struct {
	Cause5GMM        *ie.Cause5GMM        //     V,       1B, 9.11.3.2
	AuthFailureParam *ie.AuthFailureParam //   TLV,      16B, 9.11.3.14
}

const (
	AuthFailureIEIAuthFailureParam uint8 = 0x30
)

func (m *AuthFailure) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *AuthFailure) MsgType() MsgType {
	return MsgTypeAuthFailure
}

func (m *AuthFailure) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("AuthFailure len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.Cause5GMM = new(ie.Cause5GMM) // V, 1B
	if err = m.Cause5GMM.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "AuthFailure.Cause5GMM.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case AuthFailureIEIAuthFailureParam: // TLV, 16B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "AuthFailure UnmarshalBinary getIeLen of AuthFailureParam")
			}
			if m.AuthFailureParam != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthFailureParam = new(ie.AuthFailureParam)
			if err = m.AuthFailureParam.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthFailureParam = nil
					continue
				}
				return errors.Wrap(err, "AuthFailure.AuthFailureParam.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("AuthFailure unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *AuthFailure) MarshalBinary() ([]byte, error) {
	if m.Cause5GMM == nil {
		return nil, errors.Errorf("Cause5GMM=%v must present in AuthFailure",
			m.Cause5GMM)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeAuthFailure),
	})

	// cause5gmm, V, 1B
	cause5gmm, err := m.Cause5GMM.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "AuthFailure.Cause5GMM.MarshalBinary()")
	}
	writer.Write(cause5gmm)

	// m.AuthFailureParam TLV, 16B, IEI=0x30
	if m.AuthFailureParam != nil {
		out, err := m.AuthFailureParam.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "AuthFailure.AuthFailureParam.MarshalBinary()")
		}
		writer.WriteByte(AuthFailureIEIAuthFailureParam)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
