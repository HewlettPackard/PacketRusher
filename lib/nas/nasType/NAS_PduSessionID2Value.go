package nasType

// PduSessionID2Value 9.11.3.41
// PduSessionID2Value Row, sBit, len = [0, 0], 8 , 8
type PduSessionID2Value struct {
	Iei   uint8
	Octet uint8
}

func NewPduSessionID2Value(iei uint8) (pduSessionID2Value *PduSessionID2Value) {
	pduSessionID2Value = &PduSessionID2Value{}
	pduSessionID2Value.SetIei(iei)
	return pduSessionID2Value
}

// PduSessionID2Value 9.11.3.41
// Iei Row, sBit, len = [], 8, 8
func (a *PduSessionID2Value) GetIei() (iei uint8) {
	return a.Iei
}

// PduSessionID2Value 9.11.3.41
// Iei Row, sBit, len = [], 8, 8
func (a *PduSessionID2Value) SetIei(iei uint8) {
	a.Iei = iei
}

// PduSessionID2Value 9.11.3.41
// PduSessionID2Value Row, sBit, len = [0, 0], 8 , 8
func (a *PduSessionID2Value) GetPduSessionID2Value() (pduSessionID2Value uint8) {
	return a.Octet
}

// PduSessionID2Value 9.11.3.41
// PduSessionID2Value Row, sBit, len = [0, 0], 8 , 8
func (a *PduSessionID2Value) SetPduSessionID2Value(pduSessionID2Value uint8) {
	a.Octet = pduSessionID2Value
}
