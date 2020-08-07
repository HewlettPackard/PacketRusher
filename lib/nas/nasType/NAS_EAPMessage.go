package nasType

// EAPMessage 9.11.2.2
// EAPMessage Row, sBit, len = [0, 0], 8 , INF
type EAPMessage struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewEAPMessage(iei uint8) (eAPMessage *EAPMessage) {
	eAPMessage = &EAPMessage{}
	eAPMessage.SetIei(iei)
	return eAPMessage
}

// EAPMessage 9.11.2.2
// Iei Row, sBit, len = [], 8, 8
func (a *EAPMessage) GetIei() (iei uint8) {
	return a.Iei
}

// EAPMessage 9.11.2.2
// Iei Row, sBit, len = [], 8, 8
func (a *EAPMessage) SetIei(iei uint8) {
	a.Iei = iei
}

// EAPMessage 9.11.2.2
// Len Row, sBit, len = [], 8, 16
func (a *EAPMessage) GetLen() (len uint16) {
	return a.Len
}

// EAPMessage 9.11.2.2
// Len Row, sBit, len = [], 8, 16
func (a *EAPMessage) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// EAPMessage 9.11.2.2
// EAPMessage Row, sBit, len = [0, 0], 8 , INF
func (a *EAPMessage) GetEAPMessage() (eAPMessage []uint8) {
	eAPMessage = make([]uint8, len(a.Buffer))
	copy(eAPMessage, a.Buffer)
	return eAPMessage
}

// EAPMessage 9.11.2.2
// EAPMessage Row, sBit, len = [0, 0], 8 , INF
func (a *EAPMessage) SetEAPMessage(eAPMessage []uint8) {
	copy(a.Buffer, eAPMessage)
}
