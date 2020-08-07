package nasType

// IMEISV 9.11.3.4
// IdentityDigit1 Row, sBit, len = [0, 0], 8 , 4
// OddEvenIdic Row, sBit, len = [0, 0], 4 , 1
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
// IdentityDigitP_1 Row, sBit, len = [1, 1], 8 , 4
// IdentityDigitP Row, sBit, len = [1, 1], 4 , 4
// IdentityDigitP_3 Row, sBit, len = [2, 2], 8 , 4
// IdentityDigitP_2 Row, sBit, len = [2, 2], 4 , 4
// IdentityDigitP_5 Row, sBit, len = [3, 3], 8 , 4
// IdentityDigitP_4 Row, sBit, len = [3, 3], 4 , 4
// IdentityDigitP_7 Row, sBit, len = [4, 4], 8 , 4
// IdentityDigitP_6 Row, sBit, len = [4, 4], 4 , 4
// IdentityDigitP_9 Row, sBit, len = [5, 5], 8 , 4
// IdentityDigitP_8 Row, sBit, len = [5, 5], 4 , 4
// IdentityDigitP_11 Row, sBit, len = [6, 6], 8 , 4
// IdentityDigitP_10 Row, sBit, len = [6, 6], 4 , 4
// IdentityDigitP_13 Row, sBit, len = [7, 7], 8 , 4
// IdentityDigitP_12 Row, sBit, len = [7, 7], 4 , 4
// IdentityDigitP_15 Row, sBit, len = [8, 8], 8 , 4
// IdentityDigitP_14 Row, sBit, len = [8, 8], 4 , 4
type IMEISV struct {
	Iei   uint8
	Len   uint16
	Octet [9]uint8
}

func NewIMEISV(iei uint8) (iMEISV *IMEISV) {
	iMEISV = &IMEISV{}
	iMEISV.SetIei(iei)
	return iMEISV
}

// IMEISV 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *IMEISV) GetIei() (iei uint8) {
	return a.Iei
}

// IMEISV 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *IMEISV) SetIei(iei uint8) {
	a.Iei = iei
}

// IMEISV 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *IMEISV) GetLen() (len uint16) {
	return a.Len
}

// IMEISV 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *IMEISV) SetLen(len uint16) {
	a.Len = len
}

// IMEISV 9.11.3.4
// IdentityDigit1 Row, sBit, len = [0, 0], 8 , 4
func (a *IMEISV) GetIdentityDigit1() (identityDigit1 uint8) {
	return a.Octet[0] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigit1 Row, sBit, len = [0, 0], 8 , 4
func (a *IMEISV) SetIdentityDigit1(identityDigit1 uint8) {
	a.Octet[0] = (a.Octet[0] & 15) + ((identityDigit1 & 15) << 4)
}

// IMEISV 9.11.3.4
// OddEvenIdic Row, sBit, len = [0, 0], 4 , 1
func (a *IMEISV) GetOddEvenIdic() (oddEvenIdic uint8) {
	return a.Octet[0] & GetBitMask(4, 3) >> (3)
}

// IMEISV 9.11.3.4
// OddEvenIdic Row, sBit, len = [0, 0], 4 , 1
func (a *IMEISV) SetOddEvenIdic(oddEvenIdic uint8) {
	a.Octet[0] = (a.Octet[0] & 247) + ((oddEvenIdic & 1) << 3)
}

// IMEISV 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *IMEISV) GetTypeOfIdentity() (typeOfIdentity uint8) {
	return a.Octet[0] & GetBitMask(3, 0)
}

// IMEISV 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *IMEISV) SetTypeOfIdentity(typeOfIdentity uint8) {
	a.Octet[0] = (a.Octet[0] & 248) + (typeOfIdentity & 7)
}

// IMEISV 9.11.3.4
// IdentityDigitP_1 Row, sBit, len = [1, 1], 8 , 4
func (a *IMEISV) GetIdentityDigitP_1() (identityDigitP_1 uint8) {
	return a.Octet[1] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_1 Row, sBit, len = [1, 1], 8 , 4
func (a *IMEISV) SetIdentityDigitP_1(identityDigitP_1 uint8) {
	a.Octet[1] = (a.Octet[1] & 15) + ((identityDigitP_1 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP Row, sBit, len = [1, 1], 4 , 4
func (a *IMEISV) GetIdentityDigitP() (identityDigitP uint8) {
	return a.Octet[1] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP Row, sBit, len = [1, 1], 4 , 4
func (a *IMEISV) SetIdentityDigitP(identityDigitP uint8) {
	a.Octet[1] = (a.Octet[1] & 240) + (identityDigitP & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_3 Row, sBit, len = [2, 2], 8 , 4
func (a *IMEISV) GetIdentityDigitP_3() (identityDigitP_3 uint8) {
	return a.Octet[2] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_3 Row, sBit, len = [2, 2], 8 , 4
func (a *IMEISV) SetIdentityDigitP_3(identityDigitP_3 uint8) {
	a.Octet[2] = (a.Octet[2] & 15) + ((identityDigitP_3 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_2 Row, sBit, len = [2, 2], 4 , 4
func (a *IMEISV) GetIdentityDigitP_2() (identityDigitP_2 uint8) {
	return a.Octet[2] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_2 Row, sBit, len = [2, 2], 4 , 4
func (a *IMEISV) SetIdentityDigitP_2(identityDigitP_2 uint8) {
	a.Octet[2] = (a.Octet[2] & 240) + (identityDigitP_2 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_5 Row, sBit, len = [3, 3], 8 , 4
func (a *IMEISV) GetIdentityDigitP_5() (identityDigitP_5 uint8) {
	return a.Octet[3] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_5 Row, sBit, len = [3, 3], 8 , 4
func (a *IMEISV) SetIdentityDigitP_5(identityDigitP_5 uint8) {
	a.Octet[3] = (a.Octet[3] & 15) + ((identityDigitP_5 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_4 Row, sBit, len = [3, 3], 4 , 4
func (a *IMEISV) GetIdentityDigitP_4() (identityDigitP_4 uint8) {
	return a.Octet[3] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_4 Row, sBit, len = [3, 3], 4 , 4
func (a *IMEISV) SetIdentityDigitP_4(identityDigitP_4 uint8) {
	a.Octet[3] = (a.Octet[3] & 240) + (identityDigitP_4 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_7 Row, sBit, len = [4, 4], 8 , 4
func (a *IMEISV) GetIdentityDigitP_7() (identityDigitP_7 uint8) {
	return a.Octet[4] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_7 Row, sBit, len = [4, 4], 8 , 4
func (a *IMEISV) SetIdentityDigitP_7(identityDigitP_7 uint8) {
	a.Octet[4] = (a.Octet[4] & 15) + ((identityDigitP_7 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_6 Row, sBit, len = [4, 4], 4 , 4
func (a *IMEISV) GetIdentityDigitP_6() (identityDigitP_6 uint8) {
	return a.Octet[4] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_6 Row, sBit, len = [4, 4], 4 , 4
func (a *IMEISV) SetIdentityDigitP_6(identityDigitP_6 uint8) {
	a.Octet[4] = (a.Octet[4] & 240) + (identityDigitP_6 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_9 Row, sBit, len = [5, 5], 8 , 4
func (a *IMEISV) GetIdentityDigitP_9() (identityDigitP_9 uint8) {
	return a.Octet[5] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_9 Row, sBit, len = [5, 5], 8 , 4
func (a *IMEISV) SetIdentityDigitP_9(identityDigitP_9 uint8) {
	a.Octet[5] = (a.Octet[5] & 15) + ((identityDigitP_9 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_8 Row, sBit, len = [5, 5], 4 , 4
func (a *IMEISV) GetIdentityDigitP_8() (identityDigitP_8 uint8) {
	return a.Octet[5] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_8 Row, sBit, len = [5, 5], 4 , 4
func (a *IMEISV) SetIdentityDigitP_8(identityDigitP_8 uint8) {
	a.Octet[5] = (a.Octet[5] & 240) + (identityDigitP_8 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_11 Row, sBit, len = [6, 6], 8 , 4
func (a *IMEISV) GetIdentityDigitP_11() (identityDigitP_11 uint8) {
	return a.Octet[6] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_11 Row, sBit, len = [6, 6], 8 , 4
func (a *IMEISV) SetIdentityDigitP_11(identityDigitP_11 uint8) {
	a.Octet[6] = (a.Octet[6] & 15) + ((identityDigitP_11 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_10 Row, sBit, len = [6, 6], 4 , 4
func (a *IMEISV) GetIdentityDigitP_10() (identityDigitP_10 uint8) {
	return a.Octet[6] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_10 Row, sBit, len = [6, 6], 4 , 4
func (a *IMEISV) SetIdentityDigitP_10(identityDigitP_10 uint8) {
	a.Octet[6] = (a.Octet[6] & 240) + (identityDigitP_10 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_13 Row, sBit, len = [7, 7], 8 , 4
func (a *IMEISV) GetIdentityDigitP_13() (identityDigitP_13 uint8) {
	return a.Octet[7] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_13 Row, sBit, len = [7, 7], 8 , 4
func (a *IMEISV) SetIdentityDigitP_13(identityDigitP_13 uint8) {
	a.Octet[7] = (a.Octet[7] & 15) + ((identityDigitP_13 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_12 Row, sBit, len = [7, 7], 4 , 4
func (a *IMEISV) GetIdentityDigitP_12() (identityDigitP_12 uint8) {
	return a.Octet[7] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_12 Row, sBit, len = [7, 7], 4 , 4
func (a *IMEISV) SetIdentityDigitP_12(identityDigitP_12 uint8) {
	a.Octet[7] = (a.Octet[7] & 240) + (identityDigitP_12 & 15)
}

// IMEISV 9.11.3.4
// IdentityDigitP_15 Row, sBit, len = [8, 8], 8 , 4
func (a *IMEISV) GetIdentityDigitP_15() (identityDigitP_15 uint8) {
	return a.Octet[8] & GetBitMask(8, 4) >> (4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_15 Row, sBit, len = [8, 8], 8 , 4
func (a *IMEISV) SetIdentityDigitP_15(identityDigitP_15 uint8) {
	a.Octet[8] = (a.Octet[8] & 15) + ((identityDigitP_15 & 15) << 4)
}

// IMEISV 9.11.3.4
// IdentityDigitP_14 Row, sBit, len = [8, 8], 4 , 4
func (a *IMEISV) GetIdentityDigitP_14() (identityDigitP_14 uint8) {
	return a.Octet[8] & GetBitMask(4, 0)
}

// IMEISV 9.11.3.4
// IdentityDigitP_14 Row, sBit, len = [8, 8], 4 , 4
func (a *IMEISV) SetIdentityDigitP_14(identityDigitP_14 uint8) {
	a.Octet[8] = (a.Octet[8] & 240) + (identityDigitP_14 & 15)
}
