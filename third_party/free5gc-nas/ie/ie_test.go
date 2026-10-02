package ie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHalfValue(t *testing.T) {
	h85, h41 := GetHalfValue(0x76)
	assert.Equal(t, uint8(7), h85, "")
	assert.Equal(t, uint8(6), h41, "")
}

func TestSetHalfValue(t *testing.T) {
	out := SetHalfValue(0x07, 0x06)
	assert.Equal(t, uint8(0x76), out, "")

	out = SetHalfValue(0x17, 0x36)
	assert.Equal(t, uint8(0x76), out, "")
}

func TestGet4Bits41(t *testing.T) {
	out := Get4Bits41(0x30)
	assert.Equal(t, uint8(0x0), out, "")

	out = Get4Bits41(0x37)
	assert.Equal(t, uint8(0x7), out, "")

	out = Get4Bits41(0xaf)
	assert.Equal(t, uint8(0xf), out, "")
}

func TestSet4Bits41(t *testing.T) {
	out := Set4Bits41(0x30, 0x06)
	assert.Equal(t, uint8(0x36), out, "")

	out = Set4Bits41(0x37, 0x36)
	assert.Equal(t, uint8(0x36), out, "")
}

func TestGet4Bits74(t *testing.T) {
	out := Get4Bits74(0x30)
	assert.Equal(t, uint8(0x6), out, "")

	out = Get4Bits74(0x37)
	assert.Equal(t, uint8(0x6), out, "")

	out = Get4Bits74(0xaf)
	assert.Equal(t, uint8(0x5), out, "")
}

func TestSet4Bits74(t *testing.T) {
	out := Set4Bits74(0x30, 0x06)
	assert.Equal(t, uint8(0x30), out, "")

	out = Set4Bits74(0x37, 0x36)
	assert.Equal(t, uint8(0x37), out, "")
}

func TestGet4Bits85(t *testing.T) {
	out := Get4Bits85(0x30)
	assert.Equal(t, uint8(0x3), out, "")

	out = Get4Bits85(0x67)
	assert.Equal(t, uint8(0x6), out, "")

	out = Get4Bits85(0xaf)
	assert.Equal(t, uint8(0xa), out, "")
}

func TestSet4Bits85(t *testing.T) {
	out := Set4Bits85(0x30, 0x06)
	assert.Equal(t, uint8(0x60), out, "")

	out = Set4Bits85(0x37, 0x36)
	assert.Equal(t, uint8(0x67), out, "")
}

func TestGet3Bits31(t *testing.T) {
	out := Get3Bits31(0x30)
	assert.Equal(t, uint8(0x0), out, "")

	out = Get3Bits31(0x67)
	assert.Equal(t, uint8(0x7), out, "")

	out = Get3Bits31(0xaf)
	assert.Equal(t, uint8(0x7), out, "")
}

func TestSet3Bits31(t *testing.T) {
	out := Set3Bits31(0x30, 0x07)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set3Bits31(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set3Bits31(0x37, 0xff)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set3Bits31(0x3f, 0x01)
	assert.Equal(t, uint8(0x39), out, "")
}

func TestGet3Bits75(t *testing.T) {
	out := Get3Bits75(0x30)
	assert.Equal(t, uint8(0x3), out, "")

	out = Get3Bits75(0x67)
	assert.Equal(t, uint8(0x6), out, "")

	out = Get3Bits75(0xaf)
	assert.Equal(t, uint8(0x2), out, "")
}

func TestSet3Bits75(t *testing.T) {
	out := Set3Bits75(0x00, 0x07)
	assert.Equal(t, uint8(0x70), out, "")

	out = Set3Bits75(0x30, 0x07)
	assert.Equal(t, uint8(0x70), out, "")

	out = Set3Bits75(0x37, 0x37)
	assert.Equal(t, uint8(0x77), out, "")

	out = Set3Bits75(0x37, 0xff)
	assert.Equal(t, uint8(0x77), out, "")

	out = Set3Bits75(0x3f, 0x01)
	assert.Equal(t, uint8(0x1f), out, "")
}

func TestGet3Bits86(t *testing.T) {
	out := Get3Bits86(0x30)
	assert.Equal(t, uint8(0x1), out, "")

	out = Get3Bits86(0x67)
	assert.Equal(t, uint8(0x3), out, "")

	out = Get3Bits86(0xaf)
	assert.Equal(t, uint8(0x5), out, "")
}

func TestSet3Bits86(t *testing.T) {
	out := Set3Bits86(0x00, 0x07)
	assert.Equal(t, uint8(0xE0), out, "")

	out = Set3Bits86(0x30, 0x07)
	assert.Equal(t, uint8(0xf0), out, "")

	out = Set3Bits86(0x37, 0x37)
	assert.Equal(t, uint8(0xf7), out, "")

	out = Set3Bits86(0x37, 0xff)
	assert.Equal(t, uint8(0xf7), out, "")

	out = Set3Bits86(0x3f, 0x01)
	assert.Equal(t, uint8(0x3f), out, "")
}

func TestGet2Bits21(t *testing.T) {
	out := Get2Bits21(0x07)
	assert.Equal(t, uint8(0x03), out, "")

	out = Get2Bits21(0x30)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits21(0x46)
	assert.Equal(t, uint8(0x02), out, "")

	out = Get2Bits21(0x89)
	assert.Equal(t, uint8(0x01), out, "")

	out = Get2Bits21(0x2f)
	assert.Equal(t, uint8(0x03), out, "")
}

func TestSet2Bits21(t *testing.T) {
	out := Set2Bits21(0x00, 0x07)
	assert.Equal(t, uint8(0x03), out, "")

	out = Set2Bits21(0x30, 0x07)
	assert.Equal(t, uint8(0x33), out, "")

	out = Set2Bits21(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set2Bits21(0x37, 0xff)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set2Bits21(0x3f, 0x01)
	assert.Equal(t, uint8(0x3D), out, "")
}

func TestGet2Bits43(t *testing.T) {
	out := Get2Bits43(0x07)
	assert.Equal(t, uint8(0x01), out, "")

	out = Get2Bits43(0x30)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits43(0x46)
	assert.Equal(t, uint8(0x01), out, "")

	out = Get2Bits43(0x89)
	assert.Equal(t, uint8(0x02), out, "")

	out = Get2Bits43(0x2f)
	assert.Equal(t, uint8(0x03), out, "")
}

func TestSet2Bits43(t *testing.T) {
	out := Set2Bits43(0x00, 0x07)
	assert.Equal(t, uint8(0x0C), out, "")

	out = Set2Bits43(0x30, 0x07)
	assert.Equal(t, uint8(0x3C), out, "")

	out = Set2Bits43(0x37, 0x37)
	assert.Equal(t, uint8(0x3f), out, "")

	out = Set2Bits43(0x37, 0xff)
	assert.Equal(t, uint8(0x3f), out, "")

	out = Set2Bits43(0x3f, 0x01)
	assert.Equal(t, uint8(0x37), out, "")
}

func TestGet2Bits65(t *testing.T) {
	out := Get2Bits65(0x07)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits65(0x30)
	assert.Equal(t, uint8(0x03), out, "")

	out = Get2Bits65(0x46)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits65(0x89)
	assert.Equal(t, uint8(0x0), out, "")

	out = Get2Bits65(0x2f)
	assert.Equal(t, uint8(0x02), out, "")
}

func TestSet2Bits65(t *testing.T) {
	out := Set2Bits65(0x00, 0x07)
	assert.Equal(t, uint8(0x30), out, "")

	out = Set2Bits65(0x30, 0x07)
	assert.Equal(t, uint8(0x30), out, "")

	out = Set2Bits65(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set2Bits65(0x37, 0xff)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set2Bits65(0x3f, 0x01)
	assert.Equal(t, uint8(0x1f), out, "")
}

func TestGet2Bits76(t *testing.T) {
	out := Get2Bits76(0x07)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits76(0x30)
	assert.Equal(t, uint8(0x01), out, "")

	out = Get2Bits76(0x46)
	assert.Equal(t, uint8(0x02), out, "")

	out = Get2Bits76(0x89)
	assert.Equal(t, uint8(0x00), out, "")

	out = Get2Bits76(0x2f)
	assert.Equal(t, uint8(0x01), out, "")
}

func TestSet2Bits76(t *testing.T) {
	out := Set2Bits76(0x00, 0x07)
	assert.Equal(t, uint8(0x60), out, "")

	out = Set2Bits76(0x30, 0x06)
	assert.Equal(t, uint8(0x50), out, "")

	out = Set2Bits76(0x37, 0x37)
	assert.Equal(t, uint8(0x77), out, "")

	out = Set2Bits76(0x37, 0xf4)
	assert.Equal(t, uint8(0x17), out, "")

	out = Set2Bits76(0x3f, 0x08)
	assert.Equal(t, uint8(0x1f), out, "")
}

func TestGet5Bits51(t *testing.T) {
	out := Get5Bits51(0x31)
	assert.Equal(t, uint8(0x11), out, "")

	out = Get5Bits51(0xE7)
	assert.Equal(t, uint8(0x07), out, "")

	out = Get5Bits51(0xaf)
	assert.Equal(t, uint8(0x0f), out, "")
}

func TestSet5Bits51(t *testing.T) {
	out := Set5Bits51(0x30, 0x07)
	assert.Equal(t, uint8(0x27), out, "")

	out = Set5Bits51(0x37, 0x06)
	assert.Equal(t, uint8(0x26), out, "")

	out = Set5Bits51(0x37, 0xf4)
	assert.Equal(t, uint8(0x34), out, "")

	out = Set5Bits51(0x9f, 0x08)
	assert.Equal(t, uint8(0x88), out, "")
}

func TestGet6Bits61(t *testing.T) {
	out := Get6Bits61(0x31)
	assert.Equal(t, uint8(0x31), out, "")

	out = Get6Bits61(0xE7)
	assert.Equal(t, uint8(0x27), out, "")

	out = Get6Bits61(0xaf)
	assert.Equal(t, uint8(0x2f), out, "")
}

func TestSet6Bits61(t *testing.T) {
	out := Set6Bits61(0x30, 0x07)
	assert.Equal(t, uint8(0x07), out, "")

	out = Set6Bits61(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = Set6Bits61(0x37, 0xff)
	assert.Equal(t, uint8(0x3f), out, "")

	out = Set6Bits61(0x9f, 0x01)
	assert.Equal(t, uint8(0x81), out, "")
}

func TestGetBit1(t *testing.T) {
	out := GetBit1(0x31)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit1(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit1(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestGetBit2(t *testing.T) {
	out := GetBit2(0x31)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit2(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit2(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestGetBit3(t *testing.T) {
	out := GetBit3(0x31)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit3(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit3(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestGetBit4(t *testing.T) {
	out := GetBit4(0x31)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit4(0xE7)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit4(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestGetBit5(t *testing.T) {
	out := GetBit5(0x31)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit5(0xE7)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit5(0xaf)
	assert.Equal(t, uint8(0x0), out, "")
}

func TestGetBit6(t *testing.T) {
	out := GetBit6(0x31)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit6(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit6(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestGetBit7(t *testing.T) {
	out := GetBit7(0x31)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit7(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit7(0xaf)
	assert.Equal(t, uint8(0x0), out, "")
}

func TestGetBit8(t *testing.T) {
	out := GetBit8(0x31)
	assert.Equal(t, uint8(0x0), out, "")

	out = GetBit8(0xE7)
	assert.Equal(t, uint8(0x1), out, "")

	out = GetBit8(0xaf)
	assert.Equal(t, uint8(0x1), out, "")
}

func TestSetBit1(t *testing.T) {
	out := SetBit1(0x30, 0x01)
	assert.Equal(t, uint8(0x31), out, "")

	out = SetBit1(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit1(0x37, 0x01)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit1(0x9f, 0x00)
	assert.Equal(t, uint8(0x9e), out, "")
}

func TestSetBit2(t *testing.T) {
	out := SetBit2(0x30, 0x07)
	assert.Equal(t, uint8(0x32), out, "")

	out = SetBit2(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit2(0x37, 0x01)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit2(0x9f, 0x00)
	assert.Equal(t, uint8(0x9D), out, "")
}

func TestSetBit3(t *testing.T) {
	out := SetBit3(0x30, 0x07)
	assert.Equal(t, uint8(0x34), out, "")

	out = SetBit3(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit3(0x37, 0x01)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit3(0x9f, 0x00)
	assert.Equal(t, uint8(0x9b), out, "")
}

func TestSetBit4(t *testing.T) {
	out := SetBit4(0x30, 0x07)
	assert.Equal(t, uint8(0x38), out, "")

	out = SetBit4(0x37, 0x37)
	assert.Equal(t, uint8(0x3f), out, "")

	out = SetBit4(0x37, 0x04)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit4(0x9f, 0x00)
	assert.Equal(t, uint8(0x97), out, "")
}

func TestSetBit5(t *testing.T) {
	out := SetBit5(0x30, 0x07)
	assert.Equal(t, uint8(0x30), out, "")

	out = SetBit5(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit5(0x37, 0x04)
	assert.Equal(t, uint8(0x27), out, "")

	out = SetBit5(0x9f, 0x00)
	assert.Equal(t, uint8(0x8f), out, "")
}

func TestSetBit6(t *testing.T) {
	out := SetBit6(0x30, 0x07)
	assert.Equal(t, uint8(0x30), out, "")

	out = SetBit6(0x37, 0x37)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit6(0x37, 0x04)
	assert.Equal(t, uint8(0x17), out, "")

	out = SetBit6(0x9f, 0x00)
	assert.Equal(t, uint8(0x9f), out, "")
}

func TestSetBit7(t *testing.T) {
	out := SetBit7(0x30, 0x07)
	assert.Equal(t, uint8(0x70), out, "")

	out = SetBit7(0x37, 0x37)
	assert.Equal(t, uint8(0x77), out, "")

	out = SetBit7(0x37, 0x04)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit7(0x9f, 0x00)
	assert.Equal(t, uint8(0x9f), out, "")
}

func TestSetBit8(t *testing.T) {
	out := SetBit8(0x30, 0x07)
	assert.Equal(t, uint8(0xB0), out, "")

	out = SetBit8(0x37, 0x37)
	assert.Equal(t, uint8(0xB7), out, "")

	out = SetBit8(0x37, 0x04)
	assert.Equal(t, uint8(0x37), out, "")

	out = SetBit8(0x9f, 0x00)
	assert.Equal(t, uint8(0x1f), out, "")
}

func TestGetMaskedValue8(t *testing.T) {
	var intA uint8 = 0x35
	var val uint8
	var err error

	t.Log("Testing GetMaskedValue8")

	if _, err = GetMaskedValue8(intA, BIT1, BIT3); err == nil {
		t.Error("start bit > end bit test failed")
	}

	if _, err = GetMaskedValue8(intA, BIT16, BIT3); err == nil {
		t.Error("start bit > 8 test failed")
	}

	if val, err = GetMaskedValue8(intA, BIT1, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(1), val, "")

	if val, err = GetMaskedValue8(intA, BIT3, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(5), val, "")

	if val, err = GetMaskedValue8(intA, BIT6, BIT5); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(3), val, "")
}

func TestSetMaskedValue8(t *testing.T) {
	var buf [1]uint8
	var err error

	b := buf[:]
	t.Log("Testing SetMaskedValue8")

	if err = SetMaskedValue8(b, uint8(0x35), BIT1, BIT3); err == nil {
		t.Error("start bit > end bit test failed")
	}

	if err = SetMaskedValue8(b, uint8(0x02), BIT1, BIT1); err == nil {
		t.Error("value > max value test failed")
	}

	if err = SetMaskedValue8(b, uint8(0x05), BIT3, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(0x05), b[0], "")

	b[0] = 0x07
	if err = SetMaskedValue8(b, uint8(0x02), BIT3, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(0x02), b[0], "")

	b[0] = 0x0F
	if err = SetMaskedValue8(b, uint8(0x02), BIT3, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(0x0A), b[0], "")

	b[0] = 0x02
	if err = SetMaskedValue8(b, uint8(0x03), BIT6, BIT5); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint8(0x32), b[0], "")

	if err = SetMaskedValue8(b, uint8(0x35), BIT10, BIT9); err == nil {
		t.Error("start bit > end bit test failed")
	}
}

func TestGetMaskedValue16(t *testing.T) {
	var b uint16
	var err error

	t.Log("Testing GetMaskedValue16")

	if b, err = GetMaskedValue16([]byte{0x76, 0x54}, BIT16, BIT4); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint16(0xECA), b, "")

	if b, err = GetMaskedValue16([]byte{0x76, 0x54}, BIT16, BIT5); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint16(0x765), b, "")

	if b, err = GetMaskedValue16([]byte{0x76, 0x54}, BIT16, BIT7); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint16(0x1d9), b, "")

	if b, err = GetMaskedValue16([]byte{0xff, 0xff}, BIT16, BIT7); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint16(0x3ff), b, "")

	if b, err = GetMaskedValue16([]byte{0xff, 0xff}, BIT10, BIT10); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, uint16(1), b, "")
}

func TestSetMaskedValue16(t *testing.T) {
	var b [2]byte
	var err error

	t.Log("Testing SetMaskedValue16")

	slb := b[:]
	if err = SetMaskedValue16(slb, 0x7654, BIT16, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, [2]byte{0x76, 0x54}, b, "")

	if err = SetMaskedValue16(slb, 0x0, BIT16, BIT1); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, [2]byte{0x00, 0x00}, b, "")

	if err = SetMaskedValue16(slb, 0x76, BIT16, BIT9); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, [2]byte{0x76, 0x00}, b, "")

	if err = SetMaskedValue16(slb, 0x355, BIT16, BIT7); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, [2]byte{0xd5, 0x40}, b, "")
}

func BenchmarkGetMaskedValue8(b *testing.B) {
	var intA uint8 = 0x35

	for i := 0; i < b.N; i++ {
		if a, err := GetMaskedValue8(intA, BIT6, BIT5); err != nil {
			intA += a
		}
		intA += 1
	}
}
