package nasType

// PDUSESSIONRELEASECOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type PDUSESSIONRELEASECOMPLETEMessageIdentity struct {
	Octet uint8
}

func NewPDUSESSIONRELEASECOMPLETEMessageIdentity() (pDUSESSIONRELEASECOMPLETEMessageIdentity *PDUSESSIONRELEASECOMPLETEMessageIdentity) {
	pDUSESSIONRELEASECOMPLETEMessageIdentity = &PDUSESSIONRELEASECOMPLETEMessageIdentity{}
	return pDUSESSIONRELEASECOMPLETEMessageIdentity
}

// PDUSESSIONRELEASECOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONRELEASECOMPLETEMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// PDUSESSIONRELEASECOMPLETEMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONRELEASECOMPLETEMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
