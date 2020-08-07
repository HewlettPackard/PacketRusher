package nasType

// IdentityResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type IdentityResponseMessageIdentity struct {
	Octet uint8
}

func NewIdentityResponseMessageIdentity() (identityResponseMessageIdentity *IdentityResponseMessageIdentity) {
	identityResponseMessageIdentity = &IdentityResponseMessageIdentity{}
	return identityResponseMessageIdentity
}

// IdentityResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *IdentityResponseMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// IdentityResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *IdentityResponseMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
