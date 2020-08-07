package nasType

// AllowedSSCMode 9.11.4.5
// Iei Row, sBit, len = [0, 0], 8 , 4
// SSC3 Row, sBit, len = [0, 0], 3 , 1
// SSC2 Row, sBit, len = [0, 0], 2 , 1
// SSC1 Row, sBit, len = [0, 0], 1 , 1
type AllowedSSCMode struct {
	Octet uint8
}

func NewAllowedSSCMode(iei uint8) (allowedSSCMode *AllowedSSCMode) {
	allowedSSCMode = &AllowedSSCMode{}
	allowedSSCMode.SetIei(iei)
	return allowedSSCMode
}

// AllowedSSCMode 9.11.4.5
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *AllowedSSCMode) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// AllowedSSCMode 9.11.4.5
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *AllowedSSCMode) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// AllowedSSCMode 9.11.4.5
// SSC3 Row, sBit, len = [0, 0], 3 , 1
func (a *AllowedSSCMode) GetSSC3() (sSC3 uint8) {
	return a.Octet & GetBitMask(3, 2) >> (2)
}

// AllowedSSCMode 9.11.4.5
// SSC3 Row, sBit, len = [0, 0], 3 , 1
func (a *AllowedSSCMode) SetSSC3(sSC3 uint8) {
	a.Octet = (a.Octet & 251) + ((sSC3 & 1) << 2)
}

// AllowedSSCMode 9.11.4.5
// SSC2 Row, sBit, len = [0, 0], 2 , 1
func (a *AllowedSSCMode) GetSSC2() (sSC2 uint8) {
	return a.Octet & GetBitMask(2, 1) >> (1)
}

// AllowedSSCMode 9.11.4.5
// SSC2 Row, sBit, len = [0, 0], 2 , 1
func (a *AllowedSSCMode) SetSSC2(sSC2 uint8) {
	a.Octet = (a.Octet & 253) + ((sSC2 & 1) << 1)
}

// AllowedSSCMode 9.11.4.5
// SSC1 Row, sBit, len = [0, 0], 1 , 1
func (a *AllowedSSCMode) GetSSC1() (sSC1 uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// AllowedSSCMode 9.11.4.5
// SSC1 Row, sBit, len = [0, 0], 1 , 1
func (a *AllowedSSCMode) SetSSC1(sSC1 uint8) {
	a.Octet = (a.Octet & 254) + (sSC1 & 1)
}
