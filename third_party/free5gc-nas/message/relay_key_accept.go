package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &RelayKeyAccept{}

// RelayKeyAccept is detailed in 8.2.35 Relay key accept, 24.501
type RelayKeyAccept struct {
	PRTI              *ie.ProseRelayTransactionId //     V,       1B, 9.11.3.88
	RelayKeyRspParams *ie.RelayKeyRspParams       //  LV-E, 51-65537B, 9.11.3.90
	EAPMsg            *ie.EAPMsg                  // TLV-E,  7-1503B, 9.11.2.2
}

const (
	RelayKeyAcceptIEIEAPMsg uint8 = 0x78
)

func (m *RelayKeyAccept) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *RelayKeyAccept) MsgType() MsgType {
	return MsgTypeRelayKeyAccept
}

func (m *RelayKeyAccept) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("RelayKeyAccept len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.PRTI = new(ie.ProseRelayTransactionId) // V, 1B
	if err = m.PRTI.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "RelayKeyAccept.PRTI.UnmarshalBinary")
	}

	m.RelayKeyRspParams = new(ie.RelayKeyRspParams) // LV-E, 51-65537B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "RelayKeyAccept UnmarshalBinary getIeLen of RelayKeyRspParams")
	}
	if err = m.RelayKeyRspParams.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "RelayKeyAccept.RelayKeyRspParams.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case RelayKeyAcceptIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "RelayKeyAccept UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "RelayKeyAccept.EAPMsg.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("RelayKeyAccept unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *RelayKeyAccept) MarshalBinary() ([]byte, error) {
	if m.PRTI == nil || m.RelayKeyRspParams == nil {
		return nil, errors.Errorf("PRTI=%v RelayKeyRspParams=%v must present in RelayKeyAccept",
			m.PRTI, m.RelayKeyRspParams)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeRelayKeyAccept),
	})

	// prti, V, 1B
	prti, err := m.PRTI.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayKeyAccept.PRTI.MarshalBinary()")
	}
	writer.Write(prti)

	// relaykeyrspparams, LV-E, 51-65537B
	relaykeyrspparams, err := m.RelayKeyRspParams.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "RelayKeyAccept.RelayKeyRspParams.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(relaykeyrspparams))); err != nil {
		return nil, errors.Wrap(err, "RelayKeyAccept) MarshalBinary() binary write RelayKeyRspParams")
	}
	writer.Write(relaykeyrspparams)

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "RelayKeyAccept.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(RelayKeyAcceptIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "RelayKeyAccept) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
