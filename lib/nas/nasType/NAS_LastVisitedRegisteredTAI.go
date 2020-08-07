package nasType

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit2 Row, sBit, len = [0, 0], 8 , 4
// MCCDigit1 Row, sBit, len = [0, 0], 4 , 4
// MNCDigit3 Row, sBit, len = [1, 1], 8 , 4
// MCCDigit3 Row, sBit, len = [1, 1], 4 , 4
// MNCDigit2 Row, sBit, len = [2, 2], 8 , 4
// MNCDigit1 Row, sBit, len = [2, 2], 4 , 4
// TAC Row, sBit, len = [3, 5], 8 , 24
type LastVisitedRegisteredTAI struct {
	Iei   uint8
	Octet [7]uint8
}

func NewLastVisitedRegisteredTAI(iei uint8) (lastVisitedRegisteredTAI *LastVisitedRegisteredTAI) {
	lastVisitedRegisteredTAI = &LastVisitedRegisteredTAI{}
	lastVisitedRegisteredTAI.SetIei(iei)
	return lastVisitedRegisteredTAI
}

// LastVisitedRegisteredTAI 9.11.3.8
// Iei Row, sBit, len = [], 8, 8
func (a *LastVisitedRegisteredTAI) GetIei() (iei uint8) {
	return a.Iei
}

// LastVisitedRegisteredTAI 9.11.3.8
// Iei Row, sBit, len = [], 8, 8
func (a *LastVisitedRegisteredTAI) SetIei(iei uint8) {
	a.Iei = iei
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit2 Row, sBit, len = [0, 0], 8 , 4
func (a *LastVisitedRegisteredTAI) GetMCCDigit2() (mCCDigit2 uint8) {
	return a.Octet[0] & GetBitMask(8, 4) >> (4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit2 Row, sBit, len = [0, 0], 8 , 4
func (a *LastVisitedRegisteredTAI) SetMCCDigit2(mCCDigit2 uint8) {
	a.Octet[0] = (a.Octet[0] & 15) + ((mCCDigit2 & 15) << 4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit1 Row, sBit, len = [0, 0], 4 , 4
func (a *LastVisitedRegisteredTAI) GetMCCDigit1() (mCCDigit1 uint8) {
	return a.Octet[0] & GetBitMask(4, 0)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit1 Row, sBit, len = [0, 0], 4 , 4
func (a *LastVisitedRegisteredTAI) SetMCCDigit1(mCCDigit1 uint8) {
	a.Octet[0] = (a.Octet[0] & 240) + (mCCDigit1 & 15)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit3 Row, sBit, len = [1, 1], 8 , 4
func (a *LastVisitedRegisteredTAI) GetMNCDigit3() (mNCDigit3 uint8) {
	return a.Octet[1] & GetBitMask(8, 4) >> (4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit3 Row, sBit, len = [1, 1], 8 , 4
func (a *LastVisitedRegisteredTAI) SetMNCDigit3(mNCDigit3 uint8) {
	a.Octet[1] = (a.Octet[1] & 15) + ((mNCDigit3 & 15) << 4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit3 Row, sBit, len = [1, 1], 4 , 4
func (a *LastVisitedRegisteredTAI) GetMCCDigit3() (mCCDigit3 uint8) {
	return a.Octet[1] & GetBitMask(4, 0)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MCCDigit3 Row, sBit, len = [1, 1], 4 , 4
func (a *LastVisitedRegisteredTAI) SetMCCDigit3(mCCDigit3 uint8) {
	a.Octet[1] = (a.Octet[1] & 240) + (mCCDigit3 & 15)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit2 Row, sBit, len = [2, 2], 8 , 4
func (a *LastVisitedRegisteredTAI) GetMNCDigit2() (mNCDigit2 uint8) {
	return a.Octet[2] & GetBitMask(8, 4) >> (4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit2 Row, sBit, len = [2, 2], 8 , 4
func (a *LastVisitedRegisteredTAI) SetMNCDigit2(mNCDigit2 uint8) {
	a.Octet[2] = (a.Octet[2] & 15) + ((mNCDigit2 & 15) << 4)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit1 Row, sBit, len = [2, 2], 4 , 4
func (a *LastVisitedRegisteredTAI) GetMNCDigit1() (mNCDigit1 uint8) {
	return a.Octet[2] & GetBitMask(4, 0)
}

// LastVisitedRegisteredTAI 9.11.3.8
// MNCDigit1 Row, sBit, len = [2, 2], 4 , 4
func (a *LastVisitedRegisteredTAI) SetMNCDigit1(mNCDigit1 uint8) {
	a.Octet[2] = (a.Octet[2] & 240) + (mNCDigit1 & 15)
}

// LastVisitedRegisteredTAI 9.11.3.8
// TAC Row, sBit, len = [3, 5], 8 , 24
func (a *LastVisitedRegisteredTAI) GetTAC() (tAC [3]uint8) {
	copy(tAC[:], a.Octet[3:6])
	return tAC
}

// LastVisitedRegisteredTAI 9.11.3.8
// TAC Row, sBit, len = [3, 5], 8 , 24
func (a *LastVisitedRegisteredTAI) SetTAC(tAC [3]uint8) {
	copy(a.Octet[3:6], tAC[:])
}
