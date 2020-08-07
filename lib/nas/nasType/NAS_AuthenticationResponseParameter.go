package nasType

// AuthenticationResponseParameter 9.11.3.17
// RES Row, sBit, len = [0, 15], 8 , 128
type AuthenticationResponseParameter struct {
	Iei   uint8
	Len   uint8
	Octet [16]uint8
}

func NewAuthenticationResponseParameter(iei uint8) (authenticationResponseParameter *AuthenticationResponseParameter) {
	authenticationResponseParameter = &AuthenticationResponseParameter{}
	authenticationResponseParameter.SetIei(iei)
	return authenticationResponseParameter
}

// AuthenticationResponseParameter 9.11.3.17
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationResponseParameter) GetIei() (iei uint8) {
	return a.Iei
}

// AuthenticationResponseParameter 9.11.3.17
// Iei Row, sBit, len = [], 8, 8
func (a *AuthenticationResponseParameter) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthenticationResponseParameter 9.11.3.17
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationResponseParameter) GetLen() (len uint8) {
	return a.Len
}

// AuthenticationResponseParameter 9.11.3.17
// Len Row, sBit, len = [], 8, 8
func (a *AuthenticationResponseParameter) SetLen(len uint8) {
	a.Len = len
}

// AuthenticationResponseParameter 9.11.3.17
// RES Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationResponseParameter) GetRES() (rES [16]uint8) {
	copy(rES[:], a.Octet[0:16])
	return rES
}

// AuthenticationResponseParameter 9.11.3.17
// RES Row, sBit, len = [0, 15], 8 , 128
func (a *AuthenticationResponseParameter) SetRES(rES [16]uint8) {
	copy(a.Octet[0:16], rES[:])
}
