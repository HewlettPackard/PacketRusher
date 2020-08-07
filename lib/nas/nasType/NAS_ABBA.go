package nasType

// ABBA 9.11.3.10
// ABBAContents Row, sBit, len = [0, 0], 8 , INF
type ABBA struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewABBA(iei uint8) (aBBA *ABBA) {
	aBBA = &ABBA{}
	aBBA.SetIei(iei)
	return aBBA
}

// ABBA 9.11.3.10
// Iei Row, sBit, len = [], 8, 8
func (a *ABBA) GetIei() (iei uint8) {
	return a.Iei
}

// ABBA 9.11.3.10
// Iei Row, sBit, len = [], 8, 8
func (a *ABBA) SetIei(iei uint8) {
	a.Iei = iei
}

// ABBA 9.11.3.10
// Len Row, sBit, len = [], 8, 8
func (a *ABBA) GetLen() (len uint8) {
	return a.Len
}

// ABBA 9.11.3.10
// Len Row, sBit, len = [], 8, 8
func (a *ABBA) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// ABBA 9.11.3.10
// ABBAContents Row, sBit, len = [0, 0], 8 , INF
func (a *ABBA) GetABBAContents() (aBBAContents []uint8) {
	aBBAContents = make([]uint8, len(a.Buffer))
	copy(aBBAContents, a.Buffer)
	return aBBAContents
}

// ABBA 9.11.3.10
// ABBAContents Row, sBit, len = [0, 0], 8 , INF
func (a *ABBA) SetABBAContents(aBBAContents []uint8) {
	copy(a.Buffer, aBBAContents)
}
