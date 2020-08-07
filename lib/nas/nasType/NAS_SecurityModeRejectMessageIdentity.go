package nasType

// SecurityModeRejectMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
type SecurityModeRejectMessageIdentity struct {
	Octet uint8
}

func NewSecurityModeRejectMessageIdentity() (securityModeRejectMessageIdentity *SecurityModeRejectMessageIdentity) {
	securityModeRejectMessageIdentity = &SecurityModeRejectMessageIdentity{}
	return securityModeRejectMessageIdentity
}

// SecurityModeRejectMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeRejectMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// SecurityModeRejectMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeRejectMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
