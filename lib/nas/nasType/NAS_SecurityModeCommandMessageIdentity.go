package nasType

// SecurityModeCommandMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type SecurityModeCommandMessageIdentity struct {
	Octet uint8
}

func NewSecurityModeCommandMessageIdentity() (securityModeCommandMessageIdentity *SecurityModeCommandMessageIdentity) {
	securityModeCommandMessageIdentity = &SecurityModeCommandMessageIdentity{}
	return securityModeCommandMessageIdentity
}

// SecurityModeCommandMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeCommandMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// SecurityModeCommandMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeCommandMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
