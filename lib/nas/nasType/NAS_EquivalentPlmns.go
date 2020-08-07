package nasType

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN1 Row, sBit, len = [0, 0], 8 , 4
// MCCDigit1PLMN1 Row, sBit, len = [0, 0], 4 , 4
// MNCDigit3PLMN1 Row, sBit, len = [1, 1], 8 , 4
// MCCDigit3PLMN1 Row, sBit, len = [1, 1], 4 , 4
// MNCDigit2PLMN1 Row, sBit, len = [2, 2], 8 , 4
// MNCDigit1PLMN1 Row, sBit, len = [2, 2], 4 , 4
// MCCDigit2PLMN2 Row, sBit, len = [3, 3], 8 , 4
// MCCDigit1PLMN2 Row, sBit, len = [3, 3], 4 , 4
// MNCDigit3PLMN2 Row, sBit, len = [4, 4], 8 , 4
// MCCDigit3PLMN2 Row, sBit, len = [4, 4], 4 , 4
// MNCDigit2PLMN2 Row, sBit, len = [5, 5], 8 , 4
// MNCDigit1PLMN2 Row, sBit, len = [5, 5], 4 , 4
// MCCDigit2PLMN3 Row, sBit, len = [6, 6], 8 , 4
// MCCDigit1PLMN3 Row, sBit, len = [6, 6], 4 , 4
// MNCDigit3PLMN3 Row, sBit, len = [7, 7], 8 , 4
// MCCDigit3PLMN3 Row, sBit, len = [7, 7], 4 , 4
// MNCDigit2PLMN3 Row, sBit, len = [8, 8], 8 , 4
// MNCDigit1PLMN3 Row, sBit, len = [8, 8], 4 , 4
// MCCDigit2PLMN4 Row, sBit, len = [9, 9], 8 , 4
// MCCDigit1PLMN4 Row, sBit, len = [9, 9], 4 , 4
// MNCDigit3PLMN4 Row, sBit, len = [10, 10], 8 , 4
// MCCDigit3PLMN4 Row, sBit, len = [10, 10], 4 , 4
// MNCDigit2PLMN4 Row, sBit, len = [11, 11], 8 , 4
// MNCDigit1PLMN4 Row, sBit, len = [11, 11], 4 , 4
// MCCDigit2PLMN5 Row, sBit, len = [12, 12], 8 , 4
// MCCDigit1PLMN5 Row, sBit, len = [12, 12], 4 , 4
// MNCDigit3PLMN5 Row, sBit, len = [13, 13], 8 , 4
// MCCDigit3PLMN5 Row, sBit, len = [13, 13], 4 , 4
// MNCDigit2PLMN5 Row, sBit, len = [14, 14], 8 , 4
// MNCDigit1PLMN5 Row, sBit, len = [14, 14], 4 , 4
// MCCDigit2PLMN6 Row, sBit, len = [15, 15], 8 , 4
// MCCDigit1PLMN6 Row, sBit, len = [15, 15], 4 , 4
// MNCDigit3PLMN6 Row, sBit, len = [16, 16], 8 , 4
// MCCDigit3PLMN6 Row, sBit, len = [16, 16], 4 , 4
// MNCDigit2PLMN6 Row, sBit, len = [17, 17], 8 , 4
// MNCDigit1PLMN6 Row, sBit, len = [17, 17], 4 , 4
// MCCDigit2PLMN7 Row, sBit, len = [18, 18], 8 , 4
// MCCDigit1PLMN7 Row, sBit, len = [18, 18], 4 , 4
// MNCDigit3PLMN7 Row, sBit, len = [19, 19], 8 , 4
// MCCDigit3PLMN7 Row, sBit, len = [19, 19], 4 , 4
// MNCDigit2PLMN7 Row, sBit, len = [20, 20], 8 , 4
// MNCDigit1PLMN7 Row, sBit, len = [20, 20], 4 , 4
// MCCDigit2PLMN8 Row, sBit, len = [21, 21], 8 , 4
// MCCDigit1PLMN8 Row, sBit, len = [21, 21], 4 , 4
// MNCDigit3PLMN8 Row, sBit, len = [22, 22], 8 , 4
// MCCDigit3PLMN8 Row, sBit, len = [22, 22], 4 , 4
// MNCDigit2PLMN8 Row, sBit, len = [23, 23], 8 , 4
// MNCDigit1PLMN8 Row, sBit, len = [23, 23], 4 , 4
// MCCDigit2PLMN9 Row, sBit, len = [24, 24], 8 , 4
// MCCDigit1PLMN9 Row, sBit, len = [24, 24], 4 , 4
// MNCDigit3PLMN9 Row, sBit, len = [25, 25], 8 , 4
// MCCDigit3PLMN9 Row, sBit, len = [25, 25], 4 , 4
// MNCDigit2PLMN9 Row, sBit, len = [26, 26], 8 , 4
// MNCDigit1PLMN9 Row, sBit, len = [26, 26], 4 , 4
// MCCDigit2PLMN10 Row, sBit, len = [27, 27], 8 , 4
// MCCDigit1PLMN10 Row, sBit, len = [27, 27], 4 , 4
// MNCDigit3PLMN10 Row, sBit, len = [28, 28], 8 , 4
// MCCDigit3PLMN10 Row, sBit, len = [28, 28], 4 , 4
// MNCDigit2PLMN10 Row, sBit, len = [29, 29], 8 , 4
// MNCDigit1PLMN10 Row, sBit, len = [29, 29], 4 , 4
// MCCDigit2PLMN11 Row, sBit, len = [30, 30], 8 , 4
// MCCDigit1PLMN11 Row, sBit, len = [30, 30], 4 , 4
// MNCDigit3PLMN11 Row, sBit, len = [31, 31], 8 , 4
// MCCDigit3PLMN11 Row, sBit, len = [31, 31], 4 , 4
// MNCDigit2PLMN11 Row, sBit, len = [32, 32], 8 , 4
// MNCDigit1PLMN11 Row, sBit, len = [32, 32], 4 , 4
// MCCDigit2PLMN12 Row, sBit, len = [33, 33], 8 , 4
// MCCDigit1PLMN12 Row, sBit, len = [33, 33], 4 , 4
// MNCDigit3PLMN12 Row, sBit, len = [34, 34], 8 , 4
// MCCDigit3PLMN12 Row, sBit, len = [34, 34], 4 , 4
// MNCDigit2PLMN12 Row, sBit, len = [35, 35], 8 , 4
// MNCDigit1PLMN12 Row, sBit, len = [35, 35], 4 , 4
// MCCDigit2PLMN13 Row, sBit, len = [36, 36], 8 , 4
// MCCDigit1PLMN13 Row, sBit, len = [36, 36], 4 , 4
// MNCDigit3PLMN13 Row, sBit, len = [37, 37], 8 , 4
// MCCDigit3PLMN13 Row, sBit, len = [37, 37], 4 , 4
// MNCDigit2PLMN13 Row, sBit, len = [38, 38], 8 , 4
// MNCDigit1PLMN13 Row, sBit, len = [38, 38], 4 , 4
// MCCDigit2PLMN14 Row, sBit, len = [39, 39], 8 , 4
// MCCDigit1PLMN14 Row, sBit, len = [39, 39], 4 , 4
// MNCDigit3PLMN14 Row, sBit, len = [40, 40], 8 , 4
// MCCDigit3PLMN14 Row, sBit, len = [40, 40], 4 , 4
// MNCDigit2PLMN14 Row, sBit, len = [41, 41], 8 , 4
// MNCDigit1PLMN14 Row, sBit, len = [41, 41], 4 , 4
// MCCDigit2PLMN15 Row, sBit, len = [42, 42], 8 , 4
// MCCDigit1PLMN15 Row, sBit, len = [42, 42], 4 , 4
// MNCDigit3PLMN15 Row, sBit, len = [43, 43], 8 , 4
// MCCDigit3PLMN15 Row, sBit, len = [43, 43], 4 , 4
// MNCDigit2PLMN15 Row, sBit, len = [44, 44], 8 , 4
// MNCDigit1PLMN15 Row, sBit, len = [44, 44], 4 , 4
type EquivalentPlmns struct {
	Iei   uint8
	Len   uint8
	Octet [45]uint8
}

func NewEquivalentPlmns(iei uint8) (equivalentPlmns *EquivalentPlmns) {
	equivalentPlmns = &EquivalentPlmns{}
	equivalentPlmns.SetIei(iei)
	return equivalentPlmns
}

// EquivalentPlmns 9.11.3.45
// Iei Row, sBit, len = [], 8, 8
func (a *EquivalentPlmns) GetIei() (iei uint8) {
	return a.Iei
}

// EquivalentPlmns 9.11.3.45
// Iei Row, sBit, len = [], 8, 8
func (a *EquivalentPlmns) SetIei(iei uint8) {
	a.Iei = iei
}

// EquivalentPlmns 9.11.3.45
// Len Row, sBit, len = [], 8, 8
func (a *EquivalentPlmns) GetLen() (len uint8) {
	return a.Len
}

// EquivalentPlmns 9.11.3.45
// Len Row, sBit, len = [], 8, 8
func (a *EquivalentPlmns) SetLen(len uint8) {
	a.Len = len
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN1 Row, sBit, len = [0, 0], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN1() (mCCDigit2PLMN1 uint8) {
	return a.Octet[0] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN1 Row, sBit, len = [0, 0], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN1(mCCDigit2PLMN1 uint8) {
	a.Octet[0] = (a.Octet[0] & 15) + ((mCCDigit2PLMN1 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN1 Row, sBit, len = [0, 0], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN1() (mCCDigit1PLMN1 uint8) {
	return a.Octet[0] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN1 Row, sBit, len = [0, 0], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN1(mCCDigit1PLMN1 uint8) {
	a.Octet[0] = (a.Octet[0] & 240) + (mCCDigit1PLMN1 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN1 Row, sBit, len = [1, 1], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN1() (mNCDigit3PLMN1 uint8) {
	return a.Octet[1] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN1 Row, sBit, len = [1, 1], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN1(mNCDigit3PLMN1 uint8) {
	a.Octet[1] = (a.Octet[1] & 15) + ((mNCDigit3PLMN1 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN1 Row, sBit, len = [1, 1], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN1() (mCCDigit3PLMN1 uint8) {
	return a.Octet[1] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN1 Row, sBit, len = [1, 1], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN1(mCCDigit3PLMN1 uint8) {
	a.Octet[1] = (a.Octet[1] & 240) + (mCCDigit3PLMN1 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN1 Row, sBit, len = [2, 2], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN1() (mNCDigit2PLMN1 uint8) {
	return a.Octet[2] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN1 Row, sBit, len = [2, 2], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN1(mNCDigit2PLMN1 uint8) {
	a.Octet[2] = (a.Octet[2] & 15) + ((mNCDigit2PLMN1 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN1 Row, sBit, len = [2, 2], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN1() (mNCDigit1PLMN1 uint8) {
	return a.Octet[2] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN1 Row, sBit, len = [2, 2], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN1(mNCDigit1PLMN1 uint8) {
	a.Octet[2] = (a.Octet[2] & 240) + (mNCDigit1PLMN1 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN2 Row, sBit, len = [3, 3], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN2() (mCCDigit2PLMN2 uint8) {
	return a.Octet[3] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN2 Row, sBit, len = [3, 3], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN2(mCCDigit2PLMN2 uint8) {
	a.Octet[3] = (a.Octet[3] & 15) + ((mCCDigit2PLMN2 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN2 Row, sBit, len = [3, 3], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN2() (mCCDigit1PLMN2 uint8) {
	return a.Octet[3] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN2 Row, sBit, len = [3, 3], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN2(mCCDigit1PLMN2 uint8) {
	a.Octet[3] = (a.Octet[3] & 240) + (mCCDigit1PLMN2 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN2 Row, sBit, len = [4, 4], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN2() (mNCDigit3PLMN2 uint8) {
	return a.Octet[4] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN2 Row, sBit, len = [4, 4], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN2(mNCDigit3PLMN2 uint8) {
	a.Octet[4] = (a.Octet[4] & 15) + ((mNCDigit3PLMN2 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN2 Row, sBit, len = [4, 4], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN2() (mCCDigit3PLMN2 uint8) {
	return a.Octet[4] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN2 Row, sBit, len = [4, 4], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN2(mCCDigit3PLMN2 uint8) {
	a.Octet[4] = (a.Octet[4] & 240) + (mCCDigit3PLMN2 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN2 Row, sBit, len = [5, 5], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN2() (mNCDigit2PLMN2 uint8) {
	return a.Octet[5] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN2 Row, sBit, len = [5, 5], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN2(mNCDigit2PLMN2 uint8) {
	a.Octet[5] = (a.Octet[5] & 15) + ((mNCDigit2PLMN2 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN2 Row, sBit, len = [5, 5], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN2() (mNCDigit1PLMN2 uint8) {
	return a.Octet[5] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN2 Row, sBit, len = [5, 5], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN2(mNCDigit1PLMN2 uint8) {
	a.Octet[5] = (a.Octet[5] & 240) + (mNCDigit1PLMN2 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN3 Row, sBit, len = [6, 6], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN3() (mCCDigit2PLMN3 uint8) {
	return a.Octet[6] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN3 Row, sBit, len = [6, 6], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN3(mCCDigit2PLMN3 uint8) {
	a.Octet[6] = (a.Octet[6] & 15) + ((mCCDigit2PLMN3 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN3 Row, sBit, len = [6, 6], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN3() (mCCDigit1PLMN3 uint8) {
	return a.Octet[6] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN3 Row, sBit, len = [6, 6], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN3(mCCDigit1PLMN3 uint8) {
	a.Octet[6] = (a.Octet[6] & 240) + (mCCDigit1PLMN3 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN3 Row, sBit, len = [7, 7], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN3() (mNCDigit3PLMN3 uint8) {
	return a.Octet[7] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN3 Row, sBit, len = [7, 7], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN3(mNCDigit3PLMN3 uint8) {
	a.Octet[7] = (a.Octet[7] & 15) + ((mNCDigit3PLMN3 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN3 Row, sBit, len = [7, 7], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN3() (mCCDigit3PLMN3 uint8) {
	return a.Octet[7] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN3 Row, sBit, len = [7, 7], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN3(mCCDigit3PLMN3 uint8) {
	a.Octet[7] = (a.Octet[7] & 240) + (mCCDigit3PLMN3 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN3 Row, sBit, len = [8, 8], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN3() (mNCDigit2PLMN3 uint8) {
	return a.Octet[8] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN3 Row, sBit, len = [8, 8], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN3(mNCDigit2PLMN3 uint8) {
	a.Octet[8] = (a.Octet[8] & 15) + ((mNCDigit2PLMN3 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN3 Row, sBit, len = [8, 8], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN3() (mNCDigit1PLMN3 uint8) {
	return a.Octet[8] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN3 Row, sBit, len = [8, 8], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN3(mNCDigit1PLMN3 uint8) {
	a.Octet[8] = (a.Octet[8] & 240) + (mNCDigit1PLMN3 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN4 Row, sBit, len = [9, 9], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN4() (mCCDigit2PLMN4 uint8) {
	return a.Octet[9] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN4 Row, sBit, len = [9, 9], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN4(mCCDigit2PLMN4 uint8) {
	a.Octet[9] = (a.Octet[9] & 15) + ((mCCDigit2PLMN4 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN4 Row, sBit, len = [9, 9], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN4() (mCCDigit1PLMN4 uint8) {
	return a.Octet[9] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN4 Row, sBit, len = [9, 9], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN4(mCCDigit1PLMN4 uint8) {
	a.Octet[9] = (a.Octet[9] & 240) + (mCCDigit1PLMN4 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN4 Row, sBit, len = [10, 10], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN4() (mNCDigit3PLMN4 uint8) {
	return a.Octet[10] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN4 Row, sBit, len = [10, 10], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN4(mNCDigit3PLMN4 uint8) {
	a.Octet[10] = (a.Octet[10] & 15) + ((mNCDigit3PLMN4 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN4 Row, sBit, len = [10, 10], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN4() (mCCDigit3PLMN4 uint8) {
	return a.Octet[10] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN4 Row, sBit, len = [10, 10], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN4(mCCDigit3PLMN4 uint8) {
	a.Octet[10] = (a.Octet[10] & 240) + (mCCDigit3PLMN4 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN4 Row, sBit, len = [11, 11], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN4() (mNCDigit2PLMN4 uint8) {
	return a.Octet[11] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN4 Row, sBit, len = [11, 11], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN4(mNCDigit2PLMN4 uint8) {
	a.Octet[11] = (a.Octet[11] & 15) + ((mNCDigit2PLMN4 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN4 Row, sBit, len = [11, 11], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN4() (mNCDigit1PLMN4 uint8) {
	return a.Octet[11] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN4 Row, sBit, len = [11, 11], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN4(mNCDigit1PLMN4 uint8) {
	a.Octet[11] = (a.Octet[11] & 240) + (mNCDigit1PLMN4 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN5 Row, sBit, len = [12, 12], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN5() (mCCDigit2PLMN5 uint8) {
	return a.Octet[12] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN5 Row, sBit, len = [12, 12], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN5(mCCDigit2PLMN5 uint8) {
	a.Octet[12] = (a.Octet[12] & 15) + ((mCCDigit2PLMN5 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN5 Row, sBit, len = [12, 12], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN5() (mCCDigit1PLMN5 uint8) {
	return a.Octet[12] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN5 Row, sBit, len = [12, 12], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN5(mCCDigit1PLMN5 uint8) {
	a.Octet[12] = (a.Octet[12] & 240) + (mCCDigit1PLMN5 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN5 Row, sBit, len = [13, 13], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN5() (mNCDigit3PLMN5 uint8) {
	return a.Octet[13] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN5 Row, sBit, len = [13, 13], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN5(mNCDigit3PLMN5 uint8) {
	a.Octet[13] = (a.Octet[13] & 15) + ((mNCDigit3PLMN5 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN5 Row, sBit, len = [13, 13], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN5() (mCCDigit3PLMN5 uint8) {
	return a.Octet[13] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN5 Row, sBit, len = [13, 13], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN5(mCCDigit3PLMN5 uint8) {
	a.Octet[13] = (a.Octet[13] & 240) + (mCCDigit3PLMN5 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN5 Row, sBit, len = [14, 14], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN5() (mNCDigit2PLMN5 uint8) {
	return a.Octet[14] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN5 Row, sBit, len = [14, 14], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN5(mNCDigit2PLMN5 uint8) {
	a.Octet[14] = (a.Octet[14] & 15) + ((mNCDigit2PLMN5 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN5 Row, sBit, len = [14, 14], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN5() (mNCDigit1PLMN5 uint8) {
	return a.Octet[14] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN5 Row, sBit, len = [14, 14], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN5(mNCDigit1PLMN5 uint8) {
	a.Octet[14] = (a.Octet[14] & 240) + (mNCDigit1PLMN5 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN6 Row, sBit, len = [15, 15], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN6() (mCCDigit2PLMN6 uint8) {
	return a.Octet[15] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN6 Row, sBit, len = [15, 15], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN6(mCCDigit2PLMN6 uint8) {
	a.Octet[15] = (a.Octet[15] & 15) + ((mCCDigit2PLMN6 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN6 Row, sBit, len = [15, 15], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN6() (mCCDigit1PLMN6 uint8) {
	return a.Octet[15] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN6 Row, sBit, len = [15, 15], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN6(mCCDigit1PLMN6 uint8) {
	a.Octet[15] = (a.Octet[15] & 240) + (mCCDigit1PLMN6 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN6 Row, sBit, len = [16, 16], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN6() (mNCDigit3PLMN6 uint8) {
	return a.Octet[16] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN6 Row, sBit, len = [16, 16], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN6(mNCDigit3PLMN6 uint8) {
	a.Octet[16] = (a.Octet[16] & 15) + ((mNCDigit3PLMN6 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN6 Row, sBit, len = [16, 16], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN6() (mCCDigit3PLMN6 uint8) {
	return a.Octet[16] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN6 Row, sBit, len = [16, 16], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN6(mCCDigit3PLMN6 uint8) {
	a.Octet[16] = (a.Octet[16] & 240) + (mCCDigit3PLMN6 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN6 Row, sBit, len = [17, 17], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN6() (mNCDigit2PLMN6 uint8) {
	return a.Octet[17] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN6 Row, sBit, len = [17, 17], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN6(mNCDigit2PLMN6 uint8) {
	a.Octet[17] = (a.Octet[17] & 15) + ((mNCDigit2PLMN6 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN6 Row, sBit, len = [17, 17], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN6() (mNCDigit1PLMN6 uint8) {
	return a.Octet[17] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN6 Row, sBit, len = [17, 17], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN6(mNCDigit1PLMN6 uint8) {
	a.Octet[17] = (a.Octet[17] & 240) + (mNCDigit1PLMN6 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN7 Row, sBit, len = [18, 18], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN7() (mCCDigit2PLMN7 uint8) {
	return a.Octet[18] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN7 Row, sBit, len = [18, 18], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN7(mCCDigit2PLMN7 uint8) {
	a.Octet[18] = (a.Octet[18] & 15) + ((mCCDigit2PLMN7 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN7 Row, sBit, len = [18, 18], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN7() (mCCDigit1PLMN7 uint8) {
	return a.Octet[18] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN7 Row, sBit, len = [18, 18], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN7(mCCDigit1PLMN7 uint8) {
	a.Octet[18] = (a.Octet[18] & 240) + (mCCDigit1PLMN7 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN7 Row, sBit, len = [19, 19], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN7() (mNCDigit3PLMN7 uint8) {
	return a.Octet[19] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN7 Row, sBit, len = [19, 19], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN7(mNCDigit3PLMN7 uint8) {
	a.Octet[19] = (a.Octet[19] & 15) + ((mNCDigit3PLMN7 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN7 Row, sBit, len = [19, 19], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN7() (mCCDigit3PLMN7 uint8) {
	return a.Octet[19] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN7 Row, sBit, len = [19, 19], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN7(mCCDigit3PLMN7 uint8) {
	a.Octet[19] = (a.Octet[19] & 240) + (mCCDigit3PLMN7 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN7 Row, sBit, len = [20, 20], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN7() (mNCDigit2PLMN7 uint8) {
	return a.Octet[20] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN7 Row, sBit, len = [20, 20], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN7(mNCDigit2PLMN7 uint8) {
	a.Octet[20] = (a.Octet[20] & 15) + ((mNCDigit2PLMN7 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN7 Row, sBit, len = [20, 20], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN7() (mNCDigit1PLMN7 uint8) {
	return a.Octet[20] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN7 Row, sBit, len = [20, 20], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN7(mNCDigit1PLMN7 uint8) {
	a.Octet[20] = (a.Octet[20] & 240) + (mNCDigit1PLMN7 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN8 Row, sBit, len = [21, 21], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN8() (mCCDigit2PLMN8 uint8) {
	return a.Octet[21] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN8 Row, sBit, len = [21, 21], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN8(mCCDigit2PLMN8 uint8) {
	a.Octet[21] = (a.Octet[21] & 15) + ((mCCDigit2PLMN8 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN8 Row, sBit, len = [21, 21], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN8() (mCCDigit1PLMN8 uint8) {
	return a.Octet[21] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN8 Row, sBit, len = [21, 21], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN8(mCCDigit1PLMN8 uint8) {
	a.Octet[21] = (a.Octet[21] & 240) + (mCCDigit1PLMN8 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN8 Row, sBit, len = [22, 22], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN8() (mNCDigit3PLMN8 uint8) {
	return a.Octet[22] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN8 Row, sBit, len = [22, 22], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN8(mNCDigit3PLMN8 uint8) {
	a.Octet[22] = (a.Octet[22] & 15) + ((mNCDigit3PLMN8 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN8 Row, sBit, len = [22, 22], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN8() (mCCDigit3PLMN8 uint8) {
	return a.Octet[22] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN8 Row, sBit, len = [22, 22], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN8(mCCDigit3PLMN8 uint8) {
	a.Octet[22] = (a.Octet[22] & 240) + (mCCDigit3PLMN8 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN8 Row, sBit, len = [23, 23], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN8() (mNCDigit2PLMN8 uint8) {
	return a.Octet[23] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN8 Row, sBit, len = [23, 23], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN8(mNCDigit2PLMN8 uint8) {
	a.Octet[23] = (a.Octet[23] & 15) + ((mNCDigit2PLMN8 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN8 Row, sBit, len = [23, 23], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN8() (mNCDigit1PLMN8 uint8) {
	return a.Octet[23] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN8 Row, sBit, len = [23, 23], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN8(mNCDigit1PLMN8 uint8) {
	a.Octet[23] = (a.Octet[23] & 240) + (mNCDigit1PLMN8 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN9 Row, sBit, len = [24, 24], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN9() (mCCDigit2PLMN9 uint8) {
	return a.Octet[24] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN9 Row, sBit, len = [24, 24], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN9(mCCDigit2PLMN9 uint8) {
	a.Octet[24] = (a.Octet[24] & 15) + ((mCCDigit2PLMN9 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN9 Row, sBit, len = [24, 24], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN9() (mCCDigit1PLMN9 uint8) {
	return a.Octet[24] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN9 Row, sBit, len = [24, 24], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN9(mCCDigit1PLMN9 uint8) {
	a.Octet[24] = (a.Octet[24] & 240) + (mCCDigit1PLMN9 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN9 Row, sBit, len = [25, 25], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN9() (mNCDigit3PLMN9 uint8) {
	return a.Octet[25] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN9 Row, sBit, len = [25, 25], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN9(mNCDigit3PLMN9 uint8) {
	a.Octet[25] = (a.Octet[25] & 15) + ((mNCDigit3PLMN9 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN9 Row, sBit, len = [25, 25], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN9() (mCCDigit3PLMN9 uint8) {
	return a.Octet[25] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN9 Row, sBit, len = [25, 25], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN9(mCCDigit3PLMN9 uint8) {
	a.Octet[25] = (a.Octet[25] & 240) + (mCCDigit3PLMN9 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN9 Row, sBit, len = [26, 26], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN9() (mNCDigit2PLMN9 uint8) {
	return a.Octet[26] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN9 Row, sBit, len = [26, 26], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN9(mNCDigit2PLMN9 uint8) {
	a.Octet[26] = (a.Octet[26] & 15) + ((mNCDigit2PLMN9 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN9 Row, sBit, len = [26, 26], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN9() (mNCDigit1PLMN9 uint8) {
	return a.Octet[26] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN9 Row, sBit, len = [26, 26], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN9(mNCDigit1PLMN9 uint8) {
	a.Octet[26] = (a.Octet[26] & 240) + (mNCDigit1PLMN9 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN10 Row, sBit, len = [27, 27], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN10() (mCCDigit2PLMN10 uint8) {
	return a.Octet[27] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN10 Row, sBit, len = [27, 27], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN10(mCCDigit2PLMN10 uint8) {
	a.Octet[27] = (a.Octet[27] & 15) + ((mCCDigit2PLMN10 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN10 Row, sBit, len = [27, 27], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN10() (mCCDigit1PLMN10 uint8) {
	return a.Octet[27] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN10 Row, sBit, len = [27, 27], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN10(mCCDigit1PLMN10 uint8) {
	a.Octet[27] = (a.Octet[27] & 240) + (mCCDigit1PLMN10 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN10 Row, sBit, len = [28, 28], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN10() (mNCDigit3PLMN10 uint8) {
	return a.Octet[28] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN10 Row, sBit, len = [28, 28], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN10(mNCDigit3PLMN10 uint8) {
	a.Octet[28] = (a.Octet[28] & 15) + ((mNCDigit3PLMN10 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN10 Row, sBit, len = [28, 28], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN10() (mCCDigit3PLMN10 uint8) {
	return a.Octet[28] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN10 Row, sBit, len = [28, 28], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN10(mCCDigit3PLMN10 uint8) {
	a.Octet[28] = (a.Octet[28] & 240) + (mCCDigit3PLMN10 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN10 Row, sBit, len = [29, 29], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN10() (mNCDigit2PLMN10 uint8) {
	return a.Octet[29] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN10 Row, sBit, len = [29, 29], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN10(mNCDigit2PLMN10 uint8) {
	a.Octet[29] = (a.Octet[29] & 15) + ((mNCDigit2PLMN10 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN10 Row, sBit, len = [29, 29], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN10() (mNCDigit1PLMN10 uint8) {
	return a.Octet[29] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN10 Row, sBit, len = [29, 29], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN10(mNCDigit1PLMN10 uint8) {
	a.Octet[29] = (a.Octet[29] & 240) + (mNCDigit1PLMN10 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN11 Row, sBit, len = [30, 30], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN11() (mCCDigit2PLMN11 uint8) {
	return a.Octet[30] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN11 Row, sBit, len = [30, 30], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN11(mCCDigit2PLMN11 uint8) {
	a.Octet[30] = (a.Octet[30] & 15) + ((mCCDigit2PLMN11 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN11 Row, sBit, len = [30, 30], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN11() (mCCDigit1PLMN11 uint8) {
	return a.Octet[30] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN11 Row, sBit, len = [30, 30], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN11(mCCDigit1PLMN11 uint8) {
	a.Octet[30] = (a.Octet[30] & 240) + (mCCDigit1PLMN11 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN11 Row, sBit, len = [31, 31], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN11() (mNCDigit3PLMN11 uint8) {
	return a.Octet[31] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN11 Row, sBit, len = [31, 31], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN11(mNCDigit3PLMN11 uint8) {
	a.Octet[31] = (a.Octet[31] & 15) + ((mNCDigit3PLMN11 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN11 Row, sBit, len = [31, 31], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN11() (mCCDigit3PLMN11 uint8) {
	return a.Octet[31] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN11 Row, sBit, len = [31, 31], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN11(mCCDigit3PLMN11 uint8) {
	a.Octet[31] = (a.Octet[31] & 240) + (mCCDigit3PLMN11 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN11 Row, sBit, len = [32, 32], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN11() (mNCDigit2PLMN11 uint8) {
	return a.Octet[32] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN11 Row, sBit, len = [32, 32], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN11(mNCDigit2PLMN11 uint8) {
	a.Octet[32] = (a.Octet[32] & 15) + ((mNCDigit2PLMN11 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN11 Row, sBit, len = [32, 32], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN11() (mNCDigit1PLMN11 uint8) {
	return a.Octet[32] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN11 Row, sBit, len = [32, 32], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN11(mNCDigit1PLMN11 uint8) {
	a.Octet[32] = (a.Octet[32] & 240) + (mNCDigit1PLMN11 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN12 Row, sBit, len = [33, 33], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN12() (mCCDigit2PLMN12 uint8) {
	return a.Octet[33] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN12 Row, sBit, len = [33, 33], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN12(mCCDigit2PLMN12 uint8) {
	a.Octet[33] = (a.Octet[33] & 15) + ((mCCDigit2PLMN12 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN12 Row, sBit, len = [33, 33], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN12() (mCCDigit1PLMN12 uint8) {
	return a.Octet[33] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN12 Row, sBit, len = [33, 33], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN12(mCCDigit1PLMN12 uint8) {
	a.Octet[33] = (a.Octet[33] & 240) + (mCCDigit1PLMN12 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN12 Row, sBit, len = [34, 34], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN12() (mNCDigit3PLMN12 uint8) {
	return a.Octet[34] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN12 Row, sBit, len = [34, 34], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN12(mNCDigit3PLMN12 uint8) {
	a.Octet[34] = (a.Octet[34] & 15) + ((mNCDigit3PLMN12 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN12 Row, sBit, len = [34, 34], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN12() (mCCDigit3PLMN12 uint8) {
	return a.Octet[34] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN12 Row, sBit, len = [34, 34], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN12(mCCDigit3PLMN12 uint8) {
	a.Octet[34] = (a.Octet[34] & 240) + (mCCDigit3PLMN12 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN12 Row, sBit, len = [35, 35], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN12() (mNCDigit2PLMN12 uint8) {
	return a.Octet[35] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN12 Row, sBit, len = [35, 35], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN12(mNCDigit2PLMN12 uint8) {
	a.Octet[35] = (a.Octet[35] & 15) + ((mNCDigit2PLMN12 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN12 Row, sBit, len = [35, 35], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN12() (mNCDigit1PLMN12 uint8) {
	return a.Octet[35] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN12 Row, sBit, len = [35, 35], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN12(mNCDigit1PLMN12 uint8) {
	a.Octet[35] = (a.Octet[35] & 240) + (mNCDigit1PLMN12 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN13 Row, sBit, len = [36, 36], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN13() (mCCDigit2PLMN13 uint8) {
	return a.Octet[36] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN13 Row, sBit, len = [36, 36], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN13(mCCDigit2PLMN13 uint8) {
	a.Octet[36] = (a.Octet[36] & 15) + ((mCCDigit2PLMN13 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN13 Row, sBit, len = [36, 36], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN13() (mCCDigit1PLMN13 uint8) {
	return a.Octet[36] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN13 Row, sBit, len = [36, 36], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN13(mCCDigit1PLMN13 uint8) {
	a.Octet[36] = (a.Octet[36] & 240) + (mCCDigit1PLMN13 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN13 Row, sBit, len = [37, 37], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN13() (mNCDigit3PLMN13 uint8) {
	return a.Octet[37] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN13 Row, sBit, len = [37, 37], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN13(mNCDigit3PLMN13 uint8) {
	a.Octet[37] = (a.Octet[37] & 15) + ((mNCDigit3PLMN13 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN13 Row, sBit, len = [37, 37], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN13() (mCCDigit3PLMN13 uint8) {
	return a.Octet[37] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN13 Row, sBit, len = [37, 37], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN13(mCCDigit3PLMN13 uint8) {
	a.Octet[37] = (a.Octet[37] & 240) + (mCCDigit3PLMN13 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN13 Row, sBit, len = [38, 38], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN13() (mNCDigit2PLMN13 uint8) {
	return a.Octet[38] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN13 Row, sBit, len = [38, 38], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN13(mNCDigit2PLMN13 uint8) {
	a.Octet[38] = (a.Octet[38] & 15) + ((mNCDigit2PLMN13 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN13 Row, sBit, len = [38, 38], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN13() (mNCDigit1PLMN13 uint8) {
	return a.Octet[38] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN13 Row, sBit, len = [38, 38], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN13(mNCDigit1PLMN13 uint8) {
	a.Octet[38] = (a.Octet[38] & 240) + (mNCDigit1PLMN13 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN14 Row, sBit, len = [39, 39], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN14() (mCCDigit2PLMN14 uint8) {
	return a.Octet[39] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN14 Row, sBit, len = [39, 39], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN14(mCCDigit2PLMN14 uint8) {
	a.Octet[39] = (a.Octet[39] & 15) + ((mCCDigit2PLMN14 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN14 Row, sBit, len = [39, 39], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN14() (mCCDigit1PLMN14 uint8) {
	return a.Octet[39] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN14 Row, sBit, len = [39, 39], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN14(mCCDigit1PLMN14 uint8) {
	a.Octet[39] = (a.Octet[39] & 240) + (mCCDigit1PLMN14 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN14 Row, sBit, len = [40, 40], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN14() (mNCDigit3PLMN14 uint8) {
	return a.Octet[40] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN14 Row, sBit, len = [40, 40], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN14(mNCDigit3PLMN14 uint8) {
	a.Octet[40] = (a.Octet[40] & 15) + ((mNCDigit3PLMN14 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN14 Row, sBit, len = [40, 40], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN14() (mCCDigit3PLMN14 uint8) {
	return a.Octet[40] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN14 Row, sBit, len = [40, 40], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN14(mCCDigit3PLMN14 uint8) {
	a.Octet[40] = (a.Octet[40] & 240) + (mCCDigit3PLMN14 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN14 Row, sBit, len = [41, 41], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN14() (mNCDigit2PLMN14 uint8) {
	return a.Octet[41] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN14 Row, sBit, len = [41, 41], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN14(mNCDigit2PLMN14 uint8) {
	a.Octet[41] = (a.Octet[41] & 15) + ((mNCDigit2PLMN14 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN14 Row, sBit, len = [41, 41], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN14() (mNCDigit1PLMN14 uint8) {
	return a.Octet[41] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN14 Row, sBit, len = [41, 41], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN14(mNCDigit1PLMN14 uint8) {
	a.Octet[41] = (a.Octet[41] & 240) + (mNCDigit1PLMN14 & 15)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN15 Row, sBit, len = [42, 42], 8 , 4
func (a *EquivalentPlmns) GetMCCDigit2PLMN15() (mCCDigit2PLMN15 uint8) {
	return a.Octet[42] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit2PLMN15 Row, sBit, len = [42, 42], 8 , 4
func (a *EquivalentPlmns) SetMCCDigit2PLMN15(mCCDigit2PLMN15 uint8) {
	a.Octet[42] = (a.Octet[42] & 15) + ((mCCDigit2PLMN15 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN15 Row, sBit, len = [42, 42], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit1PLMN15() (mCCDigit1PLMN15 uint8) {
	return a.Octet[42] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit1PLMN15 Row, sBit, len = [42, 42], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit1PLMN15(mCCDigit1PLMN15 uint8) {
	a.Octet[42] = (a.Octet[42] & 240) + (mCCDigit1PLMN15 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN15 Row, sBit, len = [43, 43], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit3PLMN15() (mNCDigit3PLMN15 uint8) {
	return a.Octet[43] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit3PLMN15 Row, sBit, len = [43, 43], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit3PLMN15(mNCDigit3PLMN15 uint8) {
	a.Octet[43] = (a.Octet[43] & 15) + ((mNCDigit3PLMN15 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN15 Row, sBit, len = [43, 43], 4 , 4
func (a *EquivalentPlmns) GetMCCDigit3PLMN15() (mCCDigit3PLMN15 uint8) {
	return a.Octet[43] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MCCDigit3PLMN15 Row, sBit, len = [43, 43], 4 , 4
func (a *EquivalentPlmns) SetMCCDigit3PLMN15(mCCDigit3PLMN15 uint8) {
	a.Octet[43] = (a.Octet[43] & 240) + (mCCDigit3PLMN15 & 15)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN15 Row, sBit, len = [44, 44], 8 , 4
func (a *EquivalentPlmns) GetMNCDigit2PLMN15() (mNCDigit2PLMN15 uint8) {
	return a.Octet[44] & GetBitMask(8, 4) >> (4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit2PLMN15 Row, sBit, len = [44, 44], 8 , 4
func (a *EquivalentPlmns) SetMNCDigit2PLMN15(mNCDigit2PLMN15 uint8) {
	a.Octet[44] = (a.Octet[44] & 15) + ((mNCDigit2PLMN15 & 15) << 4)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN15 Row, sBit, len = [44, 44], 4 , 4
func (a *EquivalentPlmns) GetMNCDigit1PLMN15() (mNCDigit1PLMN15 uint8) {
	return a.Octet[44] & GetBitMask(4, 0)
}

// EquivalentPlmns 9.11.3.45
// MNCDigit1PLMN15 Row, sBit, len = [44, 44], 4 , 4
func (a *EquivalentPlmns) SetMNCDigit1PLMN15(mNCDigit1PLMN15 uint8) {
	a.Octet[44] = (a.Octet[44] & 240) + (mNCDigit1PLMN15 & 15)
}
