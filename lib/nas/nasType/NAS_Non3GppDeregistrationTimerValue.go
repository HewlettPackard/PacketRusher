package nasType

// Non3GppDeregistrationTimerValue 9.11.2.4
// GPRSTimer2Value Row, sBit, len = [0, 0], 8 , 8
type Non3GppDeregistrationTimerValue struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewNon3GppDeregistrationTimerValue(iei uint8) (non3GppDeregistrationTimerValue *Non3GppDeregistrationTimerValue) {
	non3GppDeregistrationTimerValue = &Non3GppDeregistrationTimerValue{}
	non3GppDeregistrationTimerValue.SetIei(iei)
	return non3GppDeregistrationTimerValue
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// Iei Row, sBit, len = [], 8, 8
func (a *Non3GppDeregistrationTimerValue) GetIei() (iei uint8) {
	return a.Iei
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// Iei Row, sBit, len = [], 8, 8
func (a *Non3GppDeregistrationTimerValue) SetIei(iei uint8) {
	a.Iei = iei
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// Len Row, sBit, len = [], 8, 8
func (a *Non3GppDeregistrationTimerValue) GetLen() (len uint8) {
	return a.Len
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// Len Row, sBit, len = [], 8, 8
func (a *Non3GppDeregistrationTimerValue) SetLen(len uint8) {
	a.Len = len
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// GPRSTimer2Value Row, sBit, len = [0, 0], 8 , 8
func (a *Non3GppDeregistrationTimerValue) GetGPRSTimer2Value() (gPRSTimer2Value uint8) {
	return a.Octet
}

// Non3GppDeregistrationTimerValue 9.11.2.4
// GPRSTimer2Value Row, sBit, len = [0, 0], 8 , 8
func (a *Non3GppDeregistrationTimerValue) SetGPRSTimer2Value(gPRSTimer2Value uint8) {
	a.Octet = gPRSTimer2Value
}
