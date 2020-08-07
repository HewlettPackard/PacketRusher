package nasType

// EPSNASMessageContainer 9.11.3.24
// EPANASMessageContainer Row, sBit, len = [0, 0], 8 , INF
type EPSNASMessageContainer struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewEPSNASMessageContainer(iei uint8) (ePSNASMessageContainer *EPSNASMessageContainer) {
	ePSNASMessageContainer = &EPSNASMessageContainer{}
	ePSNASMessageContainer.SetIei(iei)
	return ePSNASMessageContainer
}

// EPSNASMessageContainer 9.11.3.24
// Iei Row, sBit, len = [], 8, 8
func (a *EPSNASMessageContainer) GetIei() (iei uint8) {
	return a.Iei
}

// EPSNASMessageContainer 9.11.3.24
// Iei Row, sBit, len = [], 8, 8
func (a *EPSNASMessageContainer) SetIei(iei uint8) {
	a.Iei = iei
}

// EPSNASMessageContainer 9.11.3.24
// Len Row, sBit, len = [], 8, 16
func (a *EPSNASMessageContainer) GetLen() (len uint16) {
	return a.Len
}

// EPSNASMessageContainer 9.11.3.24
// Len Row, sBit, len = [], 8, 16
func (a *EPSNASMessageContainer) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// EPSNASMessageContainer 9.11.3.24
// EPANASMessageContainer Row, sBit, len = [0, 0], 8 , INF
func (a *EPSNASMessageContainer) GetEPANASMessageContainer() (ePANASMessageContainer []uint8) {
	ePANASMessageContainer = make([]uint8, len(a.Buffer))
	copy(ePANASMessageContainer, a.Buffer)
	return ePANASMessageContainer
}

// EPSNASMessageContainer 9.11.3.24
// EPANASMessageContainer Row, sBit, len = [0, 0], 8 , INF
func (a *EPSNASMessageContainer) SetEPANASMessageContainer(ePANASMessageContainer []uint8) {
	copy(a.Buffer, ePANASMessageContainer)
}
