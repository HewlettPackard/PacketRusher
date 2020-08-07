package nasType

// ServiceAreaList 9.11.3.49
// PartialServiceAreaList Row, sBit, len = [0, 0], 8 , INF
type ServiceAreaList struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewServiceAreaList(iei uint8) (serviceAreaList *ServiceAreaList) {
	serviceAreaList = &ServiceAreaList{}
	serviceAreaList.SetIei(iei)
	return serviceAreaList
}

// ServiceAreaList 9.11.3.49
// Iei Row, sBit, len = [], 8, 8
func (a *ServiceAreaList) GetIei() (iei uint8) {
	return a.Iei
}

// ServiceAreaList 9.11.3.49
// Iei Row, sBit, len = [], 8, 8
func (a *ServiceAreaList) SetIei(iei uint8) {
	a.Iei = iei
}

// ServiceAreaList 9.11.3.49
// Len Row, sBit, len = [], 8, 8
func (a *ServiceAreaList) GetLen() (len uint8) {
	return a.Len
}

// ServiceAreaList 9.11.3.49
// Len Row, sBit, len = [], 8, 8
func (a *ServiceAreaList) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// ServiceAreaList 9.11.3.49
// PartialServiceAreaList Row, sBit, len = [0, 0], 8 , INF
func (a *ServiceAreaList) GetPartialServiceAreaList() (partialServiceAreaList []uint8) {
	partialServiceAreaList = make([]uint8, len(a.Buffer))
	copy(partialServiceAreaList, a.Buffer)
	return partialServiceAreaList
}

// ServiceAreaList 9.11.3.49
// PartialServiceAreaList Row, sBit, len = [0, 0], 8 , INF
func (a *ServiceAreaList) SetPartialServiceAreaList(partialServiceAreaList []uint8) {
	copy(a.Buffer, partialServiceAreaList)
}
