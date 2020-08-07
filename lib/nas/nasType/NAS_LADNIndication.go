package nasType

// LADNIndication 9.11.3.29
// LADNDNNValue Row, sBit, len = [0, 0], 8 , INF
type LADNIndication struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewLADNIndication(iei uint8) (lADNIndication *LADNIndication) {
	lADNIndication = &LADNIndication{}
	lADNIndication.SetIei(iei)
	return lADNIndication
}

// LADNIndication 9.11.3.29
// Iei Row, sBit, len = [], 8, 8
func (a *LADNIndication) GetIei() (iei uint8) {
	return a.Iei
}

// LADNIndication 9.11.3.29
// Iei Row, sBit, len = [], 8, 8
func (a *LADNIndication) SetIei(iei uint8) {
	a.Iei = iei
}

// LADNIndication 9.11.3.29
// Len Row, sBit, len = [], 8, 16
func (a *LADNIndication) GetLen() (len uint16) {
	return a.Len
}

// LADNIndication 9.11.3.29
// Len Row, sBit, len = [], 8, 16
func (a *LADNIndication) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// LADNIndication 9.11.3.29
// LADNDNNValue Row, sBit, len = [0, 0], 8 , INF
func (a *LADNIndication) GetLADNDNNValue() (lADNDNNValue []uint8) {
	lADNDNNValue = make([]uint8, len(a.Buffer))
	copy(lADNDNNValue, a.Buffer)
	return lADNDNNValue
}

// LADNIndication 9.11.3.29
// LADNDNNValue Row, sBit, len = [0, 0], 8 , INF
func (a *LADNIndication) SetLADNDNNValue(lADNDNNValue []uint8) {
	copy(a.Buffer, lADNDNNValue)
}
