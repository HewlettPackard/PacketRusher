package nasType

// ServiceAcceptMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type ServiceAcceptMessageIdentity struct {
	Octet uint8
}

func NewServiceAcceptMessageIdentity() (serviceAcceptMessageIdentity *ServiceAcceptMessageIdentity) {
	serviceAcceptMessageIdentity = &ServiceAcceptMessageIdentity{}
	return serviceAcceptMessageIdentity
}

// ServiceAcceptMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *ServiceAcceptMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// ServiceAcceptMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *ServiceAcceptMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
