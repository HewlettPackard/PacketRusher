package nasType

// RegistrationCompleteMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type RegistrationCompleteMessageIdentity struct {
	Octet uint8
}

func NewRegistrationCompleteMessageIdentity() (registrationCompleteMessageIdentity *RegistrationCompleteMessageIdentity) {
	registrationCompleteMessageIdentity = &RegistrationCompleteMessageIdentity{}
	return registrationCompleteMessageIdentity
}

// RegistrationCompleteMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *RegistrationCompleteMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// RegistrationCompleteMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *RegistrationCompleteMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
