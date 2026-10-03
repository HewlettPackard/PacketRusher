package ie

import (
	"encoding/binary"
	"net"

	"github.com/pkg/errors"
)

type UEStatus3GPPPSDataOff uint8

const (
	UEStatus3GPPPSDataOff_NotPresent UEStatus3GPPPSDataOff = iota
	UEStatus3GPPPSDataOff_Deactivate
	UEStatus3GPPPSDataOff_Activate
)

// MS --> Network
type ExtCfgOptFromMs struct {
	IPv4LinkMTUReq        bool
	DNSV4Req              bool
	DNSV6Req              bool
	P_CSCF_IPv4AddrReq    bool
	UEStatus3GPPPSDataOff UEStatus3GPPPSDataOff
}

// Network --> MS
type ExtCfgOptFromNw struct {
	IPv4LinkMTU                    uint16
	DNSIPv6Addr                    net.IP
	DNSIPv4Addr                    net.IP
	P_CSCF_IPv4Addr                net.IP
	Indication3GPPPSDataOffSupport bool
}

// ExtendedProtCfgOpts is detailed in 10.5.6.3A Extended Prot configuration options, 24.00
type ExtendedProtCfgOpts struct {
	CfgProt CfgProtType
	FromMs  *ExtCfgOptFromMs
	FromNw  *ExtCfgOptFromNw
}

type CfgProtType uint8

const (
	PPP_IP_PDP_or_IP_PDN CfgProtType = 0x00
)

// TS 24.008 10.5.6.3, MS -> Nw
const (
	PCSCF_IPv6AddrReqUL                                      uint16 = 0x0001
	IMCNSubsystemSignalingFlagUL                             uint16 = 0x0002
	DNSServerIPv6AddrReqUL                                   uint16 = 0x0003
	NotSupportedUL                                           uint16 = 0x0004
	MSSupportOfNwReqedBearerCtrlIndicatorUL                  uint16 = 0x0005
	DSMIPv6HomeAgentAddrReqUL                                uint16 = 0x0007
	DSMIPv6HomeNwPrefixReqUL                                 uint16 = 0x0008
	DSMIPv6IPv4HomeAgentAddrReqUL                            uint16 = 0x0009
	IPAddrAllocationViaNASSignallingUL                       uint16 = 0x000a
	IPv4AddrAllocationViaDHCPv4UL                            uint16 = 0x000b
	PCSCF_IPv4AddrReqUL                                      uint16 = 0x000c
	DNSServerIPv4AddrReqUL                                   uint16 = 0x000d
	MSISDNReqUL                                              uint16 = 0x000e
	IFOMSupportReqUL                                         uint16 = 0x000f
	IPv4LinkMTUReqUL                                         uint16 = 0x0010
	MSSupportOfLocalAddrInTFTIndicatorUL                     uint16 = 0x0011
	PCSCF_ReSelectionSupportUL                               uint16 = 0x0012
	NBIFOMReqIndicatorUL                                     uint16 = 0x0013
	NBIFOMModeUL                                             uint16 = 0x0014
	NonIPLinkMTUReqUL                                        uint16 = 0x0015
	APNRateCtrl_SupportIndicatorUL                           uint16 = 0x0016
	UEStatus3GPPPSDataOffUL                                  uint16 = 0x0017
	ReliableDataSvcReqIndicatorUL                            uint16 = 0x0018
	AdditionalAPNRateCtrlForExceptionData_SupportIndicatorUL uint16 = 0x0019
	PDUSessIDUL                                              uint16 = 0x001a
	EthFramePayloadMTUReqUL                                  uint16 = 0x0020
	UnstructuredLinkMTUReqUL                                 uint16 = 0x0021
	I5GSMCauseValUL                                          uint16 = 0x0022 // 5GSMCauseValUL
	QoSRulesWithTheLenOf2Octets_SupportIndicatorUL           uint16 = 0x0023
	QoSFlowDescWithTheLenOf2Octets_SupportIndicatorUL        uint16 = 0x0024
	ACSInfoReqUL                                             uint16 = 0x0027
	ATSSSReqUL                                               uint16 = 0x0030
	DNSServerSecInfoIndicatorUL                              uint16 = 0x0031
	LinkCtrlProtUL                                           uint16 = 0xc021
	PushAccessCtrlProtUL                                     uint16 = 0xc023
	ChlgHandshakeAuthProtUL                                  uint16 = 0xc223
	InternetProtCtrlProtUL                                   uint16 = 0x8021
)

// TS 24.008 10.5.6.3, Nw -> MS
const (
	PCSCF_IPv6AddrDL                                       uint16 = 0x0001
	IMCNSubsystemSignalingFlagDL                           uint16 = 0x0002
	DNSServerIPv6AddrDL                                    uint16 = 0x0003
	PolicyCtrlRejCodeDL                                    uint16 = 0x0004
	SelectedBearerCtrlModeDL                               uint16 = 0x0005
	DSMIPv6HomeAgentAddrDL                                 uint16 = 0x0007
	DSMIPv6HomeNwPrefixDL                                  uint16 = 0x0008
	DSMIPv6IPv4HomeAgentAddrDL                             uint16 = 0x0009
	PCSCF_IPv4AddrDL                                       uint16 = 0x000c
	DNSServerIPv4AddrDL                                    uint16 = 0x000d
	MSISDNDL                                               uint16 = 0x000e
	IFOMSupportDL                                          uint16 = 0x000f
	IPv4LinkMTUDL                                          uint16 = 0x0010
	NwSupportOfLocalAddrInTFTIndicatorDL                   uint16 = 0x0011
	NBIFOMAcceptedIndicatorDL                              uint16 = 0x0013
	NBIFOMModeDL                                           uint16 = 0x0014
	NonIPLinkMTUDL                                         uint16 = 0x0015
	APNRateCtrlParamDL                                     uint16 = 0x0016
	Indication3GPPPSDataOffSupportDL                       uint16 = 0x0017
	ReliableDataSvcAcceptedIndicatorDL                     uint16 = 0x0018
	AdditionalAPNRateCtrlForExceptionDataParamDL           uint16 = 0x0019
	SNSSAIDL                                               uint16 = 0x001b
	QoSRulesDL                                             uint16 = 0x001c
	SessAMBRDL                                             uint16 = 0x001d
	PDUSessAddrLifetimeDL                                  uint16 = 0x001e
	QoSFlowDescDL                                          uint16 = 0x001f
	EthFramePayloadMTUDL                                   uint16 = 0x0020
	UnstructuredLinkMTUDL                                  uint16 = 0x0021
	QoSRulesWithTheLenOf2OctetsDL                          uint16 = 0x0023
	QoSFlowDescWithTheLenOf2OctetsDL                       uint16 = 0x0024
	SmallDataRateCtrlParamDL                               uint16 = 0x0025
	AdditionalSmallDataRateCtrlForExceptionDataParamDL     uint16 = 0x0026
	ACSInfoDL                                              uint16 = 0x0027
	InitSmallDataRateCtrlParamDL                           uint16 = 0x0028
	InitAdditionalSmallDataRateCtrlForExceptionDataParamDL uint16 = 0x0029
	InitAPNRateCtrlParamDL                                 uint16 = 0x002A
	InitAdditionalAPNRateCtrlForExceptionDataParamDL       uint16 = 0x002B
	ATSSSRspWithTheLenOf2OctetsDL                          uint16 = 0x0030
	DNSServerSecInfoWithLenOf2OctetsDL                     uint16 = 0x0031
)

const protIdLen int = 2

// UnmarshalBinary handles the Val part, not including IEI and Len.
func (i *ExtendedProtCfgOpts) UnmarshalFromMs(b []byte) error {
	Len := len(b)
	if Len < 1 {
		return errors.Errorf("ExtendedProtCfgOpts: UnmarshalFromMs() bad total length (%d<1)", Len)
	}

	// All other values are interpreted as PPP in this version of the protocol.
	i.CfgProt = PPP_IP_PDP_or_IP_PDN
	ofs := 1

	i.FromMs = &ExtCfgOptFromMs{}
	Len_idLen := 2 + 1

	for ofs+Len_idLen <= Len {
		protId := binary.BigEndian.Uint16(b[ofs : ofs+2])
		ofs += 2
		cntLen := b[ofs]
		ofs++
		if ofs+int(cntLen) > Len {
			i.FromMs = &ExtCfgOptFromMs{}
			return errors.Errorf("ExtendedProtCfgOpts: UnmarshalFromMs() bad length, ofs=%d, cntLen=%d, total Len=%d",
				ofs, cntLen, Len)
		}

		switch protId {
		case PCSCF_IPv6AddrReqUL:
		case IMCNSubsystemSignalingFlagUL:
		case DNSServerIPv6AddrReqUL:
			i.FromMs.DNSV6Req = true
		case NotSupportedUL:
		case MSSupportOfNwReqedBearerCtrlIndicatorUL:
		case DSMIPv6HomeAgentAddrReqUL:
		case DSMIPv6HomeNwPrefixReqUL:
		case DSMIPv6IPv4HomeAgentAddrReqUL:
		case IPAddrAllocationViaNASSignallingUL:
		case IPv4AddrAllocationViaDHCPv4UL:
		case PCSCF_IPv4AddrReqUL:
			i.FromMs.P_CSCF_IPv4AddrReq = true
		case DNSServerIPv4AddrReqUL:
			i.FromMs.DNSV4Req = true
		case MSISDNReqUL:
		case IFOMSupportReqUL:
		case IPv4LinkMTUReqUL:
			i.FromMs.IPv4LinkMTUReq = true
		case MSSupportOfLocalAddrInTFTIndicatorUL:
		case PCSCF_ReSelectionSupportUL:
		case NBIFOMReqIndicatorUL:
		case NBIFOMModeUL:
		case NonIPLinkMTUReqUL:
		case APNRateCtrl_SupportIndicatorUL:
		case UEStatus3GPPPSDataOffUL:
			i.FromMs.UEStatus3GPPPSDataOff = UEStatus3GPPPSDataOff(b[ofs : ofs+int(cntLen)][0])
		case ReliableDataSvcReqIndicatorUL:
		case AdditionalAPNRateCtrlForExceptionData_SupportIndicatorUL:
		case PDUSessIDUL:
		case EthFramePayloadMTUReqUL:
		case UnstructuredLinkMTUReqUL:
		case I5GSMCauseValUL: // 5GSMCauseValUL
		case QoSRulesWithTheLenOf2Octets_SupportIndicatorUL:
		case QoSFlowDescWithTheLenOf2Octets_SupportIndicatorUL:
		case ACSInfoReqUL:
		case ATSSSReqUL:
		case DNSServerSecInfoIndicatorUL:
		case LinkCtrlProtUL:
		case PushAccessCtrlProtUL:
		case ChlgHandshakeAuthProtUL:
		case InternetProtCtrlProtUL:
		default:
		}
		ofs += int(cntLen)
	}

	return nil
}

func (i *ExtendedProtCfgOpts) UnmarshalFromNw(b []byte) error {
	Len := len(b)
	if Len < 1 {
		return errors.Errorf("ExtendedProtCfgOpts: UnmarshalFromNw() bad total length (%d<1)", Len)
	}

	// All other values are interpreted as PPP in this version of the protocol.
	i.CfgProt = PPP_IP_PDP_or_IP_PDN
	ofs := 1

	i.FromNw = &ExtCfgOptFromNw{}
	Len_idLen := 2 + 1

	for ofs+Len_idLen <= Len {
		protId := binary.BigEndian.Uint16(b[ofs : ofs+2])
		ofs += 2
		cntLen := b[ofs]
		ofs++
		if ofs+int(cntLen) > Len {
			i.FromNw = &ExtCfgOptFromNw{}
			return errors.Errorf("ExtendedProtCfgOpts: UnmarshalFromNw() bad length, ofs=%d, cntLen=%d, total Len=%d",
				ofs, cntLen, Len)
		}

		switch protId {
		case PCSCF_IPv6AddrDL:
		case IMCNSubsystemSignalingFlagDL:
		case DNSServerIPv6AddrDL:
			i.FromNw.DNSIPv6Addr = b[ofs : ofs+int(cntLen)]
		case PolicyCtrlRejCodeDL:
		case SelectedBearerCtrlModeDL:
		case DSMIPv6HomeAgentAddrDL:
		case DSMIPv6HomeNwPrefixDL:
		case DSMIPv6IPv4HomeAgentAddrDL:
		case PCSCF_IPv4AddrDL:
			i.FromNw.P_CSCF_IPv4Addr = b[ofs : ofs+int(cntLen)]
		case DNSServerIPv4AddrDL:
			i.FromNw.DNSIPv4Addr = b[ofs : ofs+int(cntLen)]
		case MSISDNDL:
		case IFOMSupportDL:
		case IPv4LinkMTUDL:
			if cntLen != 2 {
				i.FromNw = &ExtCfgOptFromNw{}
				return errors.Errorf("ExtendedProtCfgOpts: UnmarshalFromNw() bad IPv4LinkMTUDL len, cntLen=%d",
					cntLen)
			}
			i.FromNw.IPv4LinkMTU = binary.BigEndian.Uint16(b[ofs : ofs+int(cntLen)])
		case NwSupportOfLocalAddrInTFTIndicatorDL:
		case NBIFOMAcceptedIndicatorDL:
		case NBIFOMModeDL:
		case NonIPLinkMTUDL:
		case APNRateCtrlParamDL:
		case Indication3GPPPSDataOffSupportDL:
			i.FromNw.Indication3GPPPSDataOffSupport = true
		case ReliableDataSvcAcceptedIndicatorDL:
		case AdditionalAPNRateCtrlForExceptionDataParamDL:
		case SNSSAIDL:
		case QoSRulesDL:
		case SessAMBRDL:
		case PDUSessAddrLifetimeDL:
		case QoSFlowDescDL:
		case EthFramePayloadMTUDL:
		case UnstructuredLinkMTUDL:
		case QoSRulesWithTheLenOf2OctetsDL:
		case QoSFlowDescWithTheLenOf2OctetsDL:
		case SmallDataRateCtrlParamDL:
		case AdditionalSmallDataRateCtrlForExceptionDataParamDL:
		case ACSInfoDL:
		case InitSmallDataRateCtrlParamDL:
		case InitAdditionalSmallDataRateCtrlForExceptionDataParamDL:
		case InitAPNRateCtrlParamDL:
		case InitAdditionalAPNRateCtrlForExceptionDataParamDL:
		case ATSSSRspWithTheLenOf2OctetsDL:
		case DNSServerSecInfoWithLenOf2OctetsDL:
		default:
		}
		ofs += int(cntLen)
	}
	return nil
}

func appendToBuf(b []byte, protId uint16, content []byte) []byte {
	var tmp [protIdLen]byte
	binary.BigEndian.PutUint16(tmp[:], protId)
	b = append(b, tmp[:]...)
	b = append(b, byte(len(content)))
	b = append(b, content[:]...)
	return b
}

func (i *ExtendedProtCfgOpts) marshalFromMs() ([]byte, error) {
	var b []byte

	ms := i.FromMs
	if ms == nil {
		return nil, errors.Errorf("ExtendedProtCfgOpts: marshalFromMs() i.FromMs == nil")
	}

	var b1 uint8
	b1 = SetBit8(b1, 1)
	b1 = Set3Bits31(b1, uint8(PPP_IP_PDP_or_IP_PDN))
	b = append(b, b1)

	if ms.DNSV6Req {
		b = appendToBuf(b, DNSServerIPv6AddrReqUL, nil)
	}
	if ms.P_CSCF_IPv4AddrReq {
		b = appendToBuf(b, PCSCF_IPv4AddrReqUL, nil)
	}
	if ms.DNSV4Req {
		b = appendToBuf(b, DNSServerIPv4AddrReqUL, nil)
	}
	if ms.IPv4LinkMTUReq {
		b = appendToBuf(b, IPv4LinkMTUReqUL, nil)
	}
	if ms.UEStatus3GPPPSDataOff != UEStatus3GPPPSDataOff_NotPresent {
		content := []byte{byte(ms.UEStatus3GPPPSDataOff)}
		b = appendToBuf(b, UEStatus3GPPPSDataOffUL, content)
	}
	return b, nil
}

func (i *ExtendedProtCfgOpts) marshalFromNw() ([]byte, error) {
	tmp := make([]byte, 16)
	var b []byte

	nw := i.FromNw
	if nw == nil {
		return nil, errors.Errorf("ExtendedProtCfgOpts: marshalFromNw() i.FromNw == nil")
	}

	var b1 uint8
	b1 = SetBit8(b1, 1)
	b1 = Set3Bits31(b1, uint8(PPP_IP_PDP_or_IP_PDN))
	b = append(b, b1)

	if !nw.DNSIPv6Addr.IsUnspecified() {
		if ip := nw.DNSIPv6Addr; ip != nil {
			if ip.To16() == nil {
				return nil, errors.Errorf(
					"(*ExtendedProtCfgOpts) marshalFromNw(): ip.To16() is nil")
			}
			ctntLen := uint8(16)
			b = appendToBuf(b, DNSServerIPv6AddrDL, ip.To16()[0:ctntLen])
		}
	}
	if !nw.P_CSCF_IPv4Addr.IsUnspecified() {
		if ip := nw.P_CSCF_IPv4Addr; ip != nil {
			if ip.To4() == nil {
				return nil, errors.Errorf(
					"(*ExtendedProtCfgOpts) marshalFromNw(): ip.To4() is nil")
			}
			ctntLen := uint8(4)
			b = appendToBuf(b, PCSCF_IPv4AddrDL, ip.To4()[0:ctntLen])
		}
	}
	if !nw.DNSIPv4Addr.IsUnspecified() {
		if ip := nw.DNSIPv4Addr; ip != nil {
			if ip.To4() == nil {
				return nil, errors.Errorf(
					"(*ExtendedProtCfgOpts) marshalFromNw(): ip.To4() is nil")
			}
			ctntLen := uint8(4)
			b = appendToBuf(b, DNSServerIPv4AddrDL, ip.To4()[0:ctntLen])
		}
	}
	if mtu := nw.IPv4LinkMTU; mtu != 0 {
		ctntLen := uint8(2)
		binary.BigEndian.PutUint16(tmp[0:ctntLen], mtu)
		b = appendToBuf(b, IPv4LinkMTUDL, tmp[0:ctntLen])
	}
	if nw.Indication3GPPPSDataOffSupport {
		b = appendToBuf(b, Indication3GPPPSDataOffSupportDL, nil)
	}
	return b, nil
}

// MarshalBinary returns the Val part, not including IEI and Len.
func (i *ExtendedProtCfgOpts) MarshalBinary() ([]byte, error) {
	if i.FromMs != nil && i.FromNw != nil {
		return nil, errors.Errorf("ExtendedProtCfgOpts: Both FromMS / FromNw != nil, which way to marshal?")
	}
	if i.FromMs == nil && i.FromNw == nil {
		return nil, errors.Errorf("ExtendedProtCfgOpts: Both FromMS / FromNw == nil")
	}
	if i.FromNw != nil {
		return i.marshalFromNw()
	}
	return i.marshalFromMs()
}
