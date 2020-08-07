package nasType

// NASMessageContainer 9.11.3.33
// NASMessageContainerContents Row, sBit, len = [0, 0], 8 , INF
type NASMessageContainer struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewNASMessageContainer(iei uint8) (nASMessageContainer *NASMessageContainer) {
	nASMessageContainer = &NASMessageContainer{}
	nASMessageContainer.SetIei(iei)
	return nASMessageContainer
}

// NASMessageContainer 9.11.3.33
// Iei Row, sBit, len = [], 8, 8
func (a *NASMessageContainer) GetIei() (iei uint8) {
	return a.Iei
}

// NASMessageContainer 9.11.3.33
// Iei Row, sBit, len = [], 8, 8
func (a *NASMessageContainer) SetIei(iei uint8) {
	a.Iei = iei
}

// NASMessageContainer 9.11.3.33
// Len Row, sBit, len = [], 8, 16
func (a *NASMessageContainer) GetLen() (len uint16) {
	return a.Len
}

// NASMessageContainer 9.11.3.33
// Len Row, sBit, len = [], 8, 16
func (a *NASMessageContainer) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// NASMessageContainer 9.11.3.33
// NASMessageContainerContents Row, sBit, len = [0, 0], 8 , INF
func (a *NASMessageContainer) GetNASMessageContainerContents() (nASMessageContainerContents []uint8) {
	nASMessageContainerContents = make([]uint8, len(a.Buffer))
	copy(nASMessageContainerContents, a.Buffer)
	return nASMessageContainerContents
}

// NASMessageContainer 9.11.3.33
// NASMessageContainerContents Row, sBit, len = [0, 0], 8 , INF
func (a *NASMessageContainer) SetNASMessageContainerContents(nASMessageContainerContents []uint8) {
	copy(a.Buffer, nASMessageContainerContents)
}
