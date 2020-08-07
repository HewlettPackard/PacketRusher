package nasType

// LADNInformation 9.11.3.30
// LADND Row, sBit, len = [0, 0], 8 , INF
type LADNInformation struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewLADNInformation(iei uint8) (lADNInformation *LADNInformation) {
	lADNInformation = &LADNInformation{}
	lADNInformation.SetIei(iei)
	return lADNInformation
}

// LADNInformation 9.11.3.30
// Iei Row, sBit, len = [], 8, 8
func (a *LADNInformation) GetIei() (iei uint8) {
	return a.Iei
}

// LADNInformation 9.11.3.30
// Iei Row, sBit, len = [], 8, 8
func (a *LADNInformation) SetIei(iei uint8) {
	a.Iei = iei
}

// LADNInformation 9.11.3.30
// Len Row, sBit, len = [], 8, 16
func (a *LADNInformation) GetLen() (len uint16) {
	return a.Len
}

// LADNInformation 9.11.3.30
// Len Row, sBit, len = [], 8, 16
func (a *LADNInformation) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// LADNInformation 9.11.3.30
// LADND Row, sBit, len = [0, 0], 8 , INF
func (a *LADNInformation) GetLADND() (lADND []uint8) {
	lADND = make([]uint8, len(a.Buffer))
	copy(lADND, a.Buffer)
	return lADND
}

// LADNInformation 9.11.3.30
// LADND Row, sBit, len = [0, 0], 8 , INF
func (a *LADNInformation) SetLADND(lADND []uint8) {
	copy(a.Buffer, lADND)
}
