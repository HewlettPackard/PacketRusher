package nasType

// NotificationMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type NotificationMessageIdentity struct {
	Octet uint8
}

func NewNotificationMessageIdentity() (notificationMessageIdentity *NotificationMessageIdentity) {
	notificationMessageIdentity = &NotificationMessageIdentity{}
	return notificationMessageIdentity
}

// NotificationMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *NotificationMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// NotificationMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *NotificationMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
