package message

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
)

var _ Message = &PDUSessEstAccept{}

// PDUSessEstAccept is detailed in 8.3.2 PDU session establishment accept, 24.501
type PDUSessEstAccept struct {
	PDUSessId            uint8
	PTI                  uint8
	SelectedPDUSessType  *ie.PDUSessType          //     V,     1/2B, 9.11.4.11
	SelectedSSCMode      *ie.SSCMode              //     V,     1/2B, 9.11.4.16
	AuthoQosRules        *ie.QosRules             //  LV-E, 6-65538B, 9.11.4.13
	SessAMBR             *ie.SessAMBR             //    LV,       7B, 9.11.4.14
	Cause5GSM            *ie.Cause5GSM            //    TV,       2B, 9.11.4.2
	PDUAddr              *ie.PDUAddr              //   TLV,    7-31B, 9.11.4.10
	RQTimerValue         *ie.GPRSTimer            //    TV,       2B, 9.11.2.3
	SNSSAI               *ie.SNSSAI               //   TLV,    3-10B, 9.11.2.8
	AlwaysonPDUSessInd   *ie.AlwaysonPDUSessInd   //    TV,       1B, 9.11.4.3
	MappedEPSBearerCtxs  *ie.MappedEPSBearerCtxs  // TLV-E, 7-65538B, 9.11.4.8
	EAPMsg               *ie.EAPMsg               // TLV-E,  7-1503B, 9.11.2.2
	AuthoQosFlowDescs    *ie.QosFlowDescs         // TLV-E, 6-65538B, 9.11.4.12
	ExtendedProtCfgOpts  *ie.ExtendedProtCfgOpts  // TLV-E, 4-65538B, 9.11.4.6
	DNN                  *ie.DNN                  //   TLV,   3-102B, 9.11.2.1B
	NwFeatureSupport5GSM *ie.NwFeatureSupport5GSM //   TLV,    3-15B, 9.11.4.18
	ServingPLMNRateCtrl  *ie.ServingPLMNRateCtrl  //   TLV,       4B, 9.11.4.20
	ATSSSCntr            *ie.ATSSSCntr            // TLV-E, 3-65538B, 9.11.4.22
	CtrlPlaneOnlyInd     *ie.CtrlPlaneOnlyInd     //    TV,       1B, 9.11.4.23
	IPHdrCompressionCfg  *ie.IPHdrCompressionCfg  //   TLV,   5-257B, 9.11.4.24
	EthHdrCompressionCfg *ie.EthHdrCompressionCfg //   TLV,       3B, 9.11.4.28
	SvcLvlAACntr         *ie.SvcLvlAACntr         // TLV-E, 4-65538B, 9.11.2.10
	ReceivedMBSCntr      *ie.ReceivedMBSCntr      // TLV-E, 9-65538B, 9.11.4.31
}

const (
	PDUSessEstAcceptIEICause5GSM            uint8 = 0x59
	PDUSessEstAcceptIEIPDUAddr              uint8 = 0x29
	PDUSessEstAcceptIEIRQTimerValue         uint8 = 0x56
	PDUSessEstAcceptIEISNSSAI               uint8 = 0x22
	PDUSessEstAcceptIEIAlwaysonPDUSessInd   uint8 = 0x80
	PDUSessEstAcceptIEIMappedEPSBearerCtxs  uint8 = 0x75
	PDUSessEstAcceptIEIEAPMsg               uint8 = 0x78
	PDUSessEstAcceptIEIAuthoQosFlowDescs    uint8 = 0x79
	PDUSessEstAcceptIEIExtendedProtCfgOpts  uint8 = 0x7B
	PDUSessEstAcceptIEIDNN                  uint8 = 0x25
	PDUSessEstAcceptIEINwFeatureSupport5GSM uint8 = 0x17
	PDUSessEstAcceptIEIServingPLMNRateCtrl  uint8 = 0x18
	PDUSessEstAcceptIEIATSSSCntr            uint8 = 0x77
	PDUSessEstAcceptIEICtrlPlaneOnlyInd     uint8 = 0xC0
	PDUSessEstAcceptIEIIPHdrCompressionCfg  uint8 = 0x66
	PDUSessEstAcceptIEIEthHdrCompressionCfg uint8 = 0x1F
	PDUSessEstAcceptIEISvcLvlAACntr         uint8 = 0x72
	PDUSessEstAcceptIEIReceivedMBSCntr      uint8 = 0x71
)

func (m *PDUSessEstAccept) ExtendedProtocolDiscriminator() Epd {
	return Epd5GSSessMgmtMsg
}

func (m *PDUSessEstAccept) MsgType() MsgType {
	return MsgTypePDUSessEstAccept
}

func (m *PDUSessEstAccept) PDUSessionID() uint8 {
	return m.PDUSessId
}

func (m *PDUSessEstAccept) ProcedureTransactionID() uint8 {
	return m.PTI
}

func (m *PDUSessEstAccept) UnmarshalBinary(b []byte) error {
	var ieLen uint16
	var err error
	var umerr Error

	if len(b) < int(GsmHdrLen) {
		return errors.Errorf("PDUSessEstAccept len(b)=%d, < GsmHdrLen(%d)",
			len(b), GsmHdrLen)
	}

	m.PDUSessId = b[1]
	m.PTI = b[2]

	// skip 4 bytes header (b should not be used after this call)
	reader := bytes.NewBuffer(b[4:])

	// Mandatory IE
	// 1/2B, handle with the other half
	half41, half85 := ie.GetHalfIEValue(reader.Next(1))
	m.SelectedPDUSessType = new(ie.PDUSessType) // V, 1/2B
	if err = m.SelectedPDUSessType.UnmarshalBinary(half41); err != nil {
		return errors.Wrap(err, "PDUSessEstAccept.SelectedPDUSessType.UnmarshalBinary")
	}
	m.SelectedSSCMode = new(ie.SSCMode) // V, 1/2B
	if err = m.SelectedSSCMode.UnmarshalBinary(half85); err != nil {
		return errors.Wrap(err, "PDUSessEstAccept.SelectedSSCMode.UnmarshalBinary")
	}

	m.AuthoQosRules = new(ie.QosRules) // LV-E, 6-65538B
	ieLen, err = getIeLen(reader, IELen16Bits)
	if err != nil {
		return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of AuthoQosRules")
	}
	if err = m.AuthoQosRules.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "PDUSessEstAccept.AuthoQosRules.UnmarshalBinary")
	}

	m.SessAMBR = new(ie.SessAMBR) // LV, 7B
	ieLen, err = getIeLen(reader, IELen8Bits)
	if err != nil {
		return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of SessAMBR")
	}
	if err = m.SessAMBR.UnmarshalBinary(
		reader.Next(int(ieLen))); err != nil {
		return errors.Wrap(err, "PDUSessEstAccept.SessAMBR.UnmarshalBinary")
	}

	// Optional / Conditional IE
	for reader.Len() > 0 {
		ieiByte := reader.Next(1)
		iei := extractIEI(ieiByte[0])
		switch iei {
		case PDUSessEstAcceptIEICause5GSM: // TV, 2B
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
				return errors.Wrap(err, "PDUSessEstAccept.Cause5GSM.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIPDUAddr: // TLV, 7-31B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of PDUAddr")
			}
			if m.PDUAddr != nil {
				reader.Next(int(ieLen))
				break
			}
			m.PDUAddr = new(ie.PDUAddr)
			if err = m.PDUAddr.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.PDUAddr = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.PDUAddr.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIRQTimerValue: // TV, 2B
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
				return errors.Wrap(err, "PDUSessEstAccept.RQTimerValue.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEISNSSAI: // TLV, 3-10B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of SNSSAI")
			}
			if m.SNSSAI != nil {
				reader.Next(int(ieLen))
				break
			}
			m.SNSSAI = new(ie.SNSSAI)
			if err = m.SNSSAI.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.SNSSAI = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.SNSSAI.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIAlwaysonPDUSessInd: // TV, 1B
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
				return errors.Wrap(err, "PDUSessEstAccept.AlwaysonPDUSessInd.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIMappedEPSBearerCtxs: // TLV-E, 7-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of MappedEPSBearerCtxs")
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
				return errors.Wrap(err, "PDUSessEstAccept.MappedEPSBearerCtxs.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIEAPMsg: // TLV-E, 7-1503B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of EAPMsg")
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
				return errors.Wrap(err, "PDUSessEstAccept.EAPMsg.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIAuthoQosFlowDescs: // TLV-E, 6-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of AuthoQosFlowDescs")
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
				return errors.Wrap(err, "PDUSessEstAccept.AuthoQosFlowDescs.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIExtendedProtCfgOpts: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of ExtendedProtCfgOpts")
			}
			if m.ExtendedProtCfgOpts != nil {
				reader.Next(int(ieLen))
				break
			}
			m.ExtendedProtCfgOpts = new(ie.ExtendedProtCfgOpts)
			// The PDU SESSION ESTABLISHMENT ACCEPT message is sent by the SMF to the UE
			if err = m.ExtendedProtCfgOpts.UnmarshalFromNw(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.ExtendedProtCfgOpts = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.ExtendedProtCfgOpts.UnmarshalFromNw")
			}
		case PDUSessEstAcceptIEIDNN: // TLV, 3-102B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of DNN")
			}
			if m.DNN != nil {
				reader.Next(int(ieLen))
				break
			}
			m.DNN = new(ie.DNN)
			if err = m.DNN.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.DNN = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.DNN.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEINwFeatureSupport5GSM: // TLV, 3-15B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of NwFeatureSupport5GSM")
			}
			if m.NwFeatureSupport5GSM != nil {
				reader.Next(int(ieLen))
				break
			}
			m.NwFeatureSupport5GSM = new(ie.NwFeatureSupport5GSM)
			if err = m.NwFeatureSupport5GSM.UnmarshalBinary(
				reader.Next(int(ieLen))); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.NwFeatureSupport5GSM = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.NwFeatureSupport5GSM.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIServingPLMNRateCtrl: // TLV, 4B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of ServingPLMNRateCtrl")
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
				return errors.Wrap(err, "PDUSessEstAccept.ServingPLMNRateCtrl.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIATSSSCntr: // TLV-E, 3-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of ATSSSCntr")
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
				return errors.Wrap(err, "PDUSessEstAccept.ATSSSCntr.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEICtrlPlaneOnlyInd: // TV, 1B
			if m.CtrlPlaneOnlyInd != nil {
				break
			}
			m.CtrlPlaneOnlyInd = new(ie.CtrlPlaneOnlyInd)
			if err = m.CtrlPlaneOnlyInd.UnmarshalBinary(
				ieiByte); err != nil {
				if _, ok := err.(*ie.IEToDo); ok {
					umerr.IEToDoList = append(umerr.IEToDoList, iei)
					m.CtrlPlaneOnlyInd = nil
					continue
				}
				return errors.Wrap(err, "PDUSessEstAccept.CtrlPlaneOnlyInd.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIIPHdrCompressionCfg: // TLV, 5-257B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of IPHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessEstAccept.IPHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIEthHdrCompressionCfg: // TLV, 3B
			ieLen, err = getIeLen(reader, IELen8Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of EthHdrCompressionCfg")
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
				return errors.Wrap(err, "PDUSessEstAccept.EthHdrCompressionCfg.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEISvcLvlAACntr: // TLV-E, 4-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of SvcLvlAACntr")
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
				return errors.Wrap(err, "PDUSessEstAccept.SvcLvlAACntr.UnmarshalBinary")
			}
		case PDUSessEstAcceptIEIReceivedMBSCntr: // TLV-E, 9-65538B
			ieLen, err = getIeLen(reader, IELen16Bits)
			if err != nil {
				return errors.Wrap(err, "PDUSessEstAccept UnmarshalBinary getIeLen of ReceivedMBSCntr")
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
				return errors.Wrap(err, "PDUSessEstAccept.ReceivedMBSCntr.UnmarshalBinary")
			}
		default:
			// Type and Length unknown, unable to parse.
			return errors.Errorf("PDUSessEstAccept unknown iei[%d]", iei)
		}
	}
	if len(umerr.IEToDoList) > 0 {
		return &umerr
	}
	return nil
}

func (m *PDUSessEstAccept) MarshalBinary() ([]byte, error) {
	if m.SelectedPDUSessType == nil || m.SelectedSSCMode == nil || m.AuthoQosRules == nil || m.SessAMBR == nil {
		return nil, errors.Errorf("SelectedPDUSessType=%v SelectedSSCMode=%v AuthoQosRules=%v"+
			" SessAMBR=%v must present in PDUSessEstAccept",
			m.SelectedPDUSessType, m.SelectedSSCMode, m.AuthoQosRules, m.SessAMBR)
	}

	writer := bytes.NewBuffer([]byte{
		byte(Epd5GSSessMgmtMsg),
		m.PDUSessId,
		m.PTI,
		byte(MsgTypePDUSessEstAccept),
	})

	tmp := [1]byte{}
	// selectedpdusesstype, V, 1/2B
	selectedpdusesstype, err := m.SelectedPDUSessType.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstAccept.SelectedPDUSessType.MarshalBinary()")
	}

	// selectedsscmode, V, 1/2B
	selectedsscmode, err := m.SelectedSSCMode.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstAccept.SelectedSSCMode.MarshalBinary()")
	}
	tmp[0] = ie.SetHalfValue(selectedsscmode[0], selectedpdusesstype[0])
	writer.Write(tmp[:])

	// authoqosrules, LV-E, 6-65538B
	authoqosrules, err := m.AuthoQosRules.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstAccept.AuthoQosRules.MarshalBinary()")
	}
	if err = binary.Write(writer, binary.BigEndian, uint16(len(authoqosrules))); err != nil {
		return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write AuthoQosRules")
	}
	writer.Write(authoqosrules)

	// sessambr, LV, 7B
	sessambr, err := m.SessAMBR.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "PDUSessEstAccept.SessAMBR.MarshalBinary()")
	}
	writer.WriteByte(byte(len(sessambr)))
	writer.Write(sessambr)

	// m.Cause5GSM TV, 2B, IEI=0x59
	if m.Cause5GSM != nil {
		out, err := m.Cause5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.Cause5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEICause5GSM)
		writer.Write(out)
	}

	// m.PDUAddr TLV, 7-31B, IEI=0x29
	if m.PDUAddr != nil {
		out, err := m.PDUAddr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.PDUAddr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIPDUAddr)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.RQTimerValue TV, 2B, IEI=0x56
	if m.RQTimerValue != nil {
		out, err := m.RQTimerValue.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.RQTimerValue.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIRQTimerValue)
		writer.Write(out)
	}

	// m.SNSSAI TLV, 3-10B, IEI=0x22
	if m.SNSSAI != nil {
		out, err := m.SNSSAI.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.SNSSAI.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEISNSSAI)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.AlwaysonPDUSessInd TV, 1B, IEI=0x80, >= 0x80 !
	if m.AlwaysonPDUSessInd != nil {
		out, err := m.AlwaysonPDUSessInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.AlwaysonPDUSessInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstAcceptIEIAlwaysonPDUSessInd)
	}

	// m.MappedEPSBearerCtxs TLV-E, 7-65538B, IEI=0x75
	if m.MappedEPSBearerCtxs != nil {
		out, err := m.MappedEPSBearerCtxs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.MappedEPSBearerCtxs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIMappedEPSBearerCtxs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write MappedEPSBearerCtxs")
		}
		writer.Write(out)
	}

	// m.EAPMsg TLV-E, 7-1503B, IEI=0x78
	if m.EAPMsg != nil {
		out, err := m.EAPMsg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.EAPMsg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIEAPMsg)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write EAPMsg")
		}
		writer.Write(out)
	}

	// m.AuthoQosFlowDescs TLV-E, 6-65538B, IEI=0x79
	if m.AuthoQosFlowDescs != nil {
		out, err := m.AuthoQosFlowDescs.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.AuthoQosFlowDescs.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIAuthoQosFlowDescs)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write AuthoQosFlowDescs")
		}
		writer.Write(out)
	}

	// m.ExtendedProtCfgOpts TLV-E, 4-65538B, IEI=0x7B
	if m.ExtendedProtCfgOpts != nil {
		out, err := m.ExtendedProtCfgOpts.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.ExtendedProtCfgOpts.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIExtendedProtCfgOpts)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write ExtendedProtCfgOpts")
		}
		writer.Write(out)
	}

	// m.DNN TLV, 3-102B, IEI=0x25
	if m.DNN != nil {
		out, err := m.DNN.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.DNN.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIDNN)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.NwFeatureSupport5GSM TLV, 3-15B, IEI=0x17
	if m.NwFeatureSupport5GSM != nil {
		out, err := m.NwFeatureSupport5GSM.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.NwFeatureSupport5GSM.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEINwFeatureSupport5GSM)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ServingPLMNRateCtrl TLV, 4B, IEI=0x18
	if m.ServingPLMNRateCtrl != nil {
		out, err := m.ServingPLMNRateCtrl.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.ServingPLMNRateCtrl.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIServingPLMNRateCtrl)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.ATSSSCntr TLV-E, 3-65538B, IEI=0x77
	if m.ATSSSCntr != nil {
		out, err := m.ATSSSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.ATSSSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIATSSSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write ATSSSCntr")
		}
		writer.Write(out)
	}

	// m.CtrlPlaneOnlyInd TV, 1B, IEI=0xC0, >= 0x80 !
	if m.CtrlPlaneOnlyInd != nil {
		out, err := m.CtrlPlaneOnlyInd.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.CtrlPlaneOnlyInd.MarshalBinary()")
		}
		writer.WriteByte(out[0] | PDUSessEstAcceptIEICtrlPlaneOnlyInd)
	}

	// m.IPHdrCompressionCfg TLV, 5-257B, IEI=0x66
	if m.IPHdrCompressionCfg != nil {
		out, err := m.IPHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.IPHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIIPHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.EthHdrCompressionCfg TLV, 3B, IEI=0x1F
	if m.EthHdrCompressionCfg != nil {
		out, err := m.EthHdrCompressionCfg.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.EthHdrCompressionCfg.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIEthHdrCompressionCfg)
		writer.WriteByte(byte(len(out)))
		writer.Write(out)
	}

	// m.SvcLvlAACntr TLV-E, 4-65538B, IEI=0x72
	if m.SvcLvlAACntr != nil {
		out, err := m.SvcLvlAACntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.SvcLvlAACntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEISvcLvlAACntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write SvcLvlAACntr")
		}
		writer.Write(out)
	}

	// m.ReceivedMBSCntr TLV-E, 9-65538B, IEI=0x71
	if m.ReceivedMBSCntr != nil {
		out, err := m.ReceivedMBSCntr.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept.ReceivedMBSCntr.MarshalBinary()")
		}
		writer.WriteByte(PDUSessEstAcceptIEIReceivedMBSCntr)
		if err = binary.Write(writer, binary.BigEndian, uint16(len(out))); err != nil {
			return nil, errors.Wrap(err, "PDUSessEstAccept) MarshalBinary() binary write ReceivedMBSCntr")
		}
		writer.Write(out)
	}
	return writer.Bytes(), nil
}
