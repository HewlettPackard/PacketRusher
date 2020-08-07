package nasType

// SSCMode 9.11.4.16
// Iei Row, sBit, len = [0, 0], 8 , 4
// Spare Row, sBit, len = [0, 0], 4 , 1
// SSCMode Row, sBit, len = [0, 0], 3 , 3
type SSCMode struct {
	Octet uint8
}

func NewSSCMode(iei uint8) (sSCMode *SSCMode) {
	sSCMode = &SSCMode{}
	sSCMode.SetIei(iei)
	return sSCMode
}

// SSCMode 9.11.4.16
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *SSCMode) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// SSCMode 9.11.4.16
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *SSCMode) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// SSCMode 9.11.4.16
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *SSCMode) GetSpare() (spare uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// SSCMode 9.11.4.16
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *SSCMode) SetSpare(spare uint8) {
	a.Octet = (a.Octet & 247) + ((spare & 1) << 3)
}

// SSCMode 9.11.4.16
// SSCMode Row, sBit, len = [0, 0], 3 , 3
func (a *SSCMode) GetSSCMode() (sSCMode uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// SSCMode 9.11.4.16
// SSCMode Row, sBit, len = [0, 0], 3 , 3
func (a *SSCMode) SetSSCMode(sSCMode uint8) {
	a.Octet = (a.Octet & 248) + (sSCMode & 7)
}
