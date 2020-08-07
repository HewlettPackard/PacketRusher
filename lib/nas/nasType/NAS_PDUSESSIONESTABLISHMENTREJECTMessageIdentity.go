package nasType

// PDUSESSIONESTABLISHMENTREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type PDUSESSIONESTABLISHMENTREJECTMessageIdentity struct {
	Octet uint8
}

func NewPDUSESSIONESTABLISHMENTREJECTMessageIdentity() (pDUSESSIONESTABLISHMENTREJECTMessageIdentity *PDUSESSIONESTABLISHMENTREJECTMessageIdentity) {
	pDUSESSIONESTABLISHMENTREJECTMessageIdentity = &PDUSESSIONESTABLISHMENTREJECTMessageIdentity{}
	return pDUSESSIONESTABLISHMENTREJECTMessageIdentity
}

// PDUSESSIONESTABLISHMENTREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONESTABLISHMENTREJECTMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// PDUSESSIONESTABLISHMENTREJECTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONESTABLISHMENTREJECTMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
