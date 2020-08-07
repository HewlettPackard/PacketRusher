package nasType

// SORTransparentContainer 9.11.3.51
// SORContent Row, sBit, len = [0, 0], 8 , INF
type SORTransparentContainer struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewSORTransparentContainer(iei uint8) (sORTransparentContainer *SORTransparentContainer) {
	sORTransparentContainer = &SORTransparentContainer{}
	sORTransparentContainer.SetIei(iei)
	return sORTransparentContainer
}

// SORTransparentContainer 9.11.3.51
// Iei Row, sBit, len = [], 8, 8
func (a *SORTransparentContainer) GetIei() (iei uint8) {
	return a.Iei
}

// SORTransparentContainer 9.11.3.51
// Iei Row, sBit, len = [], 8, 8
func (a *SORTransparentContainer) SetIei(iei uint8) {
	a.Iei = iei
}

// SORTransparentContainer 9.11.3.51
// Len Row, sBit, len = [], 8, 16
func (a *SORTransparentContainer) GetLen() (len uint16) {
	return a.Len
}

// SORTransparentContainer 9.11.3.51
// Len Row, sBit, len = [], 8, 16
func (a *SORTransparentContainer) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// SORTransparentContainer 9.11.3.51
// SORContent Row, sBit, len = [0, 0], 8 , INF
func (a *SORTransparentContainer) GetSORContent() (sORContent []uint8) {
	sORContent = make([]uint8, len(a.Buffer))
	copy(sORContent, a.Buffer)
	return sORContent
}

// SORTransparentContainer 9.11.3.51
// SORContent Row, sBit, len = [0, 0], 8 , INF
func (a *SORTransparentContainer) SetSORContent(sORContent []uint8) {
	copy(a.Buffer, sORContent)
}
