package nasType

// PayloadContainer 9.11.3.39
// PayloadContainerContents Row, sBit, len = [0, 0], 8 , INF
type PayloadContainer struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewPayloadContainer(iei uint8) (payloadContainer *PayloadContainer) {
	payloadContainer = &PayloadContainer{}
	payloadContainer.SetIei(iei)
	return payloadContainer
}

// PayloadContainer 9.11.3.39
// Iei Row, sBit, len = [], 8, 8
func (a *PayloadContainer) GetIei() (iei uint8) {
	return a.Iei
}

// PayloadContainer 9.11.3.39
// Iei Row, sBit, len = [], 8, 8
func (a *PayloadContainer) SetIei(iei uint8) {
	a.Iei = iei
}

// PayloadContainer 9.11.3.39
// Len Row, sBit, len = [], 8, 16
func (a *PayloadContainer) GetLen() (len uint16) {
	return a.Len
}

// PayloadContainer 9.11.3.39
// Len Row, sBit, len = [], 8, 16
func (a *PayloadContainer) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// PayloadContainer 9.11.3.39
// PayloadContainerContents Row, sBit, len = [0, 0], 8 , INF
func (a *PayloadContainer) GetPayloadContainerContents() (payloadContainerContents []uint8) {
	payloadContainerContents = make([]uint8, len(a.Buffer))
	copy(payloadContainerContents, a.Buffer)
	return payloadContainerContents
}

// PayloadContainer 9.11.3.39
// PayloadContainerContents Row, sBit, len = [0, 0], 8 , INF
func (a *PayloadContainer) SetPayloadContainerContents(payloadContainerContents []uint8) {
	copy(a.Buffer, payloadContainerContents)
}
