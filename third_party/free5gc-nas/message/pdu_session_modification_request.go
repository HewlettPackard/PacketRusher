package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessModReq{}

// PDUSessModReq is detailed in 8.3.7 PDU session modification request, 24.501
type PDUSessModReq struct {
	PDUSessId                      uint8
	PTI                            uint8
	Capability5GSM                 *ie.Capability5GSM                 //   TLV,    3-15B, 9.11.4.1
	Cause5GSM                      *ie.Cause5GSM                      //    TV,       2B, 9.11.4.2
	MaxNumOfSupportedPktFilters    *ie.MaxNumOfSupportedPktFilters    //    TV,       3B, 9.11.4.9
	AlwaysonPDUSessReq             *ie.AlwaysonPDUSessReq             //    TV,       1B, 9.11.4.4
	IntegrityProtectionMaxDataRate *ie.IntegrityProtectionMaxDataRate //    TV,       3B, 9.11.4.7
	ReqQosRules                    *ie.QosRules                       // TLV-E, 7-65538B, 9.11.4.13
	ReqQosFlowDescs                *ie.QosFlowDescs                   // TLV-E, 6-65538B, 9.11.4.12
	MappedEPSBearerCtxs            *ie.MappedEPSBearerCtxs            // TLV-E, 7-65538B, 9.11.4.8
	ExtendedProtCfgOpts            *ie.ExtendedProtCfgOpts            // TLV-E, 4-65538B, 9.11.4.6
	PortMgmtInfoCntr               *ie.PortMgmtInfoCntr               // TLV-E, 4-65538B, 9.11.4.27
	IPHdrCompressionCfg            *ie.IPHdrCompressionCfg            //   TLV,   5-257B, 9.11.4.24
	EthHdrCompressionCfg           *ie.EthHdrCompressionCfg           //   TLV,       3B, 9.11.4.28
	ReqMBSCntr                     *ie.ReqMBSCntr                     // TLV-E, 8-65538B, 9.11.4.30
	SvcLvlAACntr                   *ie.SvcLvlAACntr                   // TLV-E, 4-65538B, 9.11.2.10
}

const (
	PDUSessModReqIEICapability5GSM                 uint8 = 0x28
	PDUSessModReqIEICause5GSM                      uint8 = 0x59
	PDUSessModReqIEIMaxNumOfSupportedPktFilters    uint8 = 0x55
	PDUSessModReqIEIAlwaysonPDUSessReq             uint8 = 0xB0
	PDUSessModReqIEIIntegrityProtectionMaxDataRate uint8 = 0x13
	PDUSessModReqIEIReqQosRules                    uint8 = 0x7A
	PDUSessModReqIEIReqQosFlowDescs                uint8 = 0x79
	PDUSessModReqIEIMappedEPSBearerCtxs            uint8 = 0x75
	PDUSessModReqIEIExtendedProtCfgOpts            uint8 = 0x7B
	PDUSessModReqIEIPortMgmtInfoCntr               uint8 = 0x74
	PDUSessModReqIEIIPHdrCompressionCfg            uint8 = 0x66
	PDUSessModReqIEIEthHdrCompressionCfg           uint8 = 0x1F
	PDUSessModReqIEIReqMBSCntr                     uint8 = 0x70
	PDUSessModReqIEISvcLvlAACntr                   uint8 = 0x72
)

func (m *PDUSessModReq) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessModReq) MsgType() MsgType {
	return MsgTypePDUSessModReq
}

func (m *PDUSessModReq) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessModReq) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessModReq) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessModReq len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// This message contains 0 Mandatory IE
	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessModReqIEICapability5GSM: // TLV, 3-15B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of Capability5GSM")
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
				return errors.Wrap(err, "PDUSessModReq.Capability5GSM.UnmarshalBinary")
			}
		case PDUSessModReqIEICause5GSM: // TV, 2B
			ieLen = 1
			if m.Cause5GSM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.Cause5GSM = new(ie.Cause5GSM)
			if err = m.Cause5GSM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.Cause5GSM = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.Cause5GSM.UnmarshalBinary")
			}
		case PDUSessModReqIEIMaxNumOfSupportedPktFilters: // TV, 3B
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
				return errors.Wrap(err, "PDUSessModReq.MaxNumOfSupportedPktFilters.UnmarshalBinary")
			}
		case PDUSessModReqIEIAlwaysonPDUSessReq: // TV, 1B
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
				return errors.Wrap(err, "PDUSessModReq.AlwaysonPDUSessReq.UnmarshalBinary")
			}
		case PDUSessModReqIEIIntegrityProtectionMaxDataRate: // TV, 3B
			ieLen = 2
			if m.IntegrityProtectionMaxDataRate != nil {
				reader.Next(int(ieLen))
				break
			}
			m.IntegrityProtectionMaxDataRate = new(ie.IntegrityProtectionMaxDataRate)
			if err = m.IntegrityProtectionMaxDataRate.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.IntegrityProtectionMaxDataRate = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.IntegrityProtectionMaxDataRate.UnmarshalBinary")
			}
		case PDUSessModReqIEIReqQosRules: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of ReqQosRules")
			}
			if m.ReqQosRules != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqQosRules = new(ie.QosRules)
			if err = m.ReqQosRules.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqQosRules = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.ReqQosRules.UnmarshalBinary")
			}
		case PDUSessModReqIEIReqQosFlowDescs: // TLV-E, 6-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of ReqQosFlowDescs")
			}
			if m.ReqQosFlowDescs != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReqQosFlowDescs = new(ie.QosFlowDescs)
			if err = m.ReqQosFlowDescs.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReqQosFlowDescs = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.ReqQosFlowDescs.UnmarshalBinary")
			}
		case PDUSessModReqIEIMappedEPSBearerCtxs: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of MappedEPSBearerCtxs")
			}
			if m.MappedEPSBearerCtxs != nil {
				reader.Next(int(ieLen))
				break
			}
			m.MappedEPSBearerCtxs = new(ie.MappedEPSBearerCtxs)
			if err = m.MappedEPSBearerCtxs.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.MappedEPSBearerCtxs = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.MappedEPSBearerCtxs.UnmarshalBinary")
			}
		case PDUSessModReqIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION MODIFICATION REQUEST is sent by the UE to the SMF
			if err = m.ExtendedProtCfgOpts.UnmarshalFromMs(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModReq.ExtendedProtCfgOpts.UnmarshalFromMs")
			}
		case PDUSessModReqIEIPortMgmtInfoCntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of PortMgmtInfoCntr")
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
				return errors.Wrap(err, "PDUSessModReq.PortMgmtInfoCntr.UnmarshalBinary")
			}
		case PDUSessModReqIEIIPHdrCompressionCfg: // TLV, 5-257B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of IPHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessModReq.IPHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessModReqIEIEthHdrCompressionCfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of EthHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessModReq.EthHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessModReqIEIReqMBSCntr: // TLV-E, 8-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of ReqMBSCntr")
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
				return errors.Wrap(err, "PDUSessModReq.ReqMBSCntr.UnmarshalBinary")
			}
		case PDUSessModReqIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModReq UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "PDUSessModReq.SvcLvlAACntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessModReq unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessModReq) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessModReq),
	})

	// m.Capability5GSM TLV, 3-15B, IEI=0x28
	if m.Capability5GSM != nil {
		out, err := m.Capability5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.Capability5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEICapability5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.Cause5GSM TV, 2B, IEI=0x59
	if m.Cause5GSM != nil {
		out, err := m.Cause5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.Cause5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEICause5GSM)
		writer.Write(out)
	}

	// m.MaxNumOfSupportedPktFilters TV, 3B, IEI=0x55
	if m.MaxNumOfSupportedPktFilters != nil {
		out, err := m.MaxNumOfSupportedPktFilters.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.MaxNumOfSupportedPktFilters.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIMaxNumOfSupportedPktFilters)
		writer.Write(out)
	}

	// m.AlwaysonPDUSessReq TV, 1B, IEI=0xB0, >= 0x80 !
	if m.AlwaysonPDUSessReq != nil {
		out, err := m.AlwaysonPDUSessReq.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.AlwaysonPDUSessReq.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessModReqIEIAlwaysonPDUSessReq)
	}

	// m.IntegrityProtectionMaxDataRate TV, 3B, IEI=0x13
	if m.IntegrityProtectionMaxDataRate != nil {
		out, err := m.IntegrityProtectionMaxDataRate.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.IntegrityProtectionMaxDataRate.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIIntegrityProtectionMaxDataRate)
		writer.Write(out)
	}

	// m.ReqQosRules TLV-E, 7-65538B, IEI=0x7A
	if m.ReqQosRules != nil {
		out, err := m.ReqQosRules.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.ReqQosRules.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIReqQosRules)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write ReqQosRules")
		}
		writer.Write(out)
	}

	// m.ReqQosFlowDescs TLV-E, 6-65538B, IEI=0x79
	if m.ReqQosFlowDescs != nil {
		out, err := m.ReqQosFlowDescs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.ReqQosFlowDescs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIReqQosFlowDescs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write ReqQosFlowDescs")
		}
		writer.Write(out)
	}

	// m.MappedEPSBearerCtxs TLV-E, 7-65538B, IEI=0x75
	if m.MappedEPSBearerCtxs != nil {
		out, err := m.MappedEPSBearerCtxs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.MappedEPSBearerCtxs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIMappedEPSBearerCtxs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write MappedEPSBearerCtxs")
		}
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.PortMgmtInfoCntr TLV-E, 4-65538B, IEI=0x74
	if m.PortMgmtInfoCntr != nil {
		out, err := m.PortMgmtInfoCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.PortMgmtInfoCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIPortMgmtInfoCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write PortMgmtInfoCntr")
		}
		writer.Write(out)
	}

	// m.IPHdrCompressionCfg TLV, 5-257B, IEI=0x66
	if m.IPHdrCompressionCfg != nil {
		out, err := m.IPHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.IPHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIIPHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EthHdrCompressionCfg TLV, 3B, IEI=0x1F
	if m.EthHdrCompressionCfg != nil {
		out, err := m.EthHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.EthHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIEthHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReqMBSCntr TLV-E, 8-65538B, IEI=0x70
	if m.ReqMBSCntr != nil {
		out, err := m.ReqMBSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.ReqMBSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEIReqMBSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write ReqMBSCntr")
		}
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModReqIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModReq) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
