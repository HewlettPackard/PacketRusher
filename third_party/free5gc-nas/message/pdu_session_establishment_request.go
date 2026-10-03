package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessEstReq{}

// PDUSessEstReq is detailed in 8.3.1 PDU session establishment request, 24.501
type PDUSessEstReq struct {
	PDUSessId                      uint8
	PTI                            uint8
	IntegrityProtectionMaxDataRate *ie.IntegrityProtectionMaxDataRate //     V,       2B, 9.11.4.7
	PDUSessType                    *ie.PDUSessType                    //    TV,       1B, 9.11.4.11
	SSCMode                        *ie.SSCMode                        //    TV,       1B, 9.11.4.16
	Capability5GSM                 *ie.Capability5GSM                 //   TLV,    3-15B, 9.11.4.1
	MaxNumOfSupportedPktFilters    *ie.MaxNumOfSupportedPktFilters    //    TV,       3B, 9.11.4.9
	AlwaysonPDUSessReq             *ie.AlwaysonPDUSessReq             //    TV,       1B, 9.11.4.4
	SMPDUDNReqCntr                 *ie.SMPDUDNReqCntr                 //   TLV,   3-255B, 9.11.4.15
	ExtendedProtCfgOpts            *ie.ExtendedProtCfgOpts            // TLV-E, 4-65538B, 9.11.4.6
	IPHdrCompressionCfg            *ie.IPHdrCompressionCfg            //   TLV,   5-257B, 9.11.4.24
	DSTTEthPortMACAddr             *ie.DSTTEthPortMACAddr             //   TLV,       8B, 9.11.4.25
	UEDSTTResidenceTime            *ie.UEDSTTResidenceTime            //   TLV,      10B, 9.11.4.26
	PortMgmtInfoCntr               *ie.PortMgmtInfoCntr               // TLV-E, 8-65538B, 9.11.4.27
	EthHdrCompressionCfg           *ie.EthHdrCompressionCfg           //   TLV,       3B, 9.11.4.28
	SuggestedIfId                  *ie.PDUAddr                        //   TLV,      11B, 9.11.4.10
	SvcLvlAACntr                   *ie.SvcLvlAACntr                   // TLV-E, 4-65538B, 9.11.2.10
	ReqMBSCntr                     *ie.ReqMBSCntr                     // TLV-E, 8-65538B, 9.11.4.30
	PDUSessPairID                  *ie.PDUSessPairID                  //   TLV,       3B, 9.11.4.32
	RSN                            *ie.RSN                            //   TLV,       3B, 9.11.4.33
}

const (
	PDUSessEstReqIEIPDUSessType                 uint8 = 0x90
	PDUSessEstReqIEISSCMode                     uint8 = 0xA0
	PDUSessEstReqIEICapability5GSM              uint8 = 0x28
	PDUSessEstReqIEIMaxNumOfSupportedPktFilters uint8 = 0x55
	PDUSessEstReqIEIAlwaysonPDUSessReq          uint8 = 0xB0
	PDUSessEstReqIEISMPDUDNReqCntr              uint8 = 0x39
	PDUSessEstReqIEIExtendedProtCfgOpts         uint8 = 0x7B
	PDUSessEstReqIEIIPHdrCompressionCfg         uint8 = 0x66
	PDUSessEstReqIEIDSTTEthPortMACAddr          uint8 = 0x6E
	PDUSessEstReqIEIUEDSTTResidenceTime         uint8 = 0x6F
	PDUSessEstReqIEIPortMgmtInfoCntr            uint8 = 0x74
	PDUSessEstReqIEIEthHdrCompressionCfg        uint8 = 0x1F
	PDUSessEstReqIEISuggestedIfId               uint8 = 0x29
	PDUSessEstReqIEISvcLvlAACntr                uint8 = 0x72
	PDUSessEstReqIEIReqMBSCntr                  uint8 = 0x70
	PDUSessEstReqIEIPDUSessPairID               uint8 = 0x34
	PDUSessEstReqIEIRSN                         uint8 = 0x35
)

func (m *PDUSessEstReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessEstReq) MsgType() MsgType {
	return MsgTypePDUSessEstReq
}

func (m *PDUSessEstReq) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessEstReq) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessEstReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessEstReq len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	m.IntegrityProtectionMaxDataRate = new(ie.IntegrityProtectionMaxDataRate) // V, 2B
	if err = m.IntegrityProtectionMaxDataRate.UnmarshalBinary(
		reader.Next(2)); err != nil {
		return errors.Wrap(err, "PDUSessEstReq.IntegrityProtectionMaxDataRate.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessEstReqIEIPDUSessType: // TV, 1B
			if m.PDUSessType != nil {
				break
			}
			m.PDUSessType = new(ie.PDUSessType)
			if err = m.PDUSessType.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessType = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.PDUSessType.UnmarshalBinary")
			}
		case PDUSessEstReqIEISSCMode: // TV, 1B
			if m.SSCMode != nil {
				break
			}
			m.SSCMode = new(ie.SSCMode)
			if err = m.SSCMode.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SSCMode = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.SSCMode.UnmarshalBinary")
			}
		case PDUSessEstReqIEICapability5GSM: // TLV, 3-15B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of Capability5GSM")
			}
			if m.Capability5GSM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Capability5GSM = new(ie.Capability5GSM)
			if err = m.Capability5GSM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Capability5GSM = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.Capability5GSM.UnmarshalBinary")
			}
		case PDUSessEstReqIEIMaxNumOfSupportedPktFilters: // TV, 3B
			ieLen = 2
			if m.MaxNumOfSupportedPktFilters != nil {
				reader.Next(int(ieLen))
				break
			}
			m.MaxNumOfSupportedPktFilters = new(ie.MaxNumOfSupportedPktFilters)
			if err = m.MaxNumOfSupportedPktFilters.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.MaxNumOfSupportedPktFilters = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.MaxNumOfSupportedPktFilters.UnmarshalBinary")
			}
		case PDUSessEstReqIEIAlwaysonPDUSessReq: // TV, 1B
			if m.AlwaysonPDUSessReq != nil {
				break
			}
			m.AlwaysonPDUSessReq = new(ie.AlwaysonPDUSessReq)
			if err = m.AlwaysonPDUSessReq.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AlwaysonPDUSessReq = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.AlwaysonPDUSessReq.UnmarshalBinary")
			}
		case PDUSessEstReqIEISMPDUDNReqCntr: // TLV, 3-255B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of SMPDUDNReqCntr")
			}
			if m.SMPDUDNReqCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SMPDUDNReqCntr = new(ie.SMPDUDNReqCntr)
			if err = m.SMPDUDNReqCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SMPDUDNReqCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.SMPDUDNReqCntr.UnmarshalBinary")
			}
		case PDUSessEstReqIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// The PDU SESSION ESTABLISHMENT REQUEST message is sent by the UE to the SMF
			if err = m.ExtendedProtCfgOpts.UnmarshalFromMs(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.ExtendedProtCfgOpts.UnmarshalFromMs")
			}
		case PDUSessEstReqIEIIPHdrCompressionCfg: // TLV, 5-257B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of IPHdrCompressionCfg")
			}
			if m.IPHdrCompressionCfg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.IPHdrCompressionCfg = new(ie.IPHdrCompressionCfg)
			if err = m.IPHdrCompressionCfg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.IPHdrCompressionCfg = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.IPHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessEstReqIEIDSTTEthPortMACAddr: // TLV, 8B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of DSTTEthPortMACAddr")
			}
			if m.DSTTEthPortMACAddr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DSTTEthPortMACAddr = new(ie.DSTTEthPortMACAddr)
			if err = m.DSTTEthPortMACAddr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DSTTEthPortMACAddr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.DSTTEthPortMACAddr.UnmarshalBinary")
			}
		case PDUSessEstReqIEIUEDSTTResidenceTime: // TLV, 10B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of UEDSTTResidenceTime")
			}
			if m.UEDSTTResidenceTime != nil {
				reader.Next(int(ieLen))
				break
			}
			m.UEDSTTResidenceTime = new(ie.UEDSTTResidenceTime)
			if err = m.UEDSTTResidenceTime.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.UEDSTTResidenceTime = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.UEDSTTResidenceTime.UnmarshalBinary")
			}
		case PDUSessEstReqIEIPortMgmtInfoCntr: // TLV-E, 8-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of PortMgmtInfoCntr")
			}
			if m.PortMgmtInfoCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PortMgmtInfoCntr = new(ie.PortMgmtInfoCntr)
			if err = m.PortMgmtInfoCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PortMgmtInfoCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.PortMgmtInfoCntr.UnmarshalBinary")
			}
		case PDUSessEstReqIEIEthHdrCompressionCfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of EthHdrCompressionCfg")
			}
			if m.EthHdrCompressionCfg != nil {
				reader.Next(int(ieLen))
				break
			}
			m.EthHdrCompressionCfg = new(ie.EthHdrCompressionCfg)
			if err = m.EthHdrCompressionCfg.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.EthHdrCompressionCfg = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.EthHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessEstReqIEISuggestedIfId: // TLV, 11B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of SuggestedIfId")
			}
			if m.SuggestedIfId != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SuggestedIfId = new(ie.PDUAddr)
			if err = m.SuggestedIfId.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SuggestedIfId = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.SuggestedIfId.UnmarshalBinary")
			}
		case PDUSessEstReqIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of SvcLvlAACntr")
			}
			if m.SvcLvlAACntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SvcLvlAACntr = new(ie.SvcLvlAACntr)
			if err = m.SvcLvlAACntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SvcLvlAACntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.SvcLvlAACntr.UnmarshalBinary")
			}
		case PDUSessEstReqIEIReqMBSCntr: // TLV-E, 8-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of ReqMBSCntr")
			}
			if m.ReqMBSCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqMBSCntr = new(ie.ReqMBSCntr)
			if err = m.ReqMBSCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqMBSCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.ReqMBSCntr.UnmarshalBinary")
			}
		case PDUSessEstReqIEIPDUSessPairID: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of PDUSessPairID")
			}
			if m.PDUSessPairID != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUSessPairID = new(ie.PDUSessPairID)
			if err = m.PDUSessPairID.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUSessPairID = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.PDUSessPairID.UnmarshalBinary")
			}
		case PDUSessEstReqIEIRSN: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstReq UnmarshalBinary getIeLen of RSN")
			}
			if m.RSN != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RSN = new(ie.RSN)
			if err = m.RSN.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RSN = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstReq.RSN.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessEstReq unknown iei[%d]", iei)
		}
	}

	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessEstReq) MarshalBinary() ([]byte, error) {
	if m.IntegrityProtectionMaxDataRate == nil {
		return nil, errors.Errorf("IntegrityProtectionMaxDataRate=%v must present in PDUSessEstReq",
			m.IntegrityProtectionMaxDataRate)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessEstReq),
	})

	// integrityprotectionmaxdatarate, V, 2B
	integrityprotectionmaxdatarate, err := m.IntegrityProtectionMaxDataRate.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstReq.IntegrityProtectionMaxDataRate.MarshalBinary()")
	}
	writer.Write(integrityprotectionmaxdatarate)

	// m.PDUSessType TV, 1B, IEI=0x90, >= 0x80 !
	if m.PDUSessType != nil {
		out, err := m.PDUSessType.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.PDUSessType.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstReqIEIPDUSessType)
	}

	// m.SSCMode TV, 1B, IEI=0xA0, >= 0x80 !
	if m.SSCMode != nil {
		out, err := m.SSCMode.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.SSCMode.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstReqIEISSCMode)
	}

	// m.Capability5GSM TLV, 3-15B, IEI=0x28
	if m.Capability5GSM != nil {
		out, err := m.Capability5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.Capability5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEICapability5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.MaxNumOfSupportedPktFilters TV, 3B, IEI=0x55
	if m.MaxNumOfSupportedPktFilters != nil {
		out, err := m.MaxNumOfSupportedPktFilters.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.MaxNumOfSupportedPktFilters.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIMaxNumOfSupportedPktFilters)
		writer.Write(out)
	}

	// m.AlwaysonPDUSessReq TV, 1B, IEI=0xB0, >= 0x80 !
	if m.AlwaysonPDUSessReq != nil {
		out, err := m.AlwaysonPDUSessReq.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.AlwaysonPDUSessReq.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstReqIEIAlwaysonPDUSessReq)
	}

	// m.SMPDUDNReqCntr TLV, 3-255B, IEI=0x39
	if m.SMPDUDNReqCntr != nil {
		out, err := m.SMPDUDNReqCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.SMPDUDNReqCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEISMPDUDNReqCntr)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.IPHdrCompressionCfg TLV, 5-257B, IEI=0x66
	if m.IPHdrCompressionCfg != nil {
		out, err := m.IPHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.IPHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIIPHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.DSTTEthPortMACAddr TLV, 8B, IEI=0x6E
	if m.DSTTEthPortMACAddr != nil {
		out, err := m.DSTTEthPortMACAddr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.DSTTEthPortMACAddr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIDSTTEthPortMACAddr)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.UEDSTTResidenceTime TLV, 10B, IEI=0x6F
	if m.UEDSTTResidenceTime != nil {
		out, err := m.UEDSTTResidenceTime.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.UEDSTTResidenceTime.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIUEDSTTResidenceTime)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PortMgmtInfoCntr TLV-E, 8-65538B, IEI=0x74
	if m.PortMgmtInfoCntr != nil {
		out, err := m.PortMgmtInfoCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.PortMgmtInfoCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIPortMgmtInfoCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq) MarshalBinary() binary write PortMgmtInfoCntr")
		}
		writer.Write(out)
	}

	// m.EthHdrCompressionCfg TLV, 3B, IEI=0x1F
	if m.EthHdrCompressionCfg != nil {
		out, err := m.EthHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.EthHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIEthHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SuggestedIfId TLV, 11B, IEI=0x29
	if m.SuggestedIfId != nil {
		out, err := m.SuggestedIfId.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.SuggestedIfId.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEISuggestedIfId)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}

	// m.ReqMBSCntr TLV-E, 8-65538B, IEI=0x70
	if m.ReqMBSCntr != nil {
		out, err := m.ReqMBSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.ReqMBSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIReqMBSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq) MarshalBinary() binary write ReqMBSCntr")
		}
		writer.Write(out)
	}

	// m.PDUSessPairID TLV, 3B, IEI=0x34
	if m.PDUSessPairID != nil {
		out, err := m.PDUSessPairID.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.PDUSessPairID.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIPDUSessPairID)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RSN TLV, 3B, IEI=0x35
	if m.RSN != nil {
		out, err := m.RSN.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstReq.RSN.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstReqIEIRSN)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
