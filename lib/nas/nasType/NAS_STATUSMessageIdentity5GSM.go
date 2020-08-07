package nasType

// STATUSMessageIdentity5GSM 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
type STATUSMessageIdentity5GSM struct {
	Octet uint8
}

func NewSTATUSMessageIdentity5GSM() (sTATUSMessageIdentity5GSM *STATUSMessageIdentity5GSM) {
	sTATUSMessageIdentity5GSM = &STATUSMessageIdentity5GSM{}
	return sTATUSMessageIdentity5GSM
}

// STATUSMessageIdentity5GSM 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *STATUSMessageIdentity5GSM) GetMessageType() (messageType uint8) {
	return a.Octet
}

// STATUSMessageIdentity5GSM 9.7
// MessageType Row, sBit, len = [0, 0], 8 , 8
func (a *STATUSMessageIdentity5GSM) SetMessageType(messageType uint8) {
	a.Octet = messageType
}
