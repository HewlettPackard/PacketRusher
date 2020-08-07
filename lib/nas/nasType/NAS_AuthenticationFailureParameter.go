package nasType

// AuthenticationFailureParameter 9.11.3.14
// AuthenticationFailureParameter Row, sBit, len = [0, 13], 8 , 112
type AuthenticationFailureParameter struct {
	Iei   uint8
	Len   uint8
	Octet [14]uint8
}

func NewAuthenticationFailureParameter(iei uint8) (authenticationFailureParameter *AuthenticationFailureParameter) {
	authenticationFailureParameter = &AuthenticationFailureParameter{}
	authenticationFailureParameter.SetIei(iei)
	return authenticationFailureParameter
}

// AuthenticationFailureParameter 9.11.3.14
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationFailureParameter) GetIei() (iei uint8) {
	return a.Iei
}

// AuthenticationFailureParameter 9.11.3.14
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationFailureParameter) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthenticationFailureParameter 9.11.3.14
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationFailureParameter) GetLen() (len uint8) {
	return a.Len
}

// AuthenticationFailureParameter 9.11.3.14
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationFailureParameter) SetLen(len uint8) {
	a.Len = len
}

// AuthenticationFailureParameter 9.11.3.14
// AuthenticationFailureParameter Row, sBit, len = [0, 13], 8 , 112
func (a *AuthenticationFailureParameter) GetAuthenticationFailureParameter() (authenticationFailureParameter [14]uint8) {
	copy(authenticationFailureParameter[:], a.Octet[0:14])
	return authenticationFailureParameter
}

// AuthenticationFailureParameter 9.11.3.14
// AuthenticationFailureParameter Row, sBit, len = [0, 13], 8 , 112
func (a *AuthenticationFailureParameter) SetAuthenticationFailureParameter(authenticationFailureParameter [14]uint8) {
	copy(a.Octet[0:14], authenticationFailureParameter[:])
}
