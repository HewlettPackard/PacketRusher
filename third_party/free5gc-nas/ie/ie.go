package ie

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

type IE interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
}

type IEToDo struct {
	IEName string
}

func (e *IEToDo) Error() string {
	return fmt.Sprintf("%s: IE not implemented yet", e.IEName)
}

type BitIndex uint8

const (
	BIT1 BitIndex = iota + 1
	BIT2
	BIT3
	BIT4
	BIT5
	BIT6
	BIT7
	BIT8
	BIT9
	BIT10
	BIT11
	BIT12
	BIT13
	BIT14
	BIT15
	BIT16
)

const (
	SpareHalfOctet uint8 = 0x00
)

func bool2uint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

/** 6 bits operation */
func Get6Bits61(b uint8) uint8 {
	return (b & 0x3F)
}

func Set6Bits61(b, b61 uint8) uint8 {
	return (b & 0xC0) + (b61 & 0x3F)
}

func Get5Bits51(b uint8) uint8 {
	return (b & 0x1F)
}

func Set5Bits51(b, b51 uint8) uint8 {
	return (b & 0xE0) + (b51 & 0x1F)
}

/** 4 bits operation */
func GetHalfValue(b uint8) (half85, half41 uint8) {
	half41 = b & 0x0f
	half85 = b >> 4
	return half85, half41
}

func SetHalfValue(half85, half41 uint8) uint8 {
	return (half85 << 4) + (half41 & 0x0f)
}

func Get4Bits41(b uint8) uint8 {
	return (b & 0x0f)
}

func Set4Bits41(b, half41 uint8) uint8 {
	return (b & 0xf0) + (half41 & 0x0f)
}

func Get4Bits74(b uint8) uint8 {
	return (b & 0x78) >> 3
}

func Set4Bits74(b, val uint8) uint8 {
	return (b & 0x87) + ((val & 0xf) << 3)
}

func Get4Bits85(b uint8) uint8 {
	return (b & 0xf0) >> 4
}

func Set4Bits85(b, half85 uint8) uint8 {
	return (b & 0x0f) + (half85 << 4)
}

/** 3 bits operation */
func Get3Bits31(b uint8) uint8 {
	return (b & 0x07)
}

func Set3Bits31(b, b31 uint8) uint8 {
	return (b & 0xf8) + (b31 & 0x07)
}

func Get3Bits75(b uint8) uint8 {
	return (b & 0x70) >> 4
}

func Set3Bits75(b, b75 uint8) uint8 {
	return (b & 0x8F) + ((b75 & 0x07) << 4)
}

func Get3Bits86(b uint8) uint8 {
	return (b & 0xe0) >> 5
}

func Set3Bits86(b, b86 uint8) uint8 {
	return (b & 0x1F) + ((b86 & 0x07) << 5)
}

/** 2 bits operation */
func Get2Bits21(b uint8) uint8 {
	return (b & 0x03) >> 0
}

// to add unit test
// func Get2Bits32(b uint8) uint8 {
// 	return (b & 0x06) >> 1
// }

func Get2Bits43(b uint8) uint8 {
	return (b & 0x0c) >> 2
}

// func Get2Bits54(b uint8) uint8 {
// 	return (b & 0x18) >> 3
// }

func Get2Bits65(b uint8) uint8 {
	return (b & 0x30) >> 4
}

func Get2Bits76(b uint8) uint8 {
	return (b & 0x60) >> 5
}

// to add unit test
// func Get2Bits87(b uint8) uint8 {
// 	return (b & 0xc0) >> 6
// }

func Set2Bits21(b, b2 uint8) uint8 {
	return (b & 0xfc) + ((b2 & 0x03) << 0)
}

/*
to add unit test

func Set2Bits32(b, b3 uint8) uint8 {
        return (b & 0xf9) + ((b3 & 0x03) << 1)
}
*/

func Set2Bits43(b, b4 uint8) uint8 {
	return (b & 0xf3) + ((b4 & 0x03) << 2)
}

/*
func Set2Bits54(b, b5 uint8) uint8 {
        return (b & 0xe7) + ((b5 & 0x03) << 3)
}
*/

func Set2Bits65(b, b6 uint8) uint8 {
	return (b & 0xcf) + ((b6 & 0x03) << 4)
}

func Set2Bits76(b, b7 uint8) uint8 {
	return (b & 0x9f) + ((b7 & 0x03) << 5)
}

/*
to add unit test
func Set2Bits87(b, b8 uint8) uint8 {
        return (b & 0x3f) + ((b8 & 0x03) << 6)
}
*/

/** 1 bit operation */
func GetBit1(b uint8) uint8 {
	return (b & 0x01) >> 0
}

func SetBit1(b, b1 uint8) uint8 {
	return (b & 0xfe) + ((b1 & 0x01) << 0)
}

func GetBit2(b uint8) uint8 {
	return (b & 0x02) >> 1
}

func SetBit2(b, b2 uint8) uint8 {
	return (b & 0xfd) + ((b2 & 0x01) << 1)
}

func GetBit3(b uint8) uint8 {
	return (b & 0x04) >> 2
}

func SetBit3(b, b3 uint8) uint8 {
	return (b & 0xfb) + ((b3 & 0x01) << 2)
}

func GetBit4(b uint8) uint8 {
	return (b & 0x08) >> 3
}

func SetBit4(b, b4 uint8) uint8 {
	return (b & 0xf7) + ((b4 & 0x01) << 3)
}

func GetBit5(b uint8) uint8 {
	return (b & 0x10) >> 4
}

func SetBit5(b, b5 uint8) uint8 {
	return (b & 0xef) + ((b5 & 0x01) << 4)
}

func GetBit6(b uint8) uint8 {
	return (b & 0x20) >> 5
}

func SetBit6(b, b6 uint8) uint8 {
	return (b & 0xdf) + ((b6 & 0x01) << 5)
}

func GetBit7(b uint8) uint8 {
	return (b & 0x40) >> 6
}

func SetBit7(b, b7 uint8) uint8 {
	return (b & 0xbf) + ((b7 & 0x01) << 6)
}

func GetBit8(b uint8) uint8 {
	return (b & 0x80) >> 7
}

func SetBit8(b, b8 uint8) uint8 {
	return (b & 0x7f) + ((b8 & 0x01) << 7)
}

/** multiple bits operation */
func GetHalfIEValue(b []byte) ([]byte, []byte) {
	if len(b) < 1 {
		return nil, nil
	}
	half41, err := GetMaskedValue8(b[0], BIT4, BIT1)
	if err != nil {
		return nil, nil
	}
	half85, err := GetMaskedValue8(b[0], BIT8, BIT5)
	if err != nil {
		return nil, nil
	}
	return []byte{half41}, []byte{half85}
}

func GetMaskedValue8(b uint8, end, start BitIndex) (uint8, error) {
	shiftL := 8 - uint8(end)
	if end > 8 || start > 8 {
		return uint8(0), errors.Errorf("GetMaskedValue8 end > 8 or start > 8")
	}
	if end < start {
		return uint8(0), errors.Errorf("GetMaskedValue8 end < start")
	} else if end > start {
		return ((b << shiftL) >> (shiftL + uint8(start-1))), nil
	}
	return ((b >> (start - 1)) & 1), nil
}

func GetMaskedValue16(b []byte, end, start BitIndex) (uint16, error) {
	if len(b) != 2 {
		return 0, errors.Errorf("GetMaskedValue16 len(b) == %d, != 2", len(b))
	}
	shiftL := 16 - uint8(end)
	out := uint16(b[0])<<8 + uint16(b[1])
	if end < start {
		return uint16(0), errors.Errorf("GetMaskedValue16 end < start")
	} else if end > start {
		return ((out << shiftL) >> (shiftL + uint8(start-1))), nil
	}
	return ((out >> (start - 1)) & 1), nil
}

func SetMaskedValue8(b []uint8, v uint8, end, start BitIndex) error {
	if end > 8 || start > 8 {
		return errors.Errorf("GetMaskedValue8 end > 8 or start > 8")
	}
	if end < start {
		return errors.Errorf("SetMaskedValue8 end < start")
	}
	shiftL := uint8(start) - 1
	maxV := uint8(1<<uint(end-start+1)) - 1
	if v > maxV {
		return errors.Errorf("SetMaskedValue8 value > 0x%x", maxV)
	}
	b[0] = ((b[0] &^ (maxV << shiftL)) | (v << shiftL))
	return nil
}

func SetMaskedValue16(b []uint8, v uint16, end, start BitIndex) error {
	if len(b) != 2 {
		return errors.Errorf("SetMaskedValue16 len(b)==%d, != 2", len(b))
	}
	if end < start {
		return errors.Errorf("SetMaskedValue16 end < start")
	}
	if end > 16 || start > 16 {
		return errors.Errorf("GetMaskedValue16 end > 16 or start > 16")
	}

	shiftL := uint8(start) - 1
	maxV := uint16(1<<uint(end-start+1)) - 1
	if v > maxV {
		return errors.Errorf("SetMaskedValue16 value > 0x%x", maxV)
	}

	out := binary.BigEndian.Uint16(b)
	out = ((out &^ (maxV << shiftL)) | (v << shiftL))
	binary.BigEndian.PutUint16(b, out)
	return nil
}
