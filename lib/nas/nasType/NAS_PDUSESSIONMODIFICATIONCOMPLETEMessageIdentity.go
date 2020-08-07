package nasType

// PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity struct {
	Octet uint8
}

func NewPDUSESSIONMODIFICATIONCOMPLETEMessageIdentity() (pDUSESSIONMODIFICATIONCOMPLETEMessageIdentity *PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity) {
	pDUSESSIONMODIFICATIONCOMPLETEMessageIdentity = &PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity{}
	return pDUSESSIONMODIFICATIONCOMPLETEMessageIdentity
}

// PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
