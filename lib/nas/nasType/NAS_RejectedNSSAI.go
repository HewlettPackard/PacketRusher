package nasType

// RejectedNSSAI 9.11.3.46
// RejectedNSSAIContents Row, sBit, len = [0, 0], 0 , INF
type RejectedNSSAI struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewRejectedNSSAI(iei uint8) (rejectedNSSAI *RejectedNSSAI) {
	rejectedNSSAI = &RejectedNSSAI{}
	rejectedNSSAI.SetIei(iei)
	return rejectedNSSAI
}

// RejectedNSSAI 9.11.3.46
// Iei Row, sBit, len = [], 8, 8
func (a *RejectedNSSAI) GetIei() (iei uint8) {
	return a.Iei
}

// RejectedNSSAI 9.11.3.46
// Iei Row, sBit, len = [], 8, 8
func (a *RejectedNSSAI) SetIei(iei uint8) {
	a.Iei = iei
}

// RejectedNSSAI 9.11.3.46
// Len Row, sBit, len = [], 8, 8
func (a *RejectedNSSAI) GetLen() (len uint8) {
	return a.Len
}

// RejectedNSSAI 9.11.3.46
// Len Row, sBit, len = [], 8, 8
func (a *RejectedNSSAI) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// RejectedNSSAI 9.11.3.46
// RejectedNSSAIContents Row, sBit, len = [0, 0], 0 , INF
func (a *RejectedNSSAI) GetRejectedNSSAIContents() (rejectedNSSAIContents []uint8) {
	rejectedNSSAIContents = make([]uint8, len(a.Buffer))
	copy(rejectedNSSAIContents, a.Buffer)
	return rejectedNSSAIContents
}

// RejectedNSSAI 9.11.3.46
// RejectedNSSAIContents Row, sBit, len = [0, 0], 0 , INF
func (a *RejectedNSSAI) SetRejectedNSSAIContents(rejectedNSSAIContents []uint8) {
	copy(a.Buffer, rejectedNSSAIContents)
}
