package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &SecModeCmd{}

// SecModeCmd is detailed in 8.2.25 Security mode command, 24.501
type SecModeCmd struct {
	SelectedNASSecAlgos         *ie.NASSecAlgos         //     V,       1B, 9.11.3.34
	Ngksi                       *ie.NASKeySetId         //     V,     1/2B, 9.11.3.32
	SpareHalfOctet              struct{}                //     V,     1/2B, 9.5
	ReplayedUESecCapabilities   *ie.UESecCapability     //    LV,     3-9B, 9.11.3.54
	IMEISVReq                   *ie.IMEISVReq           //    TV,       1B, 9.11.3.28
	SelectedEPSNASSecAlgos      *ie.EPSNASSecAlgos      //    TV,       2B, 9.11.3.25
	Additional5GSecInfo         *ie.Additional5GSecInfo //   TLV,       3B, 9.11.3.12
	EAPMsg                      *ie.EAPMsg              // TLV-E,  7-1503B, 9.11.2.2
	ABBA                        *ie.ABBA                //   TLV,     4-nB, 9.11.3.10
	ReplayedS1UESecCapabilities *ie.S1UESecCapability   //   TLV,     4-7B, 9.11.3.48A
}

const (
	SecModeCmdIEIIMEISVReq                   uint8 = 0xE0
	SecModeCmdIEISelectedEPSNASSecAlgos      uint8 = 0x57
	SecModeCmdIEIAdditional5GSecInfo         uint8 = 0x36
	SecModeCmdIEIEAPMsg                      uint8 = 0x78
	SecModeCmdIEIABBA                        uint8 = 0x38
	SecModeCmdIEIReplayedS1UESecCapabilities uint8 = 0x19
)

func (m *SecModeCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSMobilityMgmtMsg
}

func (m *SecModeCmd) MsgType() MsgType {
	return MsgTypeSecModeCmd
}

func (m *SecModeCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GmmHdrLen) {
		return errors.Errorf("SecModeCmd len(b)=%d, < GmmHdrLen(%d)",
			len(b), GmmHdrLen)
	}

	// skip 3 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[3:])

	// Mandatory IE
	m.SelectedNASSecAlgos = new(ie.NASSecAlgos) // V, 1B
	if err = m.SelectedNASSecAlgos.UnmarshalBinary(
		reader.Next(1)); err != nil {
		return errors.Wrap(err, "SecModeCmd.SelectedNASSecAlgos.UnmarshalBinary")
	}

	// 1/2B, handle with the other half
	half41, _ := ie.GetHalfIEValue(reader.Next(1))
	m.Ngksi = new(ie.NASKeySetId) // V, 1/2B
	if err = m.Ngksi.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "SecModeCmd.Ngksi.UnmarshalBinary")
	}
	// m.SpareHalfOctet, V, 1/2B

	m.ReplayedUESecCapabilities = new(ie.UESecCapability) // LV, 3-9B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "SecModeCmd UnmarshalBinary getIeLen of ReplayedUESecCapabilities")
	}
	if err = m.ReplayedUESecCapabilities.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "SecModeCmd.ReplayedUESecCapabilities.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case SecModeCmdIEIIMEISVReq: // TV, 1B
			if m.IMEISVReq != nil {
				break
			}
			m.IMEISVReq = new(ie.IMEISVReq)
			if err = m.IMEISVReq.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.IMEISVReq = nil
					continue
				}
				return errors.Wrap(err, "SecModeCmd.IMEISVReq.UnmarshalBinary")
			}
		case SecModeCmdIEISelectedEPSNASSecAlgos: // TV, 2B
			ieLen = 1
			if m.SelectedEPSNASSecAlgos != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SelectedEPSNASSecAlgos = new(ie.EPSNASSecAlgos)
			if err = m.SelectedEPSNASSecAlgos.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SelectedEPSNASSecAlgos = nil
					continue
				}
				return errors.Wrap(err, "SecModeCmd.SelectedEPSNASSecAlgos.UnmarshalBinary")
			}
		case SecModeCmdIEIAdditional5GSecInfo: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeCmd UnmarshalBinary getIeLen of Additional5GSecInfo")
			}
			if m.Additional5GSecInfo != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Additional5GSecInfo = new(ie.Additional5GSecInfo)
			if err = m.Additional5GSecInfo.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Additional5GSecInfo = nil
					continue
				}
				return errors.Wrap(err, "SecModeCmd.Additional5GSecInfo.UnmarshalBinary")
			}
		case SecModeCmdIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeCmd UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "SecModeCmd.EAPMsg.UnmarshalBinary")
			}
		case SecModeCmdIEIABBA: // TLV, 4-nB
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeCmd UnmarshalBinary getIeLen of ABBA")
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
				return errors.Wrap(err, "SecModeCmd.ABBA.UnmarshalBinary")
			}
		case SecModeCmdIEIReplayedS1UESecCapabilities: // TLV, 4-7B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "SecModeCmd UnmarshalBinary getIeLen of ReplayedS1UESecCapabilities")
			}
			if m.ReplayedS1UESecCapabilities != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReplayedS1UESecCapabilities = new(ie.S1UESecCapability)
			if err = m.ReplayedS1UESecCapabilities.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReplayedS1UESecCapabilities = nil
					continue
				}
				return errors.Wrap(err, "SecModeCmd.ReplayedS1UESecCapabilities.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("SecModeCmd unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *SecModeCmd) MarshalBinary() ([]byte, error) {
	if m.SelectedNASSecAlgos == nil || m.Ngksi == nil || m.ReplayedUESecCapabilities == nil {
		return nil, errors.Errorf("SelectedNASSecAlgo=%v Ngksi=%v ReplayedUESecCapabilities=%v must present in SecModeCmd",
			m.SelectedNASSecAlgos, m.Ngksi, m.ReplayedUESecCapabilities)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSMobilityMgmtMsg),
		byte(SecHdrTypePlainNas),
		byte(MsgTypeSecModeCmd),
	})

	// selectednassecalgos, V, 1B
	selectednassecalgos, err := m.SelectedNASSecAlgos.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SecModeCmd.SelectedNASSecAlgos.MarshalBinary()")
	}
	writer.Write(selectednassecalgos)

	tmp := [1]byte{}
	// ngksi, V, 1/2B
	ngksi, err := m.Ngksi.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SecModeCmd.Ngksi.MarshalBinary()")
	}

	// sparehalfoctet, V, 1/2B
	tmp[0] = ie.SetHalfValue(ie.SpareHalfOctet, ngksi[0])
	writer.Write(tmp[:])

	// replayedueseccapabilities, LV, 3-9B
	replayedueseccapabilities, err := m.ReplayedUESecCapabilities.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "SecModeCmd.ReplayedUESecCapabilities.MarshalBinary()")
	}
	writer.WriteByte(byte(len(replayedueseccapabilities)))
	writer.Write(replayedueseccapabilities)

	// m.IMEISVReq TV, 1B, IEI=0xE0, >= 0x80 !
	if m.IMEISVReq != nil {
		out, err := m.IMEISVReq.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.IMEISVReq.MarshalBinary()")
		}
		writer.WriteByte(out[0] | SecModeCmdIEIIMEISVReq)
	}

	// m.SelectedEPSNASSecAlgos TV, 2B, IEI=0x57
	if m.SelectedEPSNASSecAlgos != nil {
		out, err := m.SelectedEPSNASSecAlgos.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.SelectedEPSNASSecAlgos.MarshalBinary()")
		}
		writer.WriteByte(SecModeCmdIEISelectedEPSNASSecAlgos)
		writer.Write(out)
	}

	// m.Additional5GSecInfo TLV, 3B, IEI=0x36
	if m.Additional5GSecInfo != nil {
		out, err := m.Additional5GSecInfo.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.Additional5GSecInfo.MarshalBinary()")
		}
		writer.WriteByte(SecModeCmdIEIAdditional5GSecInfo)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(SecModeCmdIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "SecModeCmd) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.ABBA TLV, 4-nB, IEI=0x38
	if m.ABBA != nil {
		out, err := m.ABBA.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.ABBA.MarshalBinary()")
		}
		writer.WriteByte(SecModeCmdIEIABBA)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReplayedS1UESecCapabilities TLV, 4-7B, IEI=0x19
	if m.ReplayedS1UESecCapabilities != nil {
		out, err := m.ReplayedS1UESecCapabilities.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SecModeCmd.ReplayedS1UESecCapabilities.MarshalBinary()")
		}
		writer.WriteByte(SecModeCmdIEIReplayedS1UESecCapabilities)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
