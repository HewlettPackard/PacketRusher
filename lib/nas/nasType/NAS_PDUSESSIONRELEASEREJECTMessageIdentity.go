package nasType

// PDUSESSIONRELEASEREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type PDUSESSIONRELEASEREJECTMessageIdentity struct {
	Octet uint8
}

func NewPDUSESSIONRELEASEREJECTMessageIdentity() (pDUSESSIONRELEASEREJECTMessageIdentity *PDUSESSIONRELEASEREJECTMessageIdentity) {
	pDUSESSIONRELEASEREJECTMessageIdentity = &PDUSESSIONRELEASEREJECTMessageIdentity{}
	return pDUSESSIONRELEASEREJECTMessageIdentity
}

// PDUSESSIONRELEASEREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONRELEASEREJECTMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// PDUSESSIONRELEASEREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONRELEASEREJECTMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
