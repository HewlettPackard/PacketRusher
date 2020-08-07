package nasType

// PDUSESSIONAUTHENTICATIONRESULTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type PDUSESSIONAUTHENTICATIONRESULTMessageIdentity struct {
	Octet uint8
}

func NewPDUSESSIONAUTHENTICATIONRESULTMessageIdentity() (pDUSESSIONAUTHENTICATIONRESULTMessageIdentity *PDUSESSIONAUTHENTICATIONRESULTMessageIdentity) {
	pDUSESSIONAUTHENTICATIONRESULTMessageIdentity = &PDUSESSIONAUTHENTICATIONRESULTMessageIdentity{}
	return pDUSESSIONAUTHENTICATIONRESULTMessageIdentity
}

// PDUSESSIONAUTHENTICATIONRESULTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONAUTHENTICATIONRESULTMessageIdentity) GetMessageType() (messageType uint8) {
	return a.Octet
}

// PDUSESSIONAUTHENTICATIONRESULTMessageIdentity 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSESSIONAUTHENTICATIONRESULTMessageIdentity) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
