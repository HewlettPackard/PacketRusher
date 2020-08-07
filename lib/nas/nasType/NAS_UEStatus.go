package nasType

// UEStatus 9.11.3.56
// N1ModeReg Row, sBit, len = [0, 0], 2 , 1
// S1ModeReg Row, sBit, len = [0, 0], 1 , 1
type UEStatus struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewUEStatus(iei uint8) (uEStatus *UEStatus) {
	uEStatus = &UEStatus{}
	uEStatus.SetIei(iei)
	return uEStatus
}

// UEStatus 9.11.3.56
// Iei Row, sBit, len = [], 8, 8
func (a *UEStatus) GetIei() (iei uint8) {
	return a.Iei
}

// UEStatus 9.11.3.56
// Iei Row, sBit, len = [], 8, 8
func (a *UEStatus) SetIei(iei uint8) {
	a.Iei = iei
}

// UEStatus 9.11.3.56
// Len Row, sBit, len = [], 8, 8
func (a *UEStatus) GetLen() (len uint8) {
	return a.Len
}

// UEStatus 9.11.3.56
// Len Row, sBit, len = [], 8, 8
func (a *UEStatus) SetLen(len uint8) {
	a.Len = len
}

// UEStatus 9.11.3.56
// N1ModeReg Row, sBit, len = [0, 0], 2 , 1
func (a *UEStatus) GetN1ModeReg() (n1ModeReg uint8) {
	return a.Octet & GetBitMask(2, 1) >> (1)
}

// UEStatus 9.11.3.56
// N1ModeReg Row, sBit, len = [0, 0], 2 , 1
func (a *UEStatus) SetN1ModeReg(n1ModeReg uint8) {
	a.Octet = (a.Octet & 253) + ((n1ModeReg & 1) << 1)
}

// UEStatus 9.11.3.56
// S1ModeReg Row, sBit, len = [0, 0], 1 , 1
func (a *UEStatus) GetS1ModeReg() (s1ModeReg uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// UEStatus 9.11.3.56
// S1ModeReg Row, sBit, len = [0, 0], 1 , 1
func (a *UEStatus) SetS1ModeReg(s1ModeReg uint8) {
	a.Octet = (a.Octet & 254) + (s1ModeReg & 1)
}
