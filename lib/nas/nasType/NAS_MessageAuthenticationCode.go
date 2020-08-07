package nasType

// MessageAuthenticationCode MAC 9.8
// MAC Row, sBit, len = [0, 3], 8 , 32
type MessageAuthenticationCode struct {
	Octet [4]uint8
}

func NewMessageAuthenticationCode() (messageAuthenticationCode *MessageAuthenticationCode) {
	messageAuthenticationCode = &MessageAuthenticationCode{}
	return messageAuthenticationCode
}

// MessageAuthenticationCode MAC 9.8
// MAC Row, sBit, len = [0, 3], 8 , 32
func (a *MessageAuthenticationCode) GetMAC() (mAC [4]uint8) {
	copy(mAC[:], a.Octet[0:4])
	return mAC
}

// MessageAuthenticationCode MAC 9.8
// MAC Row, sBit, len = [0, 3], 8 , 32
func (a *MessageAuthenticationCode) SetMAC(mAC [4]uint8) {
	copy(a.Octet[0:4], mAC[:])
}
