package ie

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

// SessAMBR is detailed in 9.11.4.14 Session-AMBR, 24.501
type SessAMBR struct {
	// Name, uint8, Bits, Octet
	UnitDownlink  BitRateType // 8 -> 1 ,   3 -> 3
	ValueDownlink uint16      // 8 -> 1 ,   4 -> 5
	UnitUplink    BitRateType // 8 -> 1 ,   6 -> 6
	ValueUplink   uint16      // 8 -> 1 ,   7 -> 8
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SessAMBR) UnmarshalBinary(b []byte) error {
	if len(b) != 6 {
		return errors.Errorf("SessAMBR bad IE len(%d)", len(b))
	}
	i.UnitDownlink = BitRateType(b[0])
	i.ValueDownlink = binary.BigEndian.Uint16(b[1:3])
	i.UnitUplink = BitRateType(b[3])
	i.ValueUplink = binary.BigEndian.Uint16(b[4:6])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SessAMBR) MarshalBinary() ([]byte, error) {
	b := make([]byte, 6)
	b[0] = uint8(i.UnitDownlink)
	binary.BigEndian.PutUint16(b[1:3], i.ValueDownlink)
	b[3] = uint8(i.UnitUplink)
	binary.BigEndian.PutUint16(b[4:6], i.ValueUplink)

	return b, nil
}

// Set() Example format of ambrUplink / ambrDownlink:
// - 1024 Mbps / 1024 Mbps
// - 333 Gbps
func (i *SessAMBR) Set(ambrUplink, ambrDownlink string) error {
	if b, err := parseBitRate(ambrUplink); err != nil {
		return errors.Wrap(err, "SessAMBR Uplink Set()")
	} else {
		// fmt.Printf("ul:%v, b=%v\n", ambrUplink, b)
		i.UnitUplink = BitRateType(b[0])
		i.ValueUplink = binary.BigEndian.Uint16(b[1:3])
	}
	if b, err := parseBitRate(ambrDownlink); err != nil {
		return errors.Wrap(err, "SessAMBR Uplink Set()")
	} else {
		// fmt.Printf("dl:%v, b=%v\n", ambrDownlink, b)
		i.UnitDownlink = BitRateType(b[0])
		i.ValueDownlink = binary.BigEndian.Uint16(b[1:3])
	}
	return nil
}

func (ambr *SessAMBR) String() string {
	if ambr == nil {
		return ""
	}
	brief := fmt.Sprintf("SessAMBR[UL:%d*%s,DL:%d*%s]",
		ambr.ValueUplink, BitRateStr(ambr.UnitUplink),
		ambr.ValueDownlink, BitRateStr(ambr.UnitDownlink))
	return brief
}
