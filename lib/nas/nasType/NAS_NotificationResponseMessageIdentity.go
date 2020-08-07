package nasType

// NotificationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type NotificationResponseMessageIdentity struct {
	Octet uint8
}

func NewNotificationResponseMessageIdentity() (notificationResponseMessageIdentity *NotificationResponseMessageIdentity) {
	notificationResponseMessageIdentity = &NotificationResponseMessageIdentity{}
	return notificationResponseMessageIdentity
}

// NotificationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *NotificationResponseMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// NotificationResponseMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *NotificationResponseMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
