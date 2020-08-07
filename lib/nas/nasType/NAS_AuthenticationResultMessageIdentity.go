package nasType

// AuthenticationResultMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type AuthenticationResultMessageIdentity struct {
	Octet uint8
}

func NewAuthenticationResultMessageIdentity() (authenticationResultMessageIdentity *AuthenticationResultMessageIdentity) {
	authenticationResultMessageIdentity = &AuthenticationResultMessageIdentity{}
	return authenticationResultMessageIdentity
}

// AuthenticationResultMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationResultMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// AuthenticationResultMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationResultMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
