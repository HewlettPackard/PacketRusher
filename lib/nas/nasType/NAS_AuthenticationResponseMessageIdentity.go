package nasType

// AuthenticationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type AuthenticationResponseMessageIdentity struct {
	Octet uint8
}

func NewAuthenticationResponseMessageIdentity() (authenticationResponseMessageIdentity *AuthenticationResponseMessageIdentity) {
	authenticationResponseMessageIdentity = &AuthenticationResponseMessageIdentity{}
	return authenticationResponseMessageIdentity
}

// AuthenticationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationResponseMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// AuthenticationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationResponseMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
