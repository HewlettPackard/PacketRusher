package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessModCmd{}

// PDUSessModCmd is detailed in 8.3.9 PDU session modification command, 24.501
type PDUSessModCmd struct {
	PDUSessId            uint8
	PTI                  uint8
	Cause5GSM            *ie.Cause5GSM            //    TV,       2B, 9.11.4.2
	SessAMBR             *ie.SessAMBR             //   TLV,       8B, 9.11.4.14
	RQTimerValue         *ie.GPRSTimer            //    TV,       2B, 9.11.2.3
	AlwaysonPDUSessInd   *ie.AlwaysonPDUSessInd   //    TV,       1B, 9.11.4.3
	AuthoQosRules        *ie.QosRules             // TLV-E, 7-65538B, 9.11.4.13
	MappedEPSBearerCtxs  *ie.MappedEPSBearerCtxs  // TLV-E, 7-65538B, 9.11.4.8
	AuthoQosFlowDescs    *ie.QosFlowDescs         // TLV-E, 6-65538B, 9.11.4.12
	ExtendedProtCfgOpts  *ie.ExtendedProtCfgOpts  // TLV-E, 4-65538B, 9.11.4.6
	ATSSSCntr            *ie.ATSSSCntr            // TLV-E, 3-65538B, 9.11.4.22
	IPHdrCompressionCfg  *ie.IPHdrCompressionCfg  //   TLV,   5-257B, 9.11.4.24
	PortMgmtInfoCntr     *ie.PortMgmtInfoCntr     // TLV-E, 4-65538B, 9.11.4.27
	ServingPLMNRateCtrl  *ie.ServingPLMNRateCtrl  //   TLV,       4B, 9.11.4.20
	EthHdrCompressionCfg *ie.EthHdrCompressionCfg //   TLV,       3B, 9.11.4.28
	ReceivedMBSCntr      *ie.ReceivedMBSCntr      // TLV-E, 9-65538B, 9.11.4.31
	SvcLvlAACntr         *ie.SvcLvlAACntr         // TLV-E, 4-65538B, 9.11.2.10
}

const (
	PDUSessModCmdIEICause5GSM            uint8 = 0x59
	PDUSessModCmdIEISessAMBR             uint8 = 0x2A
	PDUSessModCmdIEIRQTimerValue         uint8 = 0x56
	PDUSessModCmdIEIAlwaysonPDUSessInd   uint8 = 0x80
	PDUSessModCmdIEIAuthoQosRules        uint8 = 0x7A
	PDUSessModCmdIEIMappedEPSBearerCtxs  uint8 = 0x75
	PDUSessModCmdIEIAuthoQosFlowDescs    uint8 = 0x79
	PDUSessModCmdIEIExtendedProtCfgOpts  uint8 = 0x7B
	PDUSessModCmdIEIATSSSCntr            uint8 = 0x77
	PDUSessModCmdIEIIPHdrCompressionCfg  uint8 = 0x66
	PDUSessModCmdIEIPortMgmtInfoCntr     uint8 = 0x74
	PDUSessModCmdIEIServingPLMNRateCtrl  uint8 = 0x1E
	PDUSessModCmdIEIEthHdrCompressionCfg uint8 = 0x1F
	PDUSessModCmdIEIReceivedMBSCntr      uint8 = 0x71
	PDUSessModCmdIEISvcLvlAACntr         uint8 = 0x72
)

func (m *PDUSessModCmd) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessModCmd) MsgType() MsgType {
	return MsgTypePDUSessModCmd
}

func (m *PDUSessModCmd) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessModCmd) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessModCmd) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessModCmd len(b)=%d, < GsmHdrLen(%d)",
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
		case PDUSessModCmdIEICause5GSM: // TV, 2B
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
				return errors.Wrap(err, "PDUSessModCmd.Cause5GSM.UnmarshalBinary")
			}
		case PDUSessModCmdIEISessAMBR: // TLV, 8B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of SessAMBR")
			}
			if m.SessAMBR != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SessAMBR = new(ie.SessAMBR)
			if err = m.SessAMBR.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SessAMBR = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.SessAMBR.UnmarshalBinary")
			}
		case PDUSessModCmdIEIRQTimerValue: // TV, 2B
			ieLen = 1
			if m.RQTimerValue != nil {
				reader.Next(int(ieLen))
				break
			}
			m.RQTimerValue = new(ie.GPRSTimer)
			if err = m.RQTimerValue.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.RQTimerValue = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.RQTimerValue.UnmarshalBinary")
			}
		case PDUSessModCmdIEIAlwaysonPDUSessInd: // TV, 1B
			if m.AlwaysonPDUSessInd != nil {
				break
			}
			m.AlwaysonPDUSessInd = new(ie.AlwaysonPDUSessInd)
			if err = m.AlwaysonPDUSessInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AlwaysonPDUSessInd = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.AlwaysonPDUSessInd.UnmarshalBinary")
			}
		case PDUSessModCmdIEIAuthoQosRules: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of AuthoQosRules")
			}
			if m.AuthoQosRules != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthoQosRules = new(ie.QosRules)
			if err = m.AuthoQosRules.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthoQosRules = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.AuthoQosRules.UnmarshalBinary")
			}
		case PDUSessModCmdIEIMappedEPSBearerCtxs: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of MappedEPSBearerCtxs")
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
				return errors.Wrap(err, "PDUSessModCmd.MappedEPSBearerCtxs.UnmarshalBinary")
			}
		case PDUSessModCmdIEIAuthoQosFlowDescs: // TLV-E, 6-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of AuthoQosFlowDescs")
			}
			if m.AuthoQosFlowDescs != nil {
				reader.Next(int(ieLen))
				break
			}
			m.AuthoQosFlowDescs = new(ie.QosFlowDescs)
			if err = m.AuthoQosFlowDescs.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.AuthoQosFlowDescs = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.AuthoQosFlowDescs.UnmarshalBinary")
			}
		case PDUSessModCmdIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// PDU SESSION MODIFICATION COMMAND is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		case PDUSessModCmdIEIATSSSCntr: // TLV-E, 3-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of ATSSSCntr")
			}
			if m.ATSSSCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ATSSSCntr = new(ie.ATSSSCntr)
			if err = m.ATSSSCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ATSSSCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.ATSSSCntr.UnmarshalBinary")
			}
		case PDUSessModCmdIEIIPHdrCompressionCfg: // TLV, 5-257B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of IPHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessModCmd.IPHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessModCmdIEIPortMgmtInfoCntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of PortMgmtInfoCntr")
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
				return errors.Wrap(err, "PDUSessModCmd.PortMgmtInfoCntr.UnmarshalBinary")
			}
		case PDUSessModCmdIEIServingPLMNRateCtrl: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of ServingPLMNRateCtrl")
			}
			if m.ServingPLMNRateCtrl != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ServingPLMNRateCtrl = new(ie.ServingPLMNRateCtrl)
			if err = m.ServingPLMNRateCtrl.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ServingPLMNRateCtrl = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.ServingPLMNRateCtrl.UnmarshalBinary")
			}
		case PDUSessModCmdIEIEthHdrCompressionCfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of EthHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessModCmd.EthHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessModCmdIEIReceivedMBSCntr: // TLV-E, 9-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of ReceivedMBSCntr")
			}
			if m.ReceivedMBSCntr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ReceivedMBSCntr = new(ie.ReceivedMBSCntr)
			if err = m.ReceivedMBSCntr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ReceivedMBSCntr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessModCmd.ReceivedMBSCntr.UnmarshalBinary")
			}
		case PDUSessModCmdIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessModCmd UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "PDUSessModCmd.SvcLvlAACntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessModCmd unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessModCmd) MarshalBinary() ([]byte, error) {
	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessModCmd),
	})

	// m.Cause5GSM TV, 2B, IEI=0x59
	if m.Cause5GSM != nil {
		out, err := m.Cause5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.Cause5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEICause5GSM)
		writer.Write(out)
	}

	// m.SessAMBR TLV, 8B, IEI=0x2A
	if m.SessAMBR != nil {
		out, err := m.SessAMBR.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.SessAMBR.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEISessAMBR)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RQTimerValue TV, 2B, IEI=0x56
	if m.RQTimerValue != nil {
		out, err := m.RQTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.RQTimerValue.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIRQTimerValue)
		writer.Write(out)
	}

	// m.AlwaysonPDUSessInd TV, 1B, IEI=0x80, >= 0x80 !
	if m.AlwaysonPDUSessInd != nil {
		out, err := m.AlwaysonPDUSessInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.AlwaysonPDUSessInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessModCmdIEIAlwaysonPDUSessInd)
	}

	// m.AuthoQosRules TLV-E, 7-65538B, IEI=0x7A
	if m.AuthoQosRules != nil {
		out, err := m.AuthoQosRules.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.AuthoQosRules.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIAuthoQosRules)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write AuthoQosRules")
		}
		writer.Write(out)
	}

	// m.MappedEPSBearerCtxs TLV-E, 7-65538B, IEI=0x75
	if m.MappedEPSBearerCtxs != nil {
		out, err := m.MappedEPSBearerCtxs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.MappedEPSBearerCtxs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIMappedEPSBearerCtxs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write MappedEPSBearerCtxs")
		}
		writer.Write(out)
	}

	// m.AuthoQosFlowDescs TLV-E, 6-65538B, IEI=0x79
	if m.AuthoQosFlowDescs != nil {
		out, err := m.AuthoQosFlowDescs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.AuthoQosFlowDescs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIAuthoQosFlowDescs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write AuthoQosFlowDescs")
		}
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.ATSSSCntr TLV-E, 3-65538B, IEI=0x77
	if m.ATSSSCntr != nil {
		out, err := m.ATSSSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.ATSSSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIATSSSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write ATSSSCntr")
		}
		writer.Write(out)
	}

	// m.IPHdrCompressionCfg TLV, 5-257B, IEI=0x66
	if m.IPHdrCompressionCfg != nil {
		out, err := m.IPHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.IPHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIIPHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.PortMgmtInfoCntr TLV-E, 4-65538B, IEI=0x74
	if m.PortMgmtInfoCntr != nil {
		out, err := m.PortMgmtInfoCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.PortMgmtInfoCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIPortMgmtInfoCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write PortMgmtInfoCntr")
		}
		writer.Write(out)
	}

	// m.ServingPLMNRateCtrl TLV, 4B, IEI=0x1E
	if m.ServingPLMNRateCtrl != nil {
		out, err := m.ServingPLMNRateCtrl.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.ServingPLMNRateCtrl.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIServingPLMNRateCtrl)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EthHdrCompressionCfg TLV, 3B, IEI=0x1F
	if m.EthHdrCompressionCfg != nil {
		out, err := m.EthHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.EthHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIEthHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ReceivedMBSCntr TLV-E, 9-65538B, IEI=0x71
	if m.ReceivedMBSCntr != nil {
		out, err := m.ReceivedMBSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.ReceivedMBSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEIReceivedMBSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write ReceivedMBSCntr")
		}
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessModCmdIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessModCmd) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
