package ie

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	PFCType_MatchAllType              uint8 = 0x01
	PFCType_Ipv4RemoteAddr            uint8 = 0x10
	PFCType_Ipv4LocalAddr             uint8 = 0x11
	PFCType_Ipv6RemoteAddrPrefixLen   uint8 = 0x21
	PFCType_Ipv6LocalAddrPrefixLen    uint8 = 0x23
	PFCType_ProtocolIdentifierNextHdr uint8 = 0x30
	PFCType_SingleLocalPort           uint8 = 0x40
	PFCType_LocalPortRange            uint8 = 0x41
	PFCType_SingleRemotePort          uint8 = 0x50
	PFCType_RemotePortRange           uint8 = 0x51
	PFCType_SecurityParamIndex        uint8 = 0x60
	PFCType_TypeOfSvcTrafficClass     uint8 = 0x70
	PFCType_FlowLabel                 uint8 = 0x80
	PFCType_DstMacAddr                uint8 = 0x81
	PFCType_SrcMacAddr                uint8 = 0x82
	PFCType_802_1QCTAG_VID            uint8 = 0x83
	PFCType_802_1QSTAG_VID            uint8 = 0x84
	PFCType_802_1QCTAG_PCPDEI         uint8 = 0x85
	PFCType_802_1QSTAG_PCPDEI         uint8 = 0x86
	PFCType_Ethtype                   uint8 = 0x87
	PFCType_DstMACAddrRange           uint8 = 0x88
	PFCType_SrcMACAddrRange           uint8 = 0x89
)

type QosRuleOpCode uint8 // QoS Rule Operation Code
const (
	OpCode_CreateNewQosRule            QosRuleOpCode = 0x01
	OpCode_DelExistingQosRule          QosRuleOpCode = 0x02
	OpCode_ModifyAddPktFilters         QosRuleOpCode = 0x03
	OpCode_ModifyReplaceAllPktFilters  QosRuleOpCode = 0x04
	OpCode_ModifyDelPktFilters         QosRuleOpCode = 0x05
	OpCode_ModifyWoModifyingPktFilters QosRuleOpCode = 0x06
	OpCode_Reserved                    QosRuleOpCode = 0x07
)

const (
	PFD_Downlink uint8 = 0x01
	PFD_Uplink   uint8 = 0x02
	PFD_BiDir    uint8 = 0x03
)

const (
	szPfCnt_Addr            uint8 = 17 // 16 bytes IPv6 addr + 1 byte mask
	szPfCnt_Port            uint8 = 4  // Port range: 2 * 16bits port
	szPfCnt_PIorNH          uint8 = 1
	szPfCnt_SPI             uint8 = 4
	szPfCnt_TosTrafficClass uint8 = 2
	szPfCnt_FlowLabel       uint8 = 3 // 4bits spare + 20bits IPv6 flow label
	szPfCnt_MacAddr         uint8 = 6
)

type PacketFilterContents struct {
	MatchAll        bool
	HavePIorNH      bool
	PIorNH          uint8 // IPv4 Protocol ID or IPv6 Next Header
	_               uint8
	RemoteAddr      string
	RemotePortRange string
	LocalAddr       string
	LocalPortRange  string
	SPI             string
	TosTrafficClass string
	FlowLabel       string
	DstMacAddr      string
	SrcMacAddr      string
	DstMacAddrRange [2]string
	SrcMacAddrRange [2]string
}

type PacketFilter struct {
	Dir      uint8 // 6 -> 5  , 8  -> 8
	Id       uint8 // 4 -> 1  , 8  -> 8
	Contents PacketFilterContents
}

type QosRule struct {
	RuleId        uint8          // 8 -> 1 , 4  ->  4
	OpCode        QosRuleOpCode  // 8 -> 6 , 7  ->  7
	IsDefaultDQR  bool           // 5 -> 5 , 7  ->  7
	PktFilterList []PacketFilter // 8  -> 8  , 8 ->  m
	Precedence    uint8          // 8 -> 1 , m+1 -> m+1
	Segregation   bool           // 7 -> 7 , m+2 -> m+2
	QFI           uint8          // 6 -> 1 , m+2 -> m+2k
}

// QosRules is detailed in 9.11.4.13 QoS rules , 24.501
type QosRules struct {
	Rules []QosRule
}

// ToDo
// In each packet filter, there shall not be more than one occurrence of each packet filter component type
func (p *PacketFilterContents) UnmarshalBinary(b []byte) error {
	ttlLen := len(b)
	ofs := 0

	p.RemoteAddr = "any"
	p.LocalAddr = "assigned"

	for ofs < ttlLen {
		switch b[ofs] {
		case PFCType_MatchAllType:
			ofs++
			if ttlLen != 1 || ofs != 1 {
				return errors.Errorf("PFC MatchAll, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.MatchAll = true

		case PFCType_Ipv4RemoteAddr:
			ofs++
			if ofs+8 > ttlLen {
				return errors.Errorf("PFC v4 RemoteAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			var ipnet net.IPNet
			ipnet.IP = net.IPv4(b[ofs], b[ofs+1], b[ofs+2], b[ofs+3])
			ofs += 4
			ipnet.Mask = net.IPv4Mask(b[ofs], b[ofs+1], b[ofs+2], b[ofs+3])
			ofs += 4
			p.RemoteAddr = ipnet.String()

		case PFCType_Ipv4LocalAddr:
			ofs++
			if ofs+8 > ttlLen {
				return errors.Errorf("PFC v4 LocalAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			var ipnet net.IPNet
			ipnet.IP = net.IPv4(b[ofs], b[ofs+1], b[ofs+2], b[ofs+3])
			ofs += 4
			ipnet.Mask = net.IPv4Mask(b[ofs], b[ofs+1], b[ofs+2], b[ofs+3])
			ofs += 4
			p.LocalAddr = ipnet.String()

		case PFCType_Ipv6RemoteAddrPrefixLen:
			ofs++
			if ofs+16+1 > ttlLen {
				return errors.Errorf("PFC v6 RemoteAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			var ipnet net.IPNet
			ipnet.IP = b[ofs : ofs+16]
			ofs += 16
			ipnet.Mask = net.CIDRMask(int(b[ofs]), net.IPv6len*8)
			ofs += 1
			p.RemoteAddr = ipnet.String()

		case PFCType_Ipv6LocalAddrPrefixLen:
			ofs++
			if ofs+16+1 > ttlLen {
				return errors.Errorf("PFC v6 LocalAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			var ipnet net.IPNet
			ipnet.IP = b[ofs : ofs+16]
			ofs += 16
			ipnet.Mask = net.CIDRMask(int(b[ofs]), net.IPv6len*8)
			ofs += 1
			p.LocalAddr = ipnet.String()

		case PFCType_ProtocolIdentifierNextHdr:
			ofs++
			if ofs+1 > ttlLen {
				return errors.Errorf("PFC PIorNH, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.PIorNH = b[ofs]
			ofs++
			p.HavePIorNH = true

		case PFCType_SingleLocalPort:
			ofs++
			if ofs+2 > ttlLen {
				return errors.Errorf("PFC Single Local Port, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.LocalPortRange = fmt.Sprintf("%d", binary.BigEndian.Uint16(b[ofs:ofs+2]))
			ofs += 2

		case PFCType_LocalPortRange:
			ofs++
			if ofs+4 > ttlLen {
				return errors.Errorf("PFC Local Port Range, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.LocalPortRange = fmt.Sprintf("%d-%d",
				binary.BigEndian.Uint16(b[ofs+0:ofs+2]),
				binary.BigEndian.Uint16(b[ofs+2:ofs+4]),
			)
			ofs += 4

		case PFCType_SingleRemotePort:
			ofs++
			if ofs+2 > ttlLen {
				return errors.Errorf("PFC Single Remote Port, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.RemotePortRange = fmt.Sprintf("%d",
				binary.BigEndian.Uint16(b[ofs+0:ofs+2]),
			)
			ofs += 2

		case PFCType_RemotePortRange:
			ofs++
			if ofs+4 > ttlLen {
				return errors.Errorf("PFC Remote Port Range, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.RemotePortRange = fmt.Sprintf("%d-%d",
				binary.BigEndian.Uint16(b[ofs+0:ofs+2]),
				binary.BigEndian.Uint16(b[ofs+2:ofs+4]),
			)
			ofs += 4

		case PFCType_SecurityParamIndex:
			ofs++
			if ofs+4 > ttlLen {
				return errors.Errorf("PFC SPI, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.SPI = fmt.Sprintf("%04x", binary.BigEndian.Uint32(b[ofs:ofs+4]))
			ofs += 4

		case PFCType_TypeOfSvcTrafficClass:
			ofs++
			if ofs+2 > ttlLen {
				return errors.Errorf("PFC TOS-TC, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.TosTrafficClass = fmt.Sprintf("%02x%02x", b[ofs], b[ofs+1])
			ofs += 2

		case PFCType_FlowLabel:
			ofs++
			if ofs+3 > ttlLen {
				return errors.Errorf("PFC FlowLabel, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			tmp := (uint32(b[ofs]&0x0f) << 16) + uint32(b[ofs+1])<<8 + uint32(b[ofs+2])
			p.FlowLabel = fmt.Sprintf("%03x", tmp)
			ofs += 3

		case PFCType_SrcMacAddr:
			ofs++
			if ofs+6 > ttlLen {
				return errors.Errorf("PFC SrcMacAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.SrcMacAddr = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6

		case PFCType_DstMacAddr:
			ofs++
			if ofs+6 > ttlLen {
				return errors.Errorf("PFC DstMacAddr, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.DstMacAddr = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6

		case PFCType_SrcMACAddrRange:
			ofs++
			if ofs+12 > ttlLen {
				return errors.Errorf("PFC SrcMacAddrRange, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.SrcMacAddrRange[0] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6
			p.SrcMacAddrRange[1] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6

		case PFCType_DstMACAddrRange:
			ofs++
			if ofs+12 > ttlLen {
				return errors.Errorf("PFC DstMacAddrRange, bad offset(%d) (ttlLen=%d)", ofs, ttlLen)
			}
			p.DstMacAddrRange[0] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6
			p.DstMacAddrRange[1] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				b[ofs], b[ofs+1], b[ofs+2], b[ofs+3], b[ofs+4], b[ofs+5])
			ofs += 6

		default:
			return errors.Errorf("PFC Unmarshal: unknown type %d(0x%x)", b[ofs], b[ofs])
		}
	}
	return nil
}

func (p *PacketFilterContents) MarshalBinary() ([]byte, error) {
	var b [256]byte
	ofs := 0

	if p.RemoteAddr != "any" {
		_, ipNet, err := net.ParseCIDR(p.RemoteAddr)
		if err != nil {
			return nil, errors.Wrap(err, "PFC MarshalBinary() RA")
		}
		switch len(ipNet.IP) {
		case 4:
			b[ofs] = PFCType_Ipv4RemoteAddr
			ofs++
			copy(b[ofs:ofs+len(ipNet.IP)], ipNet.IP)
			ofs += len(ipNet.IP)
			copy(b[ofs:ofs+len(ipNet.Mask)], ipNet.Mask)
			ofs += len(ipNet.Mask)
		case 16:
			b[ofs] = PFCType_Ipv6RemoteAddrPrefixLen
			ofs++
			copy(b[ofs:ofs+len(ipNet.IP)], ipNet.IP)
			ofs += len(ipNet.IP)
			prefixLen, _ := ipNet.Mask.Size()
			b[ofs] = uint8(prefixLen)
			ofs += 1
		}
	}

	if p.LocalAddr != "any" && p.LocalAddr != "assigned" {
		_, ipNet, err := net.ParseCIDR(p.LocalAddr)
		if err != nil {
			return nil, errors.Wrap(err, "PFC MarshalBinary() LA")
		}
		ipLen := len(ipNet.IP)
		switch len(ipNet.IP) {
		case 4:
			b[ofs] = PFCType_Ipv4LocalAddr
			ofs++
			copy(b[ofs:ofs+ipLen], ipNet.IP)
			ofs += ipLen
			maskLen := len(ipNet.Mask)
			copy(b[ofs:ofs+maskLen], ipNet.Mask)
			ofs += maskLen
		case 16:
			b[ofs] = PFCType_Ipv6LocalAddrPrefixLen
			ofs++
			copy(b[ofs:ofs+ipLen], ipNet.IP)
			ofs += ipLen
			prefixLen, _ := ipNet.Mask.Size()
			b[ofs] = uint8(prefixLen)
			ofs += 1
		}
	}

	if p.RemotePortRange != "" {
		ports := strings.Split(p.RemotePortRange, "-")
		switch len(ports) {
		case 1:
			b[ofs] = PFCType_SingleRemotePort
		case 2:
			fallthrough
		default:
			ports = ports[:2]
			b[ofs] = PFCType_RemotePortRange
		}
		ofs++

		for _, port := range ports {
			if val, err := strconv.Atoi(port); err != nil {
				return nil, errors.Wrap(err, "PFC MarshalBinary() RPR")
			} else {
				binary.BigEndian.PutUint16(b[ofs:ofs+2], uint16(val))
				ofs += 2
			}
		}
	}

	if p.LocalPortRange != "" {
		ports := strings.Split(p.LocalPortRange, "-")
		switch len(ports) {
		case 1:
			b[ofs] = PFCType_SingleLocalPort
		case 2:
			fallthrough
		default:
			ports = ports[:2]
			b[ofs] = PFCType_LocalPortRange
		}
		ofs++

		for _, port := range ports {
			if val, err := strconv.Atoi(port); err != nil {
				return nil, errors.Wrap(err, "PFC MarshalBinary() LPR")
			} else {
				binary.BigEndian.PutUint16(b[ofs:ofs+2], uint16(val))
				ofs += 2
			}
		}
	}

	if p.HavePIorNH {
		b[ofs] = PFCType_ProtocolIdentifierNextHdr
		b[ofs+1] = p.PIorNH
		ofs += 2
	}

	if p.SPI != "" {
		sz := int(szPfCnt_SPI)
		if val, err := strconv.ParseUint(p.SPI, 16, 32); err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() SPI")
		} else {
			b[ofs] = PFCType_SecurityParamIndex
			ofs++
			binary.BigEndian.PutUint32(b[ofs:ofs+sz], uint32(val))
			ofs += sz
		}
	}
	if p.TosTrafficClass != "" {
		sz := int(szPfCnt_TosTrafficClass)
		if val, err := strconv.ParseUint(p.TosTrafficClass, 16, 16); err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() Tos Traffic class")
		} else {
			b[ofs] = PFCType_TypeOfSvcTrafficClass
			ofs++
			b[ofs] = uint8(val >> 8)
			b[ofs+1] = uint8(val & 0x00ff)
			ofs += sz
		}
	}
	if p.FlowLabel != "" {
		sz := int(szPfCnt_FlowLabel)
		if val, err := strconv.ParseUint(p.FlowLabel, 16, 32); err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() flow label")
		} else {
			b[ofs] = PFCType_FlowLabel
			ofs++
			tmp := [4]byte{}
			binary.BigEndian.PutUint32(tmp[:], uint32(val))
			copy(b[ofs:ofs+sz], tmp[1:])
			ofs += sz
		}
	}
	if p.SrcMacAddr != "" {
		sz := int(szPfCnt_MacAddr)
		if val, err := net.ParseMAC(p.SrcMacAddr); err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() src mac")
		} else {
			b[ofs] = PFCType_SrcMacAddr
			ofs++
			copy(b[ofs:ofs+sz], val)
			ofs += sz
		}
	}
	if p.DstMacAddr != "" {
		sz := int(szPfCnt_MacAddr)
		if val, err := net.ParseMAC(p.DstMacAddr); err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() dst mac")
		} else {
			b[ofs] = PFCType_DstMacAddr
			ofs++
			copy(b[ofs:ofs+sz], val)
			ofs += sz
		}
	}
	if p.SrcMacAddrRange[0] != "" {
		sz := int(szPfCnt_MacAddr)
		var mac1, mac2 net.HardwareAddr
		mac1, err := net.ParseMAC(p.SrcMacAddrRange[0])
		if err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() Src mac1 range")
		}
		mac2, err = net.ParseMAC(p.SrcMacAddrRange[1])
		if err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() Src mac2 range")
		}
		b[ofs] = PFCType_SrcMACAddrRange
		ofs++
		copy(b[ofs:ofs+sz], mac1)
		ofs += sz
		copy(b[ofs:ofs+sz], mac2)
		ofs += sz
	}
	if p.DstMacAddrRange[0] != "" {
		sz := int(szPfCnt_MacAddr)
		var mac1, mac2 net.HardwareAddr
		mac1, err := net.ParseMAC(p.DstMacAddrRange[0])
		if err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() dst mac1 range")
		}
		mac2, err = net.ParseMAC(p.DstMacAddrRange[1])
		if err != nil {
			return nil, errors.Wrap(err, "PCF MarshalBinary() dst mac2 range")
		}
		b[ofs] = PFCType_DstMACAddrRange
		ofs++
		copy(b[ofs:ofs+sz], mac1)
		ofs += sz
		copy(b[ofs:ofs+sz], mac2)
		ofs += sz
	}

	if ofs == 0 {
		b[ofs] = PFCType_MatchAllType
		ofs += 1
	}
	return b[:ofs], nil
}

func (p *PacketFilterContents) marshalLenEstimate() uint8 {
	lenEst := uint8(0)

	if p.RemoteAddr != "any" {
		lenEst += 1 + szPfCnt_Addr
	}
	if p.LocalAddr != "any" && p.LocalAddr != "assigned" {
		lenEst += 1 + szPfCnt_Addr
	}
	if p.RemotePortRange != "" {
		lenEst += 1 + szPfCnt_Port
	}
	if p.LocalPortRange != "" {
		lenEst += 1 + szPfCnt_Port
	}
	if p.HavePIorNH {
		lenEst += 1 + szPfCnt_PIorNH
	}
	if p.SPI != "" {
		lenEst += 1 + szPfCnt_SPI
	}
	if p.TosTrafficClass != "" {
		lenEst += 1 + szPfCnt_TosTrafficClass
	}
	if p.FlowLabel != "" {
		lenEst += 1 + szPfCnt_FlowLabel
	}
	if p.DstMacAddr != "" {
		lenEst += 1 + szPfCnt_MacAddr
	}
	if p.SrcMacAddr != "" {
		lenEst += 1 + szPfCnt_MacAddr
	}
	if p.SrcMacAddrRange[0] != "" &&
		p.SrcMacAddrRange[1] != "" {
		lenEst += 1 + szPfCnt_MacAddr*2
	}
	if p.DstMacAddrRange[0] != "" &&
		p.DstMacAddrRange[1] != "" {
		lenEst += 1 + szPfCnt_MacAddr*2
	}
	if lenEst == 0 {
		lenEst += 1
	}
	return lenEst
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *QosRules) UnmarshalBinary(b []byte) error {
	qosRulesLen := uint16(len(b))
	rule := QosRule{}
	ofs := uint16(0)
	for ofs < qosRulesLen-2 {
		if umlen, err := rule.unmarshalQosRule(b[ofs:]); err != nil {
			i.Rules = nil
			return errors.Wrap(err, "qosRule UnmarshalBinary()")
		} else {
			ofs += umlen
			if rule.OpCode != OpCode_DelExistingQosRule {
				if ofs < qosRulesLen-1 {
					rule.Precedence = b[ofs]
					ofs++
					rule.Segregation = GetBit7(b[ofs]) == 1
					rule.QFI = Get6Bits61(b[ofs])
					ofs++
				}
			}
			i.Rules = append(i.Rules, rule)
		}
	}
	return nil
}

func (r *QosRule) unmarshalQosRule(b []byte) (uint16, error) {
	if len(b) < 4 {
		return 0, errors.Errorf("qosRule: len(b) = %d < 4!", len(b))
	}
	r.RuleId = b[0]
	ruleLen := binary.BigEndian.Uint16(b[1:3])
	r.OpCode = QosRuleOpCode(Get3Bits86(b[3]))
	if r.OpCode == OpCode_DelExistingQosRule {
		if ruleLen != 1 {
			return 0, errors.Errorf("qosRule: del operation should have len==1, got %d.", len(b))
		} else {
			return 4, nil
		}
	}
	r.IsDefaultDQR = GetBit5(b[3]) == 1
	nrOfPktFilters := Get4Bits41(b[3])
	r.PktFilterList = make([]PacketFilter, nrOfPktFilters)
	ofs := uint16(4) // Offset to the start of pkt filter list
	if nil == r.PktFilterList {
		return 0, errors.Errorf("qosRule :Failed to make %d []PacketFilter", nrOfPktFilters)
	}
	for j := uint8(0); j < nrOfPktFilters; j++ {
		flt := &r.PktFilterList[j]
		// 2: i.Contents is at the 2nd octet
		if int(ofs)+2 > len(b) {
			r.PktFilterList = nil
			return 0, errors.Errorf("qosRule : bad pkt flt content? ofs=%d, len(b)=%d", ofs, len(b))
		}
		flt.Dir = Get2Bits65(b[ofs])
		flt.Id = Get4Bits41(b[ofs])

		pfcLen := uint16(b[ofs+1])
		ofs += 2
		if int(ofs+pfcLen) > len(b) {
			r.PktFilterList = nil
			return 0, errors.Errorf("qosRule : bad pkt flt content? ofs=%d, pfcLen=%d, len(b)=%d", ofs, pfcLen, len(b))
		}
		if err := flt.Contents.UnmarshalBinary(b[ofs : ofs+pfcLen]); err != nil {
			return 0, errors.Wrap(err, "qosRule unmarshalQosRule()")
		}
		ofs += pfcLen
	}

	// ruleLen + 3 : len of the whole QoS Rule, 3 = 1B "QoS Rule ID" + 2B "Len of QoS Rule"
	// ofs value here : len of the whole QoS Rule - 2 Octets (Precedence + the one w/ QFI)
	if ruleLen+3 != ofs+2 {
		return ofs, errors.Errorf("ruleLen=%d, ofs=%d", ruleLen, ofs)
	}
	return ofs, nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *QosRules) MarshalBinary() ([]byte, error) {
	qosRulesLen := uint16(0)
	szFixedFieldsInQosRule := uint16(4)
	for j := range i.Rules {
		rule := i.Rules[j]
		qosRulesLen += szFixedFieldsInQosRule

		if rule.OpCode != OpCode_DelExistingQosRule {
			// QoS rule's variable length part:
			for k := range rule.PktFilterList {
				qosRulesLen += 2 + uint16(rule.PktFilterList[k].Contents.marshalLenEstimate())
			}
		}
	}
	b := make([]byte, 0, qosRulesLen)
	for j := range i.Rules {
		rule := i.Rules[j]
		if mb, err := rule.marshalQosRule(); err != nil {
			return nil, errors.Wrap(err, "QosRules MarshalBinary()")
		} else {
			b = append(b, mb...)
		}
	}
	return b, nil
}

func (r *QosRule) marshalQosRule() ([]byte, error) {
	Sz_QoSRuleID_LenOfQosRule := uint16(3)
	maxRuleLen := Sz_QoSRuleID_LenOfQosRule

	for j := range r.PktFilterList {
		maxRuleLen += 2 + uint16(r.PktFilterList[j].Contents.marshalLenEstimate())
	}
	b := make([]byte, Sz_QoSRuleID_LenOfQosRule+maxRuleLen)

	b[0] = r.RuleId
	b[3] = Set3Bits86(b[3], uint8(r.OpCode))
	b[3] = SetBit5(b[3], bool2uint8(r.IsDefaultDQR))

	ofs := uint16(4) // Offset to the start of pkt filter list
	if r.OpCode == OpCode_DelExistingQosRule {
		binary.BigEndian.PutUint16(b[1:3], ofs-3)
		return b[:ofs], nil
	}

	nrOfPktFilters := uint8(0)
	for j := range r.PktFilterList {
		flt := r.PktFilterList[j]
		b[ofs] = Set2Bits65(b[ofs], flt.Dir)
		b[ofs] = Set4Bits41(b[ofs], flt.Id)
		ofs++
		if ctntBuf, err := flt.Contents.MarshalBinary(); err != nil {
			return nil, errors.Wrap(err, "QosRule marshalQosRule()")
		} else {
			contentsLen := uint16(len(ctntBuf))
			b[ofs] = uint8(contentsLen)
			ofs++
			copy(b[ofs:ofs+contentsLen], ctntBuf)
			ofs += contentsLen
		}
		nrOfPktFilters++
	}
	b[ofs] = r.Precedence
	ofs++
	b[ofs] = SetBit7(b[ofs], bool2uint8(r.Segregation))
	b[ofs] = Set6Bits61(b[ofs], r.QFI)
	ofs++
	b[3] = Set4Bits41(b[3], nrOfPktFilters)
	binary.BigEndian.PutUint16(b[1:3], ofs-3)

	return b[:ofs], nil
}

func (nasQR *QosRules) String() string {
	var brief string
	var operation string
	var pktFilter string

	if nasQR == nil || len(nasQR.Rules) == 0 {
		return ""
	}

	op := func(opCode QosRuleOpCode) string {
		switch opCode {
		case OpCode_CreateNewQosRule:
			return "add"
		case OpCode_DelExistingQosRule:
			return "del"
		case OpCode_ModifyReplaceAllPktFilters:
			return "mod"
		case OpCode_ModifyAddPktFilters:
			return "mod+"
		case OpCode_ModifyDelPktFilters:
			return "mod-"
		case OpCode_ModifyWoModifyingPktFilters:
			return "modW"
		}
		return "BadOpCode"
	}

	addrPort := func(addr, port string) string {
		if port == "" {
			return addr
		}
		return fmt.Sprintf("%s:%s", addr, port)
	}

	pf := func(pfs []PacketFilter) string {
		var str []string
		for _, p := range pfs {
			if p.Contents.MatchAll {
				str = append(str, "MatchAll")
				continue
			}
			str = append(str, fmt.Sprintf("%s-->%s",
				addrPort(p.Contents.RemoteAddr, p.Contents.RemotePortRange),
				addrPort(p.Contents.LocalAddr, p.Contents.LocalPortRange)))
		}
		if len(str) == 0 {
			return ""
		}
		return "," + strings.Join(str, ",")
	}

	i := 0
	for _, rule := range nasQR.Rules {
		if rule.RuleId == 0 {
			continue
		}
		if i != 0 {
			brief += ","
		}
		i++
		operation = op(rule.OpCode)
		pktFilter = pf(rule.PktFilterList)

		brief += fmt.Sprintf("%s QosRule[ID:%d,DQF:%v,prec:%d,QFI:%d%s]",
			operation, rule.RuleId, rule.IsDefaultDQR, rule.Precedence, rule.QFI,
			pktFilter)
	}

	return brief
}
