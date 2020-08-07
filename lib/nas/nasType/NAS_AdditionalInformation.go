package nasType

// AdditionalInformation 9.11.2.1
// AdditionalInformationValue Row, sBit, len = [0, 0], 8 , INF
type AdditionalInformation struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewAdditionalInformation(iei uint8) (additionalInformation *AdditionalInformation) {
	additionalInformation = &AdditionalInformation{}
	additionalInformation.SetIei(iei)
	return additionalInformation
}

// AdditionalInformation 9.11.2.1
// Iei Row, sBit, len = [], 8, 8
func (a *AdditionalInformation) GetIei() (iei uint8) {
	return a.Iei
}

// AdditionalInformation 9.11.2.1
// Iei Row, sBit, len = [], 8, 8
func (a *AdditionalInformation) SetIei(iei uint8) {
	a.Iei = iei
}

// AdditionalInformation 9.11.2.1
// Len Row, sBit, len = [], 8, 8
func (a *AdditionalInformation) GetLen() (len uint8) {
	return a.Len
}

// AdditionalInformation 9.11.2.1
// Len Row, sBit, len = [], 8, 8
func (a *AdditionalInformation) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// AdditionalInformation 9.11.2.1
// AdditionalInformationValue Row, sBit, len = [0, 0], 8 , INF
func (a *AdditionalInformation) GetAdditionalInformationValue() (additionalInformationValue []uint8) {
	additionalInformationValue = make([]uint8, len(a.Buffer))
	copy(additionalInformationValue, a.Buffer)
	return additionalInformationValue
}

// AdditionalInformation 9.11.2.1
// AdditionalInformationValue Row, sBit, len = [0, 0], 8 , INF
func (a *AdditionalInformation) SetAdditionalInformationValue(additionalInformationValue []uint8) {
	copy(a.Buffer, additionalInformationValue)
}
