package nasType

// ServiceRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type ServiceRejectMessageIdentity struct {
	Octet uint8
}

func NewServiceRejectMessageIdentity() (serviceRejectMessageIdentity *ServiceRejectMessageIdentity) {
	serviceRejectMessageIdentity = &ServiceRejectMessageIdentity{}
	return serviceRejectMessageIdentity
}

// ServiceRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *ServiceRejectMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// ServiceRejectMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *ServiceRejectMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
