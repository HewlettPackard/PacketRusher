package nasType

// GUTI5G 9.11.3.4
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
type GUTI5G struct {
	Iei   uint8
	Len   uint16
	Octet [11]uint8
}

func NewGUTI5G(iei uint8) (gUTI5G *GUTI5G) {
	gUTI5G = &GUTI5G{}
	gUTI5G.SetIei(iei)
	return gUTI5G
}

// GUTI5G 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *GUTI5G) GetIei() (iei uint8) {
	return a.Iei
}

// GUTI5G 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *GUTI5G) SetIei(iei uint8) {
	a.Iei = iei
}

// GUTI5G 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *GUTI5G) GetLen() (len uint16) {
	return a.Len
}

// GUTI5G 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *GUTI5G) SetLen(len uint16) {
	a.Len = len
}

// GUTI5G 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *GUTI5G) GetSpare() (spare uint8) {
	return a.Octet[0] & GetBitMask(4, 3) >> (3)
}

// GUTI5G 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *GUTI5G) SetSpare(spare uint8) {
	a.Octet[0] = (a.Octet[0] & 247) + ((spare & 1) << 3)
}

// GUTI5G 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *GUTI5G) GetTypeOfIdentity() (typeOfIdentity uint8) {
	return a.Octet[0] & GetBitMask(3, 0)
}

// GUTI5G 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *GUTI5G) SetTypeOfIdentity(typeOfIdentity uint8) {
	a.Octet[0] = (a.Octet[0] & 248) + (typeOfIdentity & 7)
}

// GUTI5G 9.11.3.4
// MCCDigit2 Row, sBit, len = [1, 1], 8 , 4
func (a *GUTI5G) GetMCCDigit2() (mCCDigit2 uint8) {
	return a.Octet[1] & GetBitMask(8, 4) >> (4)
}

// GUTI5G 9.11.3.4
// MCCDigit2 Row, sBit, len = [1, 1], 8 , 4
func (a *GUTI5G) SetMCCDigit2(mCCDigit2 uint8) {
	a.Octet[1] = (a.Octet[1] & 15) + ((mCCDigit2 & 15) << 4)
}

// GUTI5G 9.11.3.4
// MCCDigit1 Row, sBit, len = [1, 1], 4 , 4
func (a *GUTI5G) GetMCCDigit1() (mCCDigit1 uint8) {
	return a.Octet[1] & GetBitMask(4, 0)
}

// GUTI5G 9.11.3.4
// MCCDigit1 Row, sBit, len = [1, 1], 4 , 4
func (a *GUTI5G) SetMCCDigit1(mCCDigit1 uint8) {
	a.Octet[1] = (a.Octet[1] & 240) + (mCCDigit1 & 15)
}

// GUTI5G 9.11.3.4
// MNCDigit3 Row, sBit, len = [2, 2], 8 , 4
func (a *GUTI5G) GetMNCDigit3() (mNCDigit3 uint8) {
	return a.Octet[2] & GetBitMask(8, 4) >> (4)
}

// GUTI5G 9.11.3.4
// MNCDigit3 Row, sBit, len = [2, 2], 8 , 4
func (a *GUTI5G) SetMNCDigit3(mNCDigit3 uint8) {
	a.Octet[2] = (a.Octet[2] & 15) + ((mNCDigit3 & 15) << 4)
}

// GUTI5G 9.11.3.4
// MCCDigit3 Row, sBit, len = [2, 2], 4 , 4
func (a *GUTI5G) GetMCCDigit3() (mCCDigit3 uint8) {
	return a.Octet[2] & GetBitMask(4, 0)
}

// GUTI5G 9.11.3.4
// MCCDigit3 Row, sBit, len = [2, 2], 4 , 4
func (a *GUTI5G) SetMCCDigit3(mCCDigit3 uint8) {
	a.Octet[2] = (a.Octet[2] & 240) + (mCCDigit3 & 15)
}

// GUTI5G 9.11.3.4
// MNCDigit2 Row, sBit, len = [3, 3], 8 , 4
func (a *GUTI5G) GetMNCDigit2() (mNCDigit2 uint8) {
	return a.Octet[3] & GetBitMask(8, 4) >> (4)
}

// GUTI5G 9.11.3.4
// MNCDigit2 Row, sBit, len = [3, 3], 8 , 4
func (a *GUTI5G) SetMNCDigit2(mNCDigit2 uint8) {
	a.Octet[3] = (a.Octet[3] & 15) + ((mNCDigit2 & 15) << 4)
}

// GUTI5G 9.11.3.4
// MNCDigit1 Row, sBit, len = [3, 3], 4 , 4
func (a *GUTI5G) GetMNCDigit1() (mNCDigit1 uint8) {
	return a.Octet[3] & GetBitMask(4, 0)
}

// GUTI5G 9.11.3.4
// MNCDigit1 Row, sBit, len = [3, 3], 4 , 4
func (a *GUTI5G) SetMNCDigit1(mNCDigit1 uint8) {
	a.Octet[3] = (a.Octet[3] & 240) + (mNCDigit1 & 15)
}

// GUTI5G 9.11.3.4
// AMFRegionID Row, sBit, len = [4, 4], 8 , 8
func (a *GUTI5G) GetAMFRegionID() (aMFRegionID uint8) {
	return a.Octet[4]
}

// GUTI5G 9.11.3.4
// AMFRegionID Row, sBit, len = [4, 4], 8 , 8
func (a *GUTI5G) SetAMFRegionID(aMFRegionID uint8) {
	a.Octet[4] = aMFRegionID
}

// GUTI5G 9.11.3.4
// AMFSetID Row, sBit, len = [5, 6], 8 , 10
func (a *GUTI5G) GetAMFSetID() (aMFSetID uint16) {
	return (uint16(a.Octet[5])<<2 + uint16((a.Octet[6])&GetBitMask(8, 2))>>6)
}

// GUTI5G 9.11.3.4
// AMFSetID Row, sBit, len = [5, 6], 8 , 10
func (a *GUTI5G) SetAMFSetID(aMFSetID uint16) {
	a.Octet[5] = uint8((aMFSetID)>>2) & 255
	a.Octet[6] = a.Octet[6]&GetBitMask(6, 6) + uint8(aMFSetID&3)<<6
}

// GUTI5G 9.11.3.4
// AMFPointer Row, sBit, len = [6, 6], 6 , 6
func (a *GUTI5G) GetAMFPointer() (aMFPointer uint8) {
	return a.Octet[6] & GetBitMask(6, 0)
}

// GUTI5G 9.11.3.4
// AMFPointer Row, sBit, len = [6, 6], 6 , 6
func (a *GUTI5G) SetAMFPointer(aMFPointer uint8) {
	a.Octet[6] = (a.Octet[6] & 192) + (aMFPointer & 63)
}

// GUTI5G 9.11.3.4
// TMSI5G Row, sBit, len = [7, 10], 8 , 32
func (a *GUTI5G) GetTMSI5G() (tMSI5G [4]uint8) {
	copy(tMSI5G[:], a.Octet[7:11])
	return tMSI5G
}

// GUTI5G 9.11.3.4
// TMSI5G Row, sBit, len = [7, 10], 8 , 32
func (a *GUTI5G) SetTMSI5G(tMSI5G [4]uint8) {
	copy(a.Octet[7:11], tMSI5G[:])
}
