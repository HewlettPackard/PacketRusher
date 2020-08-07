package nasType

// AuthenticationParameterAUTN 9.11.3.15
// AUTN Row, sBit, len = [0, 15], 8 , 128
type AuthenticationParameterAUTN struct {
	Iei   uint8
	Len   uint8
	Octet [16]uint8
}

func NewAuthenticationParameterAUTN(iei uint8) (authenticationParameterAUTN *AuthenticationParameterAUTN) {
	authenticationParameterAUTN = &AuthenticationParameterAUTN{}
	authenticationParameterAUTN.SetIei(iei)
	return authenticationParameterAUTN
}

// AuthenticationParameterAUTN 9.11.3.15
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterAUTN) GetIei() (iei uint8) {
	return a.Iei
}

// AuthenticationParameterAUTN 9.11.3.15
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterAUTN) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthenticationParameterAUTN 9.11.3.15
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterAUTN) GetLen() (len uint8) {
	return a.Len
}

// AuthenticationParameterAUTN 9.11.3.15
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationParameterAUTN) SetLen(len uint8) {
	a.Len = len
}

// AuthenticationParameterAUTN 9.11.3.15
// AUTN Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationParameterAUTN) GetAUTN() (aUTN [16]uint8) {
	copy(aUTN[:], a.Octet[0:16])
	return aUTN
}

// AuthenticationParameterAUTN 9.11.3.15
// AUTN Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationParameterAUTN) SetAUTN(aUTN [16]uint8) {
	copy(a.Octet[0:16], aUTN[:])
}
