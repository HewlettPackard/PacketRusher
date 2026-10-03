package ie

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	QfdOpCode  uint8
	QfdParamId uint8
	Qfd5QIType uint8
)

const (
	// OpCode : Create / Delete
	QfdEbit_NoParamList  uint8 = 0
	QfdEbit_HasParamList uint8 = 1

	// OpCode: Modify
	QfdEbit_ExtenPrevParam   uint8 = 0
	QfdEbit_ReplacePrevParam uint8 = 1

	QFD_None   QfdOpCode = 0
	QFD_Create QfdOpCode = 1
	QFD_Del    QfdOpCode = 2
	QFD_Mod    QfdOpCode = 3

	QfdParamId_5QI             QfdParamId = 1
	QfdParamId_GFBR_Uplink     QfdParamId = 2
	QfdParamId_GFBR_Downlink   QfdParamId = 3
	QfdParamId_MFBR_Uplink     QfdParamId = 4
	QfdParamId_MFBR_Downlink   QfdParamId = 5
	QfdParamId_AveragingWindow QfdParamId = 6
	QfdParamId_EPSBearerId     QfdParamId = 7

	Qfd_Reserved1 Qfd5QIType = 0x00
	Qfd_5QI1      Qfd5QIType = 0x01
	Qfd_5QI2      Qfd5QIType = 0x02
	Qfd_5QI3      Qfd5QIType = 0x03
	Qfd_5QI4      Qfd5QIType = 0x04
	Qfd_5QI5      Qfd5QIType = 0x05
	Qfd_5QI6      Qfd5QIType = 0x06
	Qfd_5QI7      Qfd5QIType = 0x07
	Qfd_5QI8      Qfd5QIType = 0x08
	Qfd_5QI9      Qfd5QIType = 0x09
	Qfd_5QI10     Qfd5QIType = 0x0a

	Qfd_5QI65 Qfd5QIType = 0x41
	Qfd_5QI66 Qfd5QIType = 0x42
	Qfd_5QI67 Qfd5QIType = 0x43
	Qfd_5QI69 Qfd5QIType = 0x45
	Qfd_5QI70 Qfd5QIType = 0x46
	Qfd_5QI71 Qfd5QIType = 0x47
	Qfd_5QI72 Qfd5QIType = 0x48
	Qfd_5QI73 Qfd5QIType = 0x49
	Qfd_5QI74 Qfd5QIType = 0x4a
	Qfd_5QI75 Qfd5QIType = 0x4b
	Qfd_5QI76 Qfd5QIType = 0x4c

	Qfd_5QI79 Qfd5QIType = 0x4f
	Qfd_5QI80 Qfd5QIType = 0x50

	Qfd_5QI82 Qfd5QIType = 0x52
	Qfd_5QI83 Qfd5QIType = 0x53
	Qfd_5QI84 Qfd5QIType = 0x54
	Qfd_5QI85 Qfd5QIType = 0x55
	Qfd_5QI86 Qfd5QIType = 0x56
	Qfd_5QI87 Qfd5QIType = 0x57
	Qfd_5QI88 Qfd5QIType = 0x58
	Qfd_5QI89 Qfd5QIType = 0x59
	Qfd_5QI90 Qfd5QIType = 0x5a

	Qfd_OperatorSpecific5QIsBegin Qfd5QIType = 0x80
	Qfd_OperatorSpecific5QIsEnd   Qfd5QIType = 0xFE
	Qfd_Reserved2                 Qfd5QIType = 0xFF

	MAX_EPSBearerId uint8 = 15
)

type unitBps []struct {
	unit BitRateType
	bps  int64
}

type QFP_EPSBearerId struct {
	NumOfBearerId uint8
	Id            [MAX_EPSBearerId]uint8
}

// QosFlowDescs is detailed in 9.11.4.12 QoS flow descriptions, 24.501
type QosFlowDescs struct {
	Descs []QosFlowDesc
}

type QosFlowDesc struct {
	// Name, uint8, Bits, Octet
	QFI          uint8     // 6 -> 1 , 4 -> 4
	OpCode       QfdOpCode // 8 -> 6 , 5 -> 5
	EBit         uint8     // 7 -> 7 , 6 -> 6
	FiveQI       uint8
	GFBRUplink   string
	GFBRDownlink string
	MFBRUplink   string
	MFBRDownlink string
	AvgWin       uint16
	EPSBearerIds QFP_EPSBearerId
}

func (d *QosFlowDesc) calcParamsMarshalLength() uint8 {
	length := uint8(0)
	if d.FiveQI != uint8(Qfd_Reserved1) {
		length += 2 + 1
	}
	if d.GFBRUplink != "" {
		length += 2 + 3
	}
	if d.GFBRDownlink != "" {
		length += 2 + 3
	}
	if d.MFBRUplink != "" {
		length += 2 + 3
	}
	if d.MFBRDownlink != "" {
		length += 2 + 3
	}
	if d.AvgWin != 0 {
		length += 2 + 2
	}
	length += d.EPSBearerIds.NumOfBearerId * 3
	return length
}

func (d *QosFlowDesc) calcNumOfParams() uint8 {
	numOfParams := uint8(0)
	if d.FiveQI != uint8(Qfd_Reserved1) {
		numOfParams++
	}
	if d.GFBRUplink != "" {
		numOfParams++
	}
	if d.GFBRDownlink != "" {
		numOfParams++
	}
	if d.MFBRUplink != "" {
		numOfParams++
	}
	if d.MFBRDownlink != "" {
		numOfParams++
	}
	if d.AvgWin != 0 {
		numOfParams++
	}
	if d.EPSBearerIds.NumOfBearerId > 0 {
		numOfParams += d.EPSBearerIds.NumOfBearerId
	}
	return numOfParams
}

const (
	kb = int64(1000)
	mb = int64(1000 * 1000)
	gb = int64(1000 * 1000 * 1000)
	tb = int64(1000 * 1000 * 1000 * 1000)
	pb = int64(1000 * 1000 * 1000 * 1000 * 1000)
)

func parseBitRate(bitrate string) ([]byte, error) {
	s := strings.Split(bitrate, " ")
	if len(s) < 2 {
		return nil, errors.Errorf("bad bitrate:%v", bitrate)
	}
	var digit int64
	unitBpss := unitBps{
		{unit: Rate_1Kbps, bps: 1 * kb},
		{unit: Rate_4Kbps, bps: 4 * kb},
		{unit: Rate_16Kbps, bps: 16 * kb},
		{unit: Rate_64Kbps, bps: 64 * kb},
		{unit: Rate_256Kbps, bps: 256 * kb},
		{unit: Rate_1Mbps, bps: 1 * mb},
		{unit: Rate_4Mbps, bps: 4 * mb},
		{unit: Rate_16Mbps, bps: 16 * mb},
		{unit: Rate_64Mbps, bps: 64 * mb},
		{unit: Rate_256Mbps, bps: 256 * mb},
		{unit: Rate_1Gbps, bps: 1 * gb},
		{unit: Rate_4Gbps, bps: 4 * gb},
		{unit: Rate_16Gbps, bps: 16 * gb},
		{unit: Rate_64Gbps, bps: 64 * gb},
		{unit: Rate_256Gbps, bps: 256 * gb},
		{unit: Rate_1Tbps, bps: 1 * tb},
		{unit: Rate_4Tbps, bps: 4 * tb},
		{unit: Rate_16Tbps, bps: 16 * tb},
		{unit: Rate_64Tbps, bps: 64 * tb},
		{unit: Rate_256Tbps, bps: 256 * tb},
		{unit: Rate_1Pbps, bps: 1 * pb},
		{unit: Rate_4Pbps, bps: 4 * pb},
		{unit: Rate_16Pbps, bps: 16 * pb},
		{unit: Rate_64Pbps, bps: 64 * pb},
		{unit: Rate_256Pbps, bps: 256 * pb},
	}

	if n, err := strconv.ParseInt(s[0], 10, 64); err != nil || n < 0 {
		return nil, errors.Wrap(err, "parseBitRate failed to ParseInt")
	} else {
		digit = n
	}
	b := make([]byte, 3)

	unit := Rate_1Kbps
	mul := kb
	switch strings.ToLower(s[1]) {
	case "bps":
		// Kbps is the minimal unit
		digit = digit / 1000
	case "kbps":
		mul = kb
	case "mbps":
		unit = Rate_1Mbps
		mul = mb
	case "gbps":
		unit = Rate_1Gbps
		mul = gb
	case "tbps":
		unit = Rate_1Tbps
		mul = tb
	case "pbps":
		unit = Rate_1Pbps
		mul = pb
	default:
		// use Kbps by default
	}
	for i := range unitBpss {
		unitBps := unitBpss[i]
		if unit > unitBps.unit {
			continue
		}
		v := digit / (unitBps.bps / mul)
		if v <= 0xffff {
			b[0] = byte(unitBps.unit)
			binary.BigEndian.PutUint16(b[1:3], uint16(v))
			break
		}
	}
	return b, nil
}

type bitRateUnit struct {
	mult uint32
	unit string
}

var paramId2Str = map[BitRateType]bitRateUnit{
	Rate_1Kbps:   {mult: 1, unit: "Kbps"},
	Rate_4Kbps:   {mult: 4, unit: "Kbps"},
	Rate_16Kbps:  {mult: 16, unit: "Kbps"},
	Rate_64Kbps:  {mult: 64, unit: "Kbps"},
	Rate_256Kbps: {mult: 256, unit: "Kbps"},
	Rate_1Mbps:   {mult: 1, unit: "Mbps"},
	Rate_4Mbps:   {mult: 4, unit: "Mbps"},
	Rate_16Mbps:  {mult: 16, unit: "Mbps"},
	Rate_64Mbps:  {mult: 64, unit: "Mbps"},
	Rate_256Mbps: {mult: 256, unit: "Mbps"},
	Rate_1Gbps:   {mult: 1, unit: "Gbps"},
	Rate_4Gbps:   {mult: 4, unit: "Gbps"},
	Rate_16Gbps:  {mult: 16, unit: "Gbps"},
	Rate_64Gbps:  {mult: 64, unit: "Gbps"},
	Rate_256Gbps: {mult: 256, unit: "Gbps"},
	Rate_1Tbps:   {mult: 1, unit: "Tbps"},
	Rate_4Tbps:   {mult: 4, unit: "Tbps"},
	Rate_16Tbps:  {mult: 16, unit: "Tbps"},
	Rate_64Tbps:  {mult: 64, unit: "Tbps"},
	Rate_256Tbps: {mult: 256, unit: "Tbps"},
	Rate_1Pbps:   {mult: 1, unit: "Pbps"},
	Rate_4Pbps:   {mult: 4, unit: "Pbps"},
	Rate_16Pbps:  {mult: 16, unit: "Pbps"},
	Rate_64Pbps:  {mult: 64, unit: "Pbps"},
	Rate_256Pbps: {mult: 256, unit: "Pbps"},
}

func (qfd *QosFlowDesc) Validate(numOfParams uint8) error {
	if QFD_Del == qfd.OpCode &&
		QfdEbit_NoParamList != qfd.EBit {
		return errors.Errorf("QosFlowDescs: OpCode==del, but Ebit != 0")
	}
	if QFD_Del == qfd.OpCode &&
		QfdEbit_NoParamList == qfd.EBit &&
		numOfParams > 0 {
		return errors.Errorf("QosFlowDescs: OpCode==Del, but numOfParams > 0")
	}
	if QFD_Create == qfd.OpCode {
		if QfdEbit_HasParamList == qfd.EBit && numOfParams == 0 {
			return errors.Errorf("QosFlowDescs: OpCode==Create, ebit=true, but numOfParams == 0")
		}
		if QfdEbit_NoParamList == qfd.EBit && numOfParams != 0 {
			return errors.Errorf("QosFlowDescs: OpCode==Create, ebit=false, but numOfParams != 0")
		}
	}
	if QFD_Mod == qfd.OpCode &&
		numOfParams == 0 {
		return errors.Errorf("QosFlowDescs: OpCode==Mod, but numOfParams == 0")
	}
	return nil
}

func (d *QosFlowDesc) unmarshalParams(b []byte) (int, error) {
	ttlLen := len(b)
	ofs := 0

	minParamCtntLen := 3
	// for ofs+minParamCtntLen <= ttlLen {
	if ofs+minParamCtntLen > ttlLen {
		return 0, errors.Errorf("QFD unmarshalParams() bad ofs(%d) + ttlLen=%d",
			ofs, ttlLen)
	}
	length := int(b[ofs+1])
	if ofs+length > ttlLen {
		return 0, errors.Errorf("QFD unmarshalParams() bad ofs(%d) + length (%d), ttlLen=%d",
			ofs, length, ttlLen)
	}
	switch QfdParamId(b[ofs]) {
	case QfdParamId_5QI:
		if length != 1 {
			return 0, errors.Errorf("QFD unmarshalParams() 5QI bad length %d", length)
		}
		d.FiveQI = b[ofs+2]
		ofs += 2 + length
	case QfdParamId_GFBR_Uplink:
		if length != 3 {
			return 0, errors.Errorf("QFD unmarshalParams() GFBRUplink bad length %d", length)
		}
		bru := paramId2Str[BitRateType(b[ofs+2])]
		val := uint32(binary.BigEndian.Uint16(b[ofs+3 : ofs+5]))
		d.GFBRUplink = fmt.Sprintf("%d %s", val*bru.mult, bru.unit)
		ofs += 2 + length
	case QfdParamId_GFBR_Downlink:
		if length != 3 {
			return 0, errors.Errorf("QFD unmarshalParams() GFBRDownlink bad length %d", length)
		}
		bru := paramId2Str[BitRateType(b[ofs+2])]
		val := uint32(binary.BigEndian.Uint16(b[ofs+3 : ofs+5]))
		d.GFBRDownlink = fmt.Sprintf("%d %s", val*bru.mult, bru.unit)
		ofs += 2 + length
	case QfdParamId_MFBR_Uplink:
		if length != 3 {
			return 0, errors.Errorf("QFD unmarshalParams() MFBRUplink bad length %d", length)
		}
		bru := paramId2Str[BitRateType(b[ofs+2])]
		val := uint32(binary.BigEndian.Uint16(b[ofs+3 : ofs+5]))
		d.MFBRUplink = fmt.Sprintf("%d %s", val*bru.mult, bru.unit)
		ofs += 2 + length
	case QfdParamId_MFBR_Downlink:
		if length != 3 {
			return 0, errors.Errorf("QFD unmarshalParams() MFBRDownlink bad length %d", length)
		}
		bru := paramId2Str[BitRateType(b[ofs+2])]
		val := uint32(binary.BigEndian.Uint16(b[ofs+3 : ofs+5]))
		d.MFBRDownlink = fmt.Sprintf("%d %s", val*bru.mult, bru.unit)
		ofs += 2 + length
	case QfdParamId_AveragingWindow:
		if length != 2 {
			return 0, errors.Errorf("QFD unmarshalParams() AvgWin bad length %d", length)
		}
		d.AvgWin = binary.BigEndian.Uint16(b[ofs+2 : ofs+4])
		ofs += 2 + length
	case QfdParamId_EPSBearerId:
		if length != 1 {
			return 0, errors.Errorf("QFD unmarshalParams() EPSBearerId bad length %d", length)
		}
		if d.EPSBearerIds.NumOfBearerId >= MAX_EPSBearerId {
			// this condition check should be in "Descs" (not "Desc")
			// The total number of EPS bearer identities included
			// in all QoS flow descriptions of a UE cannot exceed fifteen.
			return 0, errors.Errorf("QFD unmarshalParams() too many EPSBearerId")
		}
		d.EPSBearerIds.Id[d.EPSBearerIds.NumOfBearerId] = Get4Bits85(b[ofs+2])
		d.EPSBearerIds.NumOfBearerId++
		ofs += 2 + length
	default:
		// If the parameters list contains a parameter identifier that is not supported by
		// the receiving entity the corresponding parameter shall be discarded.
		ofs += 2 + length
	}
	return ofs, nil
}

func (i *QosFlowDescs) unmarshalQfd(b []byte) (int, error) {
	if len(b) < 3 {
		return 0, errors.Errorf("unmarshalQfd: bad len(b)=%d", len(b))
	}
	qfd := &QosFlowDesc{}
	qfd.QFI = Get6Bits61(b[0])
	qfd.OpCode = QfdOpCode(Get3Bits86(b[1]))
	qfd.EBit = GetBit7(b[2])
	numOfParams := Get6Bits61(b[2])
	ofs := 3
	j := uint8(0)

	if (QFD_Del == qfd.OpCode && QfdEbit_NoParamList == qfd.EBit) || (numOfParams == 0) {
		goto OUT
	} else if len(b)-ofs < 3*int(numOfParams) {
		return 0, errors.Errorf("unmarshalQfd: bad length (%d), #param=%d %v", len(b), numOfParams, b)
	}
	for ; j < numOfParams && ofs < len(b); j++ {
		if umlen, err := qfd.unmarshalParams(b[ofs:]); nil != err {
			qfd = nil
			return 0, errors.Wrap(err, "QosFlowDescs unmarshalQfd()")
		} else {
			ofs += umlen
		}
	}
	if j != numOfParams {
		return 0, errors.Errorf("unmarshalQfd: bad numOfParams, Got=%d, Exp=%d", j, numOfParams)
	}

OUT:
	if err := qfd.Validate(numOfParams); err != nil {
		return 0, errors.Wrap(err, "QosFlowDescs unmarshalQfd()")
	}
	i.Descs = append(i.Descs, *qfd)
	return ofs, nil
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *QosFlowDescs) UnmarshalBinary(b []byte) error {
	ofs := 0 // offset
	for ofs < len(b) {
		if umlen, err := i.unmarshalQfd(b[ofs:]); nil != err {
			i.Descs = nil
			return errors.Wrap(err, "QosFlowDescs UnmarshalBinary()")
		} else {
			ofs += umlen
		}
	}
	return nil
}

func (d *QosFlowDesc) marshalQfdParam(b []byte) (int, error) {
	ofs := 0
	if b == nil {
		return 0, errors.Errorf("marshalQfdParam b is nil")
	}
	if d.FiveQI != uint8(Qfd_Reserved1) {
		b[ofs] = byte(QfdParamId_5QI)
		b[ofs+1] = 1
		b[ofs+2] = d.FiveQI
		ofs += 3
	}
	if d.GFBRUplink != "" {
		b[ofs] = byte(QfdParamId_GFBR_Uplink)
		b[ofs+1] = 3
		if br, err := parseBitRate(d.GFBRUplink); err != nil {
			return 0, errors.Wrap(err, "QosFlowDesc marshalQfdParam GFBRUplink")
		} else {
			copy(b[ofs+2:ofs+5], br)
			ofs += 5
		}
	}
	if d.GFBRDownlink != "" {
		b[ofs] = byte(QfdParamId_GFBR_Downlink)
		b[ofs+1] = 3
		if br, err := parseBitRate(d.GFBRDownlink); err != nil {
			return 0, errors.Wrap(err, "QosFlowDesc marshalQfdParam GFBRDownlink")
		} else {
			copy(b[ofs+2:ofs+5], br)
			ofs += 5
		}
	}
	if d.MFBRUplink != "" {
		b[ofs] = byte(QfdParamId_MFBR_Uplink)
		b[ofs+1] = 3
		if br, err := parseBitRate(d.MFBRUplink); err != nil {
			return 0, errors.Wrap(err, "QosFlowDesc marshalQfdParam MFBRUplink")
		} else {
			copy(b[ofs+2:ofs+5], br)
			ofs += 5
		}
	}
	if d.MFBRDownlink != "" {
		b[ofs] = byte(QfdParamId_MFBR_Downlink)
		b[ofs+1] = 3
		if br, err := parseBitRate(d.MFBRDownlink); err != nil {
			return 0, errors.Wrap(err, "QosFlowDesc marshalQfdParam MFBRDownlink")
		} else {
			copy(b[ofs+2:ofs+5], br)
			ofs += 5
		}
	}
	if d.AvgWin != 0 {
		b[ofs] = byte(QfdParamId_AveragingWindow)
		b[ofs+1] = 2
		binary.BigEndian.PutUint16(b[ofs+2:ofs+4], d.AvgWin)
		ofs += 4
	}
	for i := 0; i < int(d.EPSBearerIds.NumOfBearerId); i++ {
		b[ofs] = byte(QfdParamId_EPSBearerId)
		b[ofs+1] = 1
		b[ofs+2] = Set4Bits85(b[ofs+2], d.EPSBearerIds.Id[i])
		ofs += 3
	}
	return ofs, nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *QosFlowDescs) MarshalBinary() ([]byte, error) {
	// malloc the buffer
	Len := 0
	for j := range i.Descs {
		Len += 3
		Len += int(i.Descs[j].calcParamsMarshalLength())
	}
	b := make([]byte, Len)

	// Fill the buffer
	ofs := 0
	for j := range i.Descs {
		qfd := i.Descs[j]
		numOfParams := qfd.calcNumOfParams()

		b[ofs+0] = Set6Bits61(b[ofs+0], qfd.QFI)
		b[ofs+1] = Set3Bits86(b[ofs+1], uint8(qfd.OpCode))
		// E-bit
		if numOfParams > 0 || qfd.OpCode == QFD_Mod {
			b[ofs+2] = SetBit7(b[ofs+2], 1)
		}
		b[ofs+2] = Set6Bits61(b[ofs+2], numOfParams)
		ofs += 3
		if offset, err := qfd.marshalQfdParam(b[ofs:]); err != nil {
			return nil, errors.Wrap(err, "QosFlowDescs MarshalBinary()")
		} else {
			ofs += offset
		}
	}
	return b, nil
}

func (nasQfd *QosFlowDescs) String() string {
	var brief string
	var operation string
	if nasQfd == nil || len(nasQfd.Descs) == 0 {
		return ""
	}

	op := func(opCode QfdOpCode) string {
		switch opCode {
		case QFD_Create:
			return "add"
		case QFD_Del:
			return "del"
		case QFD_Mod:
			return "mod"
		}
		return "BadOpCode"
	}

	i := 0
	for _, desc := range nasQfd.Descs {
		if desc.OpCode == QFD_None {
			continue
		}
		if i != 0 {
			brief += ","
		}
		i++
		operation = op(desc.OpCode)
		brief += fmt.Sprintf("%s QFD[QFI:%d",
			operation, desc.QFI)
		if desc.FiveQI != 0 {
			brief += fmt.Sprintf(",5QI:%v", desc.FiveQI)
		}
		if desc.GFBRUplink != "" || desc.GFBRDownlink != "" ||
			desc.MFBRUplink != "" || desc.MFBRDownlink != "" {
			brief += fmt.Sprintf(",GBR(UL:%s,DL:%s),MBR(UL:%s,DL:%s)",
				noBps(desc.GFBRUplink), noBps(desc.GFBRDownlink),
				noBps(desc.MFBRUplink), noBps(desc.MFBRDownlink))
		}
		brief += "]"
	}

	return brief
}

func noBps(bps string) string {
	if strings.HasSuffix(bps, "bps") {
		return strings.ReplaceAll(bps[:len(bps)-3], " ", "")
	}
	return bps
}
