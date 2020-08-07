package nasType

// PDUSessionStatus 9.11.3.44
// PSI7 Row, sBit, len = [0, 0], 8 , 1
// PSI6 Row, sBit, len = [0, 0], 7 , 1
// PSI5 Row, sBit, len = [0, 0], 6 , 1
// PSI4 Row, sBit, len = [0, 0], 5 , 1
// PSI3 Row, sBit, len = [0, 0], 4 , 1
// PSI2 Row, sBit, len = [0, 0], 3 , 1
// PSI1 Row, sBit, len = [0, 0], 2 , 1
// PSI0 Row, sBit, len = [0, 0], 1 , 1
// PSI15 Row, sBit, len = [1, 1], 8 , 1
// PSI14 Row, sBit, len = [1, 1], 7 , 1
// PSI13 Row, sBit, len = [1, 1], 6 , 1
// PSI12 Row, sBit, len = [1, 1], 5 , 1
// PSI11 Row, sBit, len = [1, 1], 4 , 1
// PSI10 Row, sBit, len = [1, 1], 3 , 1
// PSI9 Row, sBit, len = [1, 1], 2 , 1
// PSI8 Row, sBit, len = [1, 1], 1 , 1
// Spare Row, sBit, len = [2, 2], 1 , INF
type PDUSessionStatus struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewPDUSessionStatus(iei uint8) (pDUSessionStatus *PDUSessionStatus) {
	pDUSessionStatus = &PDUSessionStatus{}
	pDUSessionStatus.SetIei(iei)
	return pDUSessionStatus
}

// PDUSessionStatus 9.11.3.44
// Iei Row, sBit, len = [], 8, 8
func (a *PDUSessionStatus) GetIei() (iei uint8) {
	return a.Iei
}

// PDUSessionStatus 9.11.3.44
// Iei Row, sBit, len = [], 8, 8
func (a *PDUSessionStatus) SetIei(iei uint8) {
	a.Iei = iei
}

// PDUSessionStatus 9.11.3.44
// Len Row, sBit, len = [], 8, 8
func (a *PDUSessionStatus) GetLen() (len uint8) {
	return a.Len
}

// PDUSessionStatus 9.11.3.44
// Len Row, sBit, len = [], 8, 8
func (a *PDUSessionStatus) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// PDUSessionStatus 9.11.3.44
// PSI7 Row, sBit, len = [0, 0], 8 , 1
func (a *PDUSessionStatus) GetPSI7() (pSI7 uint8) {
	return a.Buffer[0] & GetBitMask(8, 7) >> (7)
}

// PDUSessionStatus 9.11.3.44
// PSI7 Row, sBit, len = [0, 0], 8 , 1
func (a *PDUSessionStatus) SetPSI7(pSI7 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 127) + ((pSI7 & 1) << 7)
}

// PDUSessionStatus 9.11.3.44
// PSI6 Row, sBit, len = [0, 0], 7 , 1
func (a *PDUSessionStatus) GetPSI6() (pSI6 uint8) {
	return a.Buffer[0] & GetBitMask(7, 6) >> (6)
}

// PDUSessionStatus 9.11.3.44
// PSI6 Row, sBit, len = [0, 0], 7 , 1
func (a *PDUSessionStatus) SetPSI6(pSI6 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 191) + ((pSI6 & 1) << 6)
}

// PDUSessionStatus 9.11.3.44
// PSI5 Row, sBit, len = [0, 0], 6 , 1
func (a *PDUSessionStatus) GetPSI5() (pSI5 uint8) {
	return a.Buffer[0] & GetBitMask(6, 5) >> (5)
}

// PDUSessionStatus 9.11.3.44
// PSI5 Row, sBit, len = [0, 0], 6 , 1
func (a *PDUSessionStatus) SetPSI5(pSI5 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 223) + ((pSI5 & 1) << 5)
}

// PDUSessionStatus 9.11.3.44
// PSI4 Row, sBit, len = [0, 0], 5 , 1
func (a *PDUSessionStatus) GetPSI4() (pSI4 uint8) {
	return a.Buffer[0] & GetBitMask(5, 4) >> (4)
}

// PDUSessionStatus 9.11.3.44
// PSI4 Row, sBit, len = [0, 0], 5 , 1
func (a *PDUSessionStatus) SetPSI4(pSI4 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 239) + ((pSI4 & 1) << 4)
}

// PDUSessionStatus 9.11.3.44
// PSI3 Row, sBit, len = [0, 0], 4 , 1
func (a *PDUSessionStatus) GetPSI3() (pSI3 uint8) {
	return a.Buffer[0] & GetBitMask(4, 3) >> (3)
}

// PDUSessionStatus 9.11.3.44
// PSI3 Row, sBit, len = [0, 0], 4 , 1
func (a *PDUSessionStatus) SetPSI3(pSI3 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 247) + ((pSI3 & 1) << 3)
}

// PDUSessionStatus 9.11.3.44
// PSI2 Row, sBit, len = [0, 0], 3 , 1
func (a *PDUSessionStatus) GetPSI2() (pSI2 uint8) {
	return a.Buffer[0] & GetBitMask(3, 2) >> (2)
}

// PDUSessionStatus 9.11.3.44
// PSI2 Row, sBit, len = [0, 0], 3 , 1
func (a *PDUSessionStatus) SetPSI2(pSI2 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 251) + ((pSI2 & 1) << 2)
}

// PDUSessionStatus 9.11.3.44
// PSI1 Row, sBit, len = [0, 0], 2 , 1
func (a *PDUSessionStatus) GetPSI1() (pSI1 uint8) {
	return a.Buffer[0] & GetBitMask(2, 1) >> (1)
}

// PDUSessionStatus 9.11.3.44
// PSI1 Row, sBit, len = [0, 0], 2 , 1
func (a *PDUSessionStatus) SetPSI1(pSI1 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 253) + ((pSI1 & 1) << 1)
}

// PDUSessionStatus 9.11.3.44
// PSI0 Row, sBit, len = [0, 0], 1 , 1
func (a *PDUSessionStatus) GetPSI0() (pSI0 uint8) {
	return a.Buffer[0] & GetBitMask(1, 0)
}

// PDUSessionStatus 9.11.3.44
// PSI0 Row, sBit, len = [0, 0], 1 , 1
func (a *PDUSessionStatus) SetPSI0(pSI0 uint8) {
	a.Buffer[0] = (a.Buffer[0] & 254) + (pSI0 & 1)
}

// PDUSessionStatus 9.11.3.44
// PSI15 Row, sBit, len = [1, 1], 8 , 1
func (a *PDUSessionStatus) GetPSI15() (pSI15 uint8) {
	return a.Buffer[1] & GetBitMask(8, 7) >> (7)
}

// PDUSessionStatus 9.11.3.44
// PSI15 Row, sBit, len = [1, 1], 8 , 1
func (a *PDUSessionStatus) SetPSI15(pSI15 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 127) + ((pSI15 & 1) << 7)
}

// PDUSessionStatus 9.11.3.44
// PSI14 Row, sBit, len = [1, 1], 7 , 1
func (a *PDUSessionStatus) GetPSI14() (pSI14 uint8) {
	return a.Buffer[1] & GetBitMask(7, 6) >> (6)
}

// PDUSessionStatus 9.11.3.44
// PSI14 Row, sBit, len = [1, 1], 7 , 1
func (a *PDUSessionStatus) SetPSI14(pSI14 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 191) + ((pSI14 & 1) << 6)
}

// PDUSessionStatus 9.11.3.44
// PSI13 Row, sBit, len = [1, 1], 6 , 1
func (a *PDUSessionStatus) GetPSI13() (pSI13 uint8) {
	return a.Buffer[1] & GetBitMask(6, 5) >> (5)
}

// PDUSessionStatus 9.11.3.44
// PSI13 Row, sBit, len = [1, 1], 6 , 1
func (a *PDUSessionStatus) SetPSI13(pSI13 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 223) + ((pSI13 & 1) << 5)
}

// PDUSessionStatus 9.11.3.44
// PSI12 Row, sBit, len = [1, 1], 5 , 1
func (a *PDUSessionStatus) GetPSI12() (pSI12 uint8) {
	return a.Buffer[1] & GetBitMask(5, 4) >> (4)
}

// PDUSessionStatus 9.11.3.44
// PSI12 Row, sBit, len = [1, 1], 5 , 1
func (a *PDUSessionStatus) SetPSI12(pSI12 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 239) + ((pSI12 & 1) << 4)
}

// PDUSessionStatus 9.11.3.44
// PSI11 Row, sBit, len = [1, 1], 4 , 1
func (a *PDUSessionStatus) GetPSI11() (pSI11 uint8) {
	return a.Buffer[1] & GetBitMask(4, 3) >> (3)
}

// PDUSessionStatus 9.11.3.44
// PSI11 Row, sBit, len = [1, 1], 4 , 1
func (a *PDUSessionStatus) SetPSI11(pSI11 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 247) + ((pSI11 & 1) << 3)
}

// PDUSessionStatus 9.11.3.44
// PSI10 Row, sBit, len = [1, 1], 3 , 1
func (a *PDUSessionStatus) GetPSI10() (pSI10 uint8) {
	return a.Buffer[1] & GetBitMask(3, 2) >> (2)
}

// PDUSessionStatus 9.11.3.44
// PSI10 Row, sBit, len = [1, 1], 3 , 1
func (a *PDUSessionStatus) SetPSI10(pSI10 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 251) + ((pSI10 & 1) << 2)
}

// PDUSessionStatus 9.11.3.44
// PSI9 Row, sBit, len = [1, 1], 2 , 1
func (a *PDUSessionStatus) GetPSI9() (pSI9 uint8) {
	return a.Buffer[1] & GetBitMask(2, 1) >> (1)
}

// PDUSessionStatus 9.11.3.44
// PSI9 Row, sBit, len = [1, 1], 2 , 1
func (a *PDUSessionStatus) SetPSI9(pSI9 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 253) + ((pSI9 & 1) << 1)
}

// PDUSessionStatus 9.11.3.44
// PSI8 Row, sBit, len = [1, 1], 1 , 1
func (a *PDUSessionStatus) GetPSI8() (pSI8 uint8) {
	return a.Buffer[1] & GetBitMask(1, 0)
}

// PDUSessionStatus 9.11.3.44
// PSI8 Row, sBit, len = [1, 1], 1 , 1
func (a *PDUSessionStatus) SetPSI8(pSI8 uint8) {
	a.Buffer[1] = (a.Buffer[1] & 254) + (pSI8 & 1)
}

// PDUSessionStatus 9.11.3.44
// Spare Row, sBit, len = [2, 2], 1 , INF
func (a *PDUSessionStatus) GetSpare() (spare []uint8) {
	spare = make([]uint8, len(a.Buffer)-2)
	copy(spare, a.Buffer[2:])
	return spare
}

// PDUSessionStatus 9.11.3.44
// Spare Row, sBit, len = [2, 2], 1 , INF
func (a *PDUSessionStatus) SetSpare(spare []uint8) {
	copy(a.Buffer[2:], spare)
}
