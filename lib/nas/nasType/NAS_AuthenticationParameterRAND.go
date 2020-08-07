package nasType

// AuthenticationParameterRAND 9.11.3.16
// RANDValue Row, sBit, len = [0, 15], 8 , 128
type AuthenticationParameterRAND struct {
	Iei   uint8
	Octet [16]uint8
}

func NewAuthenticationParameterRAND(iei uint8) (authenticationParameterRAND *AuthenticationParameterRAND) {
	authenticationParameterRAND = &AuthenticationParameterRAND{}
	authenticationParameterRAND.SetIei(iei)
	return authenticationParameterRAND
}

// AuthenticationParameterRAND 9.11.3.16
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterRAND) GetIei() (iei uint8) {
	return a.Iei
}

// AuthenticationParameterRAND 9.11.3.16
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterRAND) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthenticationParameterRAND 9.11.3.16
// RANDValue Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationParameterRAND) GetRANDValue() (rANDValue [16]uint8) {
	copy(rANDValue[:], a.Octet[0:16])
	return rANDValue
}

// AuthenticationParameterRAND 9.11.3.16
// RANDValue Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationParameterRAND) SetRANDValue(rANDValue [16]uint8) {
	copy(a.Octet[0:16], rANDValue[:])
}
