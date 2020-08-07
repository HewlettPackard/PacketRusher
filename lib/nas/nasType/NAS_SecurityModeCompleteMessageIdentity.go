package nasType

// SecurityModeCompleteMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
type SecurityModeCompleteMessageIdentity struct {
	Octet uint8
}

func NewSecurityModeCompleteMessageIdentity() (securityModeCompleteMessageIdentity *SecurityModeCompleteMessageIdentity) {
	securityModeCompleteMessageIdentity = &SecurityModeCompleteMessageIdentity{}
	return securityModeCompleteMessageIdentity
}

// SecurityModeCompleteMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeCompleteMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// SecurityModeCompleteMessageIdentity 9.6
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *SecurityModeCompleteMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
