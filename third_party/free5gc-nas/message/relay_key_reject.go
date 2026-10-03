package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RelayKeyRej{}

// RelayKeyRej is detailed in 8.2.36 Relay key reject, 24.501
type RelayKeyRej struct {
	PRTI   *ie.ProseRelayTransactionId //     V,       1B, 9.11.3.88
	EAPMsg *ie.EAPMsg                  //  LV-E,  6-1502B, 9.11.2.2
}

const (
	RelayKeyRejIEIEAPMsg uint8 = 0x78
)

func (m *RelayKeyRej) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RelayKeyRej) MsgType() MsgType {
	return MsgTypeRelayKeyRej
}

func (m *RelayKeyRej) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RelayKeyRej len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.PRTI = new(ie.ProseRelayTransactionId) // V, 1B
	if err = m.PRTI.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "RelayKeyRej.PRTI.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case RelayKeyRejIEIEAPMsg: // LV-E, 6-1502B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RelayKeyRej UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "RelayKeyRej.EAPMsg.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RelayKeyRej unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RelayKeyRej) MarshalBinary() ([]byte, error) {
	if m.PRTI == nil {
		return nil, errors.Errorf("PRTI=%v must present in RelayKeyRej",
			m.PRTI)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRelayKeyRej),
	})

	// prti, V, 1B
	prti, err := m.PRTI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayKeyRej.PRTI.MarshalBinary()")
	}
	writer.Write(prti)

	// m.EAPMsg LV-E, 6-1502B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RelayKeyRej.EAPMsg.MarshalBinary()")
		}
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RelayKeyRej) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
