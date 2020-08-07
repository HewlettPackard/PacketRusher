package nasType

// EmergencyNumberList 9.11.3.23
// Lengthof1EmergencyNumberInformation Row, sBit, len = [0, 0], 8 , 8
// EmergencyServiceCategoryValue Row, sBit, len = [1, 1], 5 , 5
// EmergencyInformation Row, sBit, len = [0, 0], 8 , INF
type EmergencyNumberList struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewEmergencyNumberList(iei uint8) (emergencyNumberList *EmergencyNumberList) {
	emergencyNumberList = &EmergencyNumberList{}
	emergencyNumberList.SetIei(iei)
	return emergencyNumberList
}

// EmergencyNumberList 9.11.3.23
// Iei Row, sBit, len = [], 8, 8
func (a *EmergencyNumberList) GetIei() (iei uint8) {
	return a.Iei
}

// EmergencyNumberList 9.11.3.23
// Iei Row, sBit, len = [], 8, 8
func (a *EmergencyNumberList) SetIei(iei uint8) {
	a.Iei = iei
}

// EmergencyNumberList 9.11.3.23
// Len Row, sBit, len = [], 8, 8
func (a *EmergencyNumberList) GetLen() (len uint8) {
	return a.Len
}

// EmergencyNumberList 9.11.3.23
// Len Row, sBit, len = [], 8, 8
func (a *EmergencyNumberList) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// EmergencyNumberList 9.11.3.23
// Lengthof1EmergencyNumberInformation Row, sBit, len = [0, 0], 8 , 8
func (a *EmergencyNumberList) GetLengthof1EmergencyNumberInformation() (lengthof1EmergencyNumberInformation uint8) {
	return a.Buffer[0]
}

// EmergencyNumberList 9.11.3.23
// Lengthof1EmergencyNumberInformation Row, sBit, len = [0, 0], 8 , 8
func (a *EmergencyNumberList) SetLengthof1EmergencyNumberInformation(lengthof1EmergencyNumberInformation uint8) {
	a.Buffer[0] = lengthof1EmergencyNumberInformation
}

// EmergencyNumberList 9.11.3.23
// EmergencyServiceCategoryValue Row, sBit, len = [1, 1], 5 , 5
func (a *EmergencyNumberList) GetEmergencyServiceCategoryValue() (emergencyServiceCategoryValue uint8) {
	return a.Buffer[1] & GetBitMask(5, 0)
}

// EmergencyNumberList 9.11.3.23
// EmergencyServiceCategoryValue Row, sBit, len = [1, 1], 5 , 5
func (a *EmergencyNumberList) SetEmergencyServiceCategoryValue(emergencyServiceCategoryValue uint8) {
	a.Buffer[1] = (a.Buffer[1] & 224) + (emergencyServiceCategoryValue & 31)
}

// EmergencyNumberList 9.11.3.23
// EmergencyInformation Row, sBit, len = [0, 0], 8 , INF
func (a *EmergencyNumberList) GetEmergencyInformation() (emergencyInformation []uint8) {
	emergencyInformation = make([]uint8, len(a.Buffer))
	copy(emergencyInformation, a.Buffer)
	return emergencyInformation
}

// EmergencyNumberList 9.11.3.23
// EmergencyInformation Row, sBit, len = [0, 0], 8 , INF
func (a *EmergencyNumberList) SetEmergencyInformation(emergencyInformation []uint8) {
	copy(a.Buffer, emergencyInformation)
}
