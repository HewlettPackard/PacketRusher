package nasType

// AdditionalGUTI 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
// MCCDigit2 Row, sBit, len = [1, 1], 8 , 4
// MCCDigit1 Row, sBit, len = [1, 1], 4 , 4
// MNCDigit3 Row, sBit, len = [2, 2], 8 , 4
// MCCDigit3 Row, sBit, len = [2, 2], 4 , 4
// MNCDigit2 Row, sBit, len = [3, 3], 8 , 4
// MNCDigit1 Row, sBit, len = [3, 3], 4 , 4
// AMFRegionID Row, sBit, len = [4, 4], 8 , 8
// AMFSetID Row, sBit, len = [5, 6], 8 , 10
// AMFPointer Row, sBit, len = [6, 6], 6 , 6
// TMSI5G Row, sBit, len = [7, 10], 8 , 32
type AdditionalGUTI struct {
	Iei   uint8
	Len   uint16
	Octet [11]uint8
}

func NewAdditionalGUTI(iei uint8) (additionalGUTI *AdditionalGUTI) {
	additionalGUTI = &AdditionalGUTI{}
	additionalGUTI.SetIei(iei)
	return additionalGUTI
}

// AdditionalGUTI 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *AdditionalGUTI) GetIei() (iei uint8) {
	return a.Iei
}

// AdditionalGUTI 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *AdditionalGUTI) SetIei(iei uint8) {
	a.Iei = iei
}

// AdditionalGUTI 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *AdditionalGUTI) GetLen() (len uint16) {
	return a.Len
}

// AdditionalGUTI 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *AdditionalGUTI) SetLen(len uint16) {
	a.Len = len
}

// AdditionalGUTI 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *AdditionalGUTI) GetSpare() (spare uint8) {
	return a.Octet[0] & GetBitMask(4, 3) >> (3)
}

// AdditionalGUTI 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *AdditionalGUTI) SetSpare(spare uint8) {
	a.Octet[0] = (a.Octet[0] & 247) + ((spare & 1) << 3)
}

// AdditionalGUTI 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *AdditionalGUTI) GetTypeOfIdentity() (typeOfIdentity uint8) {
	return a.Octet[0] & GetBitMask(3, 0)
}

// AdditionalGUTI 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *AdditionalGUTI) SetTypeOfIdentity(typeOfIdentity uint8) {
	a.Octet[0] = (a.Octet[0] & 248) + (typeOfIdentity & 7)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit2 Row, sBit, len = [1, 1], 8 , 4
func (a *AdditionalGUTI) GetMCCDigit2() (mCCDigit2 uint8) {
	return a.Octet[1] & GetBitMask(8, 4) >> (4)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit2 Row, sBit, len = [1, 1], 8 , 4
func (a *AdditionalGUTI) SetMCCDigit2(mCCDigit2 uint8) {
	a.Octet[1] = (a.Octet[1] & 15) + ((mCCDigit2 & 15) << 4)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit1 Row, sBit, len = [1, 1], 4 , 4
func (a *AdditionalGUTI) GetMCCDigit1() (mCCDigit1 uint8) {
	return a.Octet[1] & GetBitMask(4, 0)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit1 Row, sBit, len = [1, 1], 4 , 4
func (a *AdditionalGUTI) SetMCCDigit1(mCCDigit1 uint8) {
	a.Octet[1] = (a.Octet[1] & 240) + (mCCDigit1 & 15)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit3 Row, sBit, len = [2, 2], 8 , 4
func (a *AdditionalGUTI) GetMNCDigit3() (mNCDigit3 uint8) {
	return a.Octet[2] & GetBitMask(8, 4) >> (4)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit3 Row, sBit, len = [2, 2], 8 , 4
func (a *AdditionalGUTI) SetMNCDigit3(mNCDigit3 uint8) {
	a.Octet[2] = (a.Octet[2] & 15) + ((mNCDigit3 & 15) << 4)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit3 Row, sBit, len = [2, 2], 4 , 4
func (a *AdditionalGUTI) GetMCCDigit3() (mCCDigit3 uint8) {
	return a.Octet[2] & GetBitMask(4, 0)
}

// AdditionalGUTI 9.11.3.4
// MCCDigit3 Row, sBit, len = [2, 2], 4 , 4
func (a *AdditionalGUTI) SetMCCDigit3(mCCDigit3 uint8) {
	a.Octet[2] = (a.Octet[2] & 240) + (mCCDigit3 & 15)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit2 Row, sBit, len = [3, 3], 8 , 4
func (a *AdditionalGUTI) GetMNCDigit2() (mNCDigit2 uint8) {
	return a.Octet[3] & GetBitMask(8, 4) >> (4)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit2 Row, sBit, len = [3, 3], 8 , 4
func (a *AdditionalGUTI) SetMNCDigit2(mNCDigit2 uint8) {
	a.Octet[3] = (a.Octet[3] & 15) + ((mNCDigit2 & 15) << 4)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit1 Row, sBit, len = [3, 3], 4 , 4
func (a *AdditionalGUTI) GetMNCDigit1() (mNCDigit1 uint8) {
	return a.Octet[3] & GetBitMask(4, 0)
}

// AdditionalGUTI 9.11.3.4
// MNCDigit1 Row, sBit, len = [3, 3], 4 , 4
func (a *AdditionalGUTI) SetMNCDigit1(mNCDigit1 uint8) {
	a.Octet[3] = (a.Octet[3] & 240) + (mNCDigit1 & 15)
}

// AdditionalGUTI 9.11.3.4
// AMFRegionID Row, sBit, len = [4, 4], 8 , 8
func (a *AdditionalGUTI) GetAMFRegionID() (aMFRegionID uint8) {
	return a.Octet[4]
}

// AdditionalGUTI 9.11.3.4
// AMFRegionID Row, sBit, len = [4, 4], 8 , 8
func (a *AdditionalGUTI) SetAMFRegionID(aMFRegionID uint8) {
	a.Octet[4] = aMFRegionID
}

// AdditionalGUTI 9.11.3.4
// AMFSetID Row, sBit, len = [5, 6], 8 , 10
func (a *AdditionalGUTI) GetAMFSetID() (aMFSetID uint16) {
	return (uint16(a.Octet[5])<<2 + uint16((a.Octet[6])&GetBitMask(8, 2))>>6)
}

// AdditionalGUTI 9.11.3.4
// AMFSetID Row, sBit, len = [5, 6], 8 , 10
func (a *AdditionalGUTI) SetAMFSetID(aMFSetID uint16) {
	a.Octet[5] = uint8((aMFSetID)>>2) & 255
	a.Octet[6] = a.Octet[6]&GetBitMask(6, 6) + uint8(aMFSetID&3)<<6
}

// AdditionalGUTI 9.11.3.4
// AMFPointer Row, sBit, len = [6, 6], 6 , 6
func (a *AdditionalGUTI) GetAMFPointer() (aMFPointer uint8) {
	return a.Octet[6] & GetBitMask(6, 0)
}

// AdditionalGUTI 9.11.3.4
// AMFPointer Row, sBit, len = [6, 6], 6 , 6
func (a *AdditionalGUTI) SetAMFPointer(aMFPointer uint8) {
	a.Octet[6] = (a.Octet[6] & 192) + (aMFPointer & 63)
}

// AdditionalGUTI 9.11.3.4
// TMSI5G Row, sBit, len = [7, 10], 8 , 32
func (a *AdditionalGUTI) GetTMSI5G() (tMSI5G [4]uint8) {
	copy(tMSI5G[:], a.Octet[7:11])
	return tMSI5G
}

// AdditionalGUTI 9.11.3.4
// TMSI5G Row, sBit, len = [7, 10], 8 , 32
func (a *AdditionalGUTI) SetTMSI5G(tMSI5G [4]uint8) {
	copy(a.Octet[7:11], tMSI5G[:])
}
