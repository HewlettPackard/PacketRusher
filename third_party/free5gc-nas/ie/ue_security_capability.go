package ie

import (
	"github.com/pkg/errors"
)

// UESecCapability is detailed in 9.11.3.54 UE security capability, 24.501
type UESecCapability struct {
	Length int // value should be 2~8, even number

	// Name, bool, Bits, Octet
	EA05G      bool // 8 -> 8 ,   3 -> 3
	EA1_128_5G bool // 7 -> 7 ,   3 -> 3
	EA2_128_5G bool // 6 -> 6 ,   3 -> 3
	EA3_128_5G bool // 5 -> 5 ,   3 -> 3
	EA45G      bool // 4 -> 4 ,   3 -> 3
	EA55G      bool // 3 -> 3 ,   3 -> 3
	EA65G      bool // 2 -> 2 ,   3 -> 3
	EA75G      bool // 1 -> 1 ,   3 -> 3
	IA05G      bool // 8 -> 8 ,   4 -> 4
	IA1_128_5G bool // 7 -> 7 ,   4 -> 4
	IA2_128_5G bool // 6 -> 6 ,   4 -> 4
	IA3_128_5G bool // 5 -> 5 ,   4 -> 4
	IA45G      bool // 4 -> 4 ,   4 -> 4
	IA55G      bool // 3 -> 3 ,   4 -> 4
	IA65G      bool // 2 -> 2 ,   4 -> 4
	IA75G      bool // 1 -> 1 ,   4 -> 4
	EEA0       bool // 8 -> 8 ,   5 -> 5
	EEA1_128   bool // 7 -> 7 ,   5 -> 5
	EEA2_128   bool // 6 -> 6 ,   5 -> 5
	EEA3_128   bool // 5 -> 5 ,   5 -> 5
	EEA4       bool // 4 -> 4 ,   5 -> 5
	EEA5       bool // 3 -> 3 ,   5 -> 5
	EEA6       bool // 2 -> 2 ,   5 -> 5
	EEA7       bool // 1 -> 1 ,   5 -> 5
	EIA0       bool // 8 -> 8 ,   6 -> 6
	EIA1_128   bool // 7 -> 7 ,   6 -> 6
	EIA2_128   bool // 6 -> 6 ,   6 -> 6
	EIA3_128   bool // 5 -> 5 ,   6 -> 6
	EIA4       bool // 4 -> 4 ,   6 -> 6
	EIA5       bool // 3 -> 3 ,   6 -> 6
	EIA6       bool // 2 -> 2 ,   6 -> 6
	EIA7       bool // 1 -> 1 ,   6 -> 6
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UESecCapability) UnmarshalBinary(b []byte) error {
	i.Length = len(b)

	if (i.Length < 2) || (i.Length > 8) || (i.Length%2 != 0) {
		return errors.Errorf("UESecCapability bad IE len(%d)", i.Length)
	}

	if i.Length >= 2 {
		i.EA05G = GetBit8(b[0]) == 1
		i.EA1_128_5G = GetBit7(b[0]) == 1
		i.EA2_128_5G = GetBit6(b[0]) == 1
		i.EA3_128_5G = GetBit5(b[0]) == 1
		i.EA45G = GetBit4(b[0]) == 1
		i.EA55G = GetBit3(b[0]) == 1
		i.EA65G = GetBit2(b[0]) == 1
		i.EA75G = GetBit1(b[0]) == 1
		i.IA05G = GetBit8(b[1]) == 1
		i.IA1_128_5G = GetBit7(b[1]) == 1
		i.IA2_128_5G = GetBit6(b[1]) == 1
		i.IA3_128_5G = GetBit5(b[1]) == 1
		i.IA45G = GetBit4(b[1]) == 1
		i.IA55G = GetBit3(b[1]) == 1
		i.IA65G = GetBit2(b[1]) == 1
		i.IA75G = GetBit1(b[1]) == 1
	}
	if i.Length >= 4 {
		i.EEA0 = GetBit8(b[2]) == 1
		i.EEA1_128 = GetBit7(b[2]) == 1
		i.EEA2_128 = GetBit6(b[2]) == 1
		i.EEA3_128 = GetBit5(b[2]) == 1
		i.EEA4 = GetBit4(b[2]) == 1
		i.EEA5 = GetBit3(b[2]) == 1
		i.EEA6 = GetBit2(b[2]) == 1
		i.EEA7 = GetBit1(b[2]) == 1
		i.EIA0 = GetBit8(b[3]) == 1
		i.EIA1_128 = GetBit7(b[3]) == 1
		i.EIA2_128 = GetBit6(b[3]) == 1
		i.EIA3_128 = GetBit5(b[3]) == 1
		i.EIA4 = GetBit4(b[3]) == 1
		i.EIA5 = GetBit3(b[3]) == 1
		i.EIA6 = GetBit2(b[3]) == 1
		i.EIA7 = GetBit1(b[3]) == 1
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UESecCapability) MarshalBinary() ([]byte, error) {
	if (i.Length < 2) || (i.Length > 8) || (i.Length%2 != 0) {
		return nil, errors.Errorf("UESecCapability bad IE len(%d)", i.Length)
	}

	b := make([]byte, i.Length)
	b[0] = SetBit8(b[0], bool2uint8(i.EA05G))
	b[0] = SetBit7(b[0], bool2uint8(i.EA1_128_5G))
	b[0] = SetBit6(b[0], bool2uint8(i.EA2_128_5G))
	b[0] = SetBit5(b[0], bool2uint8(i.EA3_128_5G))
	b[0] = SetBit4(b[0], bool2uint8(i.EA45G))
	b[0] = SetBit3(b[0], bool2uint8(i.EA55G))
	b[0] = SetBit2(b[0], bool2uint8(i.EA65G))
	b[0] = SetBit1(b[0], bool2uint8(i.EA75G))

	b[1] = SetBit8(b[1], bool2uint8(i.IA05G))
	b[1] = SetBit7(b[1], bool2uint8(i.IA1_128_5G))
	b[1] = SetBit6(b[1], bool2uint8(i.IA2_128_5G))
	b[1] = SetBit5(b[1], bool2uint8(i.IA3_128_5G))
	b[1] = SetBit4(b[1], bool2uint8(i.IA45G))
	b[1] = SetBit3(b[1], bool2uint8(i.IA55G))
	b[1] = SetBit2(b[1], bool2uint8(i.IA65G))
	b[1] = SetBit1(b[1], bool2uint8(i.IA75G))

	if i.Length >= 4 {
		b[2] = SetBit8(b[2], bool2uint8(i.EEA0))
		b[2] = SetBit7(b[2], bool2uint8(i.EEA1_128))
		b[2] = SetBit6(b[2], bool2uint8(i.EEA2_128))
		b[2] = SetBit5(b[2], bool2uint8(i.EEA3_128))
		b[2] = SetBit4(b[2], bool2uint8(i.EEA4))
		b[2] = SetBit3(b[2], bool2uint8(i.EEA5))
		b[2] = SetBit2(b[2], bool2uint8(i.EEA6))
		b[2] = SetBit1(b[2], bool2uint8(i.EEA7))

		b[3] = SetBit8(b[3], bool2uint8(i.EIA0))
		b[3] = SetBit7(b[3], bool2uint8(i.EIA1_128))
		b[3] = SetBit6(b[3], bool2uint8(i.EIA2_128))
		b[3] = SetBit5(b[3], bool2uint8(i.EIA3_128))
		b[3] = SetBit4(b[3], bool2uint8(i.EIA4))
		b[3] = SetBit3(b[3], bool2uint8(i.EIA5))
		b[3] = SetBit2(b[3], bool2uint8(i.EIA6))
		b[3] = SetBit1(b[3], bool2uint8(i.EIA7))
	}

	return b, nil
}
