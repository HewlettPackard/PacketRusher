package nasType

// AuthenticationRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type AuthenticationRejectMessageIdentity struct {
	Octet uint8
}

func NewAuthenticationRejectMessageIdentity() (authenticationRejectMessageIdentity *AuthenticationRejectMessageIdentity) {
	authenticationRejectMessageIdentity = &AuthenticationRejectMessageIdentity{}
	return authenticationRejectMessageIdentity
}

// AuthenticationRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationRejectMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// AuthenticationRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *AuthenticationRejectMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
