package nasType

// TAIList 9.11.3.9
// PartialTrackingAreaIdentityList Row, sBit, len = [0, 0], 8 , INF
type TAIList struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewTAIList(iei uint8) (tAIList *TAIList) {
	tAIList = &TAIList{}
	tAIList.SetIei(iei)
	return tAIList
}

// TAIList 9.11.3.9
// Iei Row, sBit, len = [], 8, 8
func (a *TAIList) GetIei() (iei uint8) {
	return a.Iei
}

// TAIList 9.11.3.9
// Iei Row, sBit, len = [], 8, 8
func (a *TAIList) SetIei(iei uint8) {
	a.Iei = iei
}

// TAIList 9.11.3.9
// Len Row, sBit, len = [], 8, 8
func (a *TAIList) GetLen() (len uint8) {
	return a.Len
}

// TAIList 9.11.3.9
// Len Row, sBit, len = [], 8, 8
func (a *TAIList) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// TAIList 9.11.3.9
// PartialTrackingAreaIdentityList Row, sBit, len = [0, 0], 8 , INF
func (a *TAIList) GetPartialTrackingAreaIdentityList() (partialTrackingAreaIdentityList []uint8) {
	partialTrackingAreaIdentityList = make([]uint8, len(a.Buffer))
	copy(partialTrackingAreaIdentityList, a.Buffer)
	return partialTrackingAreaIdentityList
}

// TAIList 9.11.3.9
// PartialTrackingAreaIdentityList Row, sBit, len = [0, 0], 8 , INF
func (a *TAIList) SetPartialTrackingAreaIdentityList(partialTrackingAreaIdentityList []uint8) {
	copy(a.Buffer, partialTrackingAreaIdentityList)
}
