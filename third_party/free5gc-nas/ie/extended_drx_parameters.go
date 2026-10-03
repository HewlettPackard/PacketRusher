package ie

import (
	"fmt"

	"github.com/pkg/errors"
)

// ExtendedDRXParams is detailed in 10.5.5.32 Extended DRX parameters, 24.008
type ExtendedDRXParams struct {
	// Name, uint8, Bits, Octet
	PagingTimeWindow NRPagingTimeWindow // 8 -> 5, 3 -> 3
	EDRXValue        NReDRXValue        // 4 -> 1, 3 -> 3
}

func (i *ExtendedDRXParams) String() string {
	return fmt.Sprintf(
		"PagingTimeWindow: %s; eDRXValue: %s",
		i.PagingTimeWindow, i.EDRXValue)
}

// String returns a human-readable representation of the NRPagingTimeWindow constant
func (p NRPagingTimeWindow) String() string {
	switch p {
	case NRPagingTimeWindow1_28s:
		return "1.28s"
	case NRPagingTimeWindow2_56s:
		return "2.56s"
	case NRPagingTimeWindow3_84s:
		return "3.84s"
	case NRPagingTimeWindow5_12s:
		return "5.12s"
	case NRPagingTimeWindow6_4s:
		return "6.4s"
	case NRPagingTimeWindow7_68s:
		return "7.68s"
	case NRPagingTimeWindow8_96s:
		return "8.96s"
	case NRPagingTimeWindow10_24s:
		return "10.24s"
	case NRPagingTimeWindow11_52s:
		return "11.52s"
	case NRPagingTimeWindow12_8s:
		return "12.8s"
	case NRPagingTimeWindow14_08s:
		return "14.08s"
	case NRPagingTimeWindow15_36s:
		return "15.36s"
	case NRPagingTimeWindow16_64s:
		return "16.64s"
	case NRPagingTimeWindow17_92s:
		return "17.92s"
	case NRPagingTimeWindow19_2s:
		return "19.2s"
	case NRPagingTimeWindow20_48s:
		return "20.48s"
	default:
		return fmt.Sprintf("Unknown(%d)", p)
	}
}

// Different RAT may have different codes for PagingTimeWindow.
// We consider only NR for now.
const (
	NRPagingTimeWindow1_28s  NRPagingTimeWindow = 0x00
	NRPagingTimeWindow2_56s  NRPagingTimeWindow = 0x01
	NRPagingTimeWindow3_84s  NRPagingTimeWindow = 0x02
	NRPagingTimeWindow5_12s  NRPagingTimeWindow = 0x03
	NRPagingTimeWindow6_4s   NRPagingTimeWindow = 0x04
	NRPagingTimeWindow7_68s  NRPagingTimeWindow = 0x05
	NRPagingTimeWindow8_96s  NRPagingTimeWindow = 0x06
	NRPagingTimeWindow10_24s NRPagingTimeWindow = 0x07
	NRPagingTimeWindow11_52s NRPagingTimeWindow = 0x08
	NRPagingTimeWindow12_8s  NRPagingTimeWindow = 0x09
	NRPagingTimeWindow14_08s NRPagingTimeWindow = 0x0a
	NRPagingTimeWindow15_36s NRPagingTimeWindow = 0x0b
	NRPagingTimeWindow16_64s NRPagingTimeWindow = 0x0c
	NRPagingTimeWindow17_92s NRPagingTimeWindow = 0x0d
	NRPagingTimeWindow19_2s  NRPagingTimeWindow = 0x0e
	NRPagingTimeWindow20_48s NRPagingTimeWindow = 0x0f
)

// NRPagingTimeWindow represents NR paging time window values
type NRPagingTimeWindow uint8

// According to 3GPP TS 24.008, eDRX value implies both the eDRX cycle length duration
// and the eDRX cycle parameter 'TeDRX' as defined in 3GPP TS 38.304 [183].
// Here, we use the eDRX cycle length duration value to define a more human-readable
// variable names for the eDRX Value constants.
// Mapped values for 'TeDRX' can be retrieved using TeDRXValueMap.

// String returns a human-readable representation of the NReDRXValue constant
func (e NReDRXValue) String() string {
	teDRX := TeDRXValueMap[e]
	switch e {
	case NReDRXValue2_56s:
		return fmt.Sprintf("Len:2.56s, TeDRX:%s", teDRX)
	case NReDRXValue5_12s:
		return fmt.Sprintf("Len:5.12s, TeDRX:%s", teDRX)
	case NReDRXValue10_24s:
		return fmt.Sprintf("Len:10.24s, TeDRX:%s", teDRX)
	case NReDRXValue20_48s:
		return fmt.Sprintf("Len:20.48s, TeDRX:%s", teDRX)
	case NReDRXValue40_96s:
		return fmt.Sprintf("Len:40.96s, TeDRX:%s", teDRX)
	case NReDRXValue81_92s:
		return fmt.Sprintf("Len:81.92s, TeDRX:%s", teDRX)
	case NReDRXValue163_84s:
		return fmt.Sprintf("Len:163.84s, TeDRX:%s", teDRX)
	case NReDRXValue327_68s:
		return fmt.Sprintf("Len:327.68s, TeDRX:%s", teDRX)
	case NReDRXValue655_36s:
		return fmt.Sprintf("Len:655.36s, TeDRX:%s", teDRX)
	case NReDRXValue1310_72s:
		return fmt.Sprintf("Len:1310.72s, TeDRX:%s", teDRX)
	case NReDRXValue2621_44s:
		return fmt.Sprintf("Len:2621.44s, TeDRX:%s", teDRX)
	case NReDRXValue5242_88s:
		return fmt.Sprintf("Len:5242.88s, TeDRX:%s", teDRX)
	case NReDRXValue10485_76s:
		return fmt.Sprintf("Len:10485.76s, TeDRX:%s", teDRX)
	default:
		return fmt.Sprintf("Unknown(%d)", e)
	}
}

// Different RAT may have different codes for eDRXValue.
// We consider only NR for now.
const (
	NReDRXValue2_56s     NReDRXValue = 0x00
	NReDRXValue5_12s     NReDRXValue = 0x01
	NReDRXValue10_24s    NReDRXValue = 0x02
	NReDRXValue20_48s    NReDRXValue = 0x03
	NReDRXValue40_96s    NReDRXValue = 0x04
	NReDRXValue81_92s    NReDRXValue = 0x05
	NReDRXValue163_84s   NReDRXValue = 0x06
	NReDRXValue327_68s   NReDRXValue = 0x07
	NReDRXValue655_36s   NReDRXValue = 0x08
	NReDRXValue1310_72s  NReDRXValue = 0x09
	NReDRXValue2621_44s  NReDRXValue = 0x0a
	NReDRXValue5242_88s  NReDRXValue = 0x0b
	NReDRXValue10485_76s NReDRXValue = 0x0c
)

// NReDRXValue represents NR eDRX values
type NReDRXValue uint8

const (
	NRTeDRXNotUsed  NRTeDRXValue = 0x00
	NRTeDRXTwoPow0  NRTeDRXValue = 1 << 0  // 2^0 = 1
	NRTeDRXTwoPow1  NRTeDRXValue = 1 << 1  // 2^1 = 2
	NRTeDRXTwoPow2  NRTeDRXValue = 1 << 2  // 2^2 = 4
	NRTeDRXTwoPow3  NRTeDRXValue = 1 << 3  // 2^3 = 8
	NRTeDRXTwoPow4  NRTeDRXValue = 1 << 4  // 2^4 = 16
	NRTeDRXTwoPow5  NRTeDRXValue = 1 << 5  // 2^5 = 32
	NRTeDRXTwoPow6  NRTeDRXValue = 1 << 6  // 2^6 = 64
	NRTeDRXTwoPow7  NRTeDRXValue = 1 << 7  // 2^7 = 128
	NRTeDRXTwoPow8  NRTeDRXValue = 1 << 8  // 2^8 = 256
	NRTeDRXTwoPow9  NRTeDRXValue = 1 << 9  // 2^9 = 512
	NRTeDRXTwoPow10 NRTeDRXValue = 1 << 10 // 2^10 = 1024
)

// NRTeDRXValue represents NR TeDRX values
type NRTeDRXValue uint16

// String returns a human-readable representation of the NRTeDRXValue constant
func (t NRTeDRXValue) String() string {
	switch t {
	case NRTeDRXNotUsed:
		return "NotUsed"
	case NRTeDRXTwoPow0:
		return "2^0=1"
	case NRTeDRXTwoPow1:
		return "2^1=2"
	case NRTeDRXTwoPow2:
		return "2^2=4"
	case NRTeDRXTwoPow3:
		return "2^3=8"
	case NRTeDRXTwoPow4:
		return "2^4=16"
	case NRTeDRXTwoPow5:
		return "2^5=32"
	case NRTeDRXTwoPow6:
		return "2^6=64"
	case NRTeDRXTwoPow7:
		return "2^7=128"
	case NRTeDRXTwoPow8:
		return "2^8=256"
	case NRTeDRXTwoPow9:
		return "2^9=512"
	case NRTeDRXTwoPow10:
		return "2^10=1024"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

var TeDRXValueMap = map[NReDRXValue]NRTeDRXValue{
	// For NR connected to 5GCN, eDRX cycle length durations of 2,56 seconds
	// or 5,12 seconds the eDRX cycle parameter 'TeDRX' is not used as a different
	// algorithm compared to the other values is applied. See
	// 3GPP TS 38.304 [183] for details.
	NReDRXValue2_56s: NRTeDRXNotUsed,
	NReDRXValue5_12s: NRTeDRXNotUsed,

	NReDRXValue10_24s:    NRTeDRXTwoPow0,
	NReDRXValue20_48s:    NRTeDRXTwoPow1,
	NReDRXValue40_96s:    NRTeDRXTwoPow2,
	NReDRXValue81_92s:    NRTeDRXTwoPow3,
	NReDRXValue163_84s:   NRTeDRXTwoPow4,
	NReDRXValue327_68s:   NRTeDRXTwoPow5,
	NReDRXValue655_36s:   NRTeDRXTwoPow6,
	NReDRXValue1310_72s:  NRTeDRXTwoPow7,
	NReDRXValue2621_44s:  NRTeDRXTwoPow8,
	NReDRXValue5242_88s:  NRTeDRXTwoPow9,
	NReDRXValue10485_76s: NRTeDRXTwoPow10,
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ExtendedDRXParams) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("ExtendedDRXParams buffer len(%d) != 1", len(b))
	}
	// 4-bit must <= NReDRXValue10485_76s, no further check is needed.
	i.PagingTimeWindow = NRPagingTimeWindow(Get4Bits85(b[0]))
	if eDRXValue := NReDRXValue(Get4Bits41(b[0])); eDRXValue > NReDRXValue10485_76s {
		// TS 24.008
		// All other values shall be interpreted as 0000 by this version of the protocol.
		i.EDRXValue = NReDRXValue2_56s
	} else {
		i.EDRXValue = eDRXValue
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ExtendedDRXParams) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set4Bits85(b[0], uint8(i.PagingTimeWindow))
	b[0] = Set4Bits41(b[0], uint8(i.EDRXValue))
	return b, nil
}
