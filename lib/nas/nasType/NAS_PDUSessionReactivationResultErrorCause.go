package nasType

// PDUSessionReactivationResultErrorCause 9.11.3.43
// PDUSessionIDAndCauseValue Row, sBit, len = [0, 0], 8 , INF
type PDUSessionReactivationResultErrorCause struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewPDUSessionReactivationResultErrorCause(iei uint8) (pDUSessionReactivationResultErrorCause *PDUSessionReactivationResultErrorCause) {
	pDUSessionReactivationResultErrorCause = &PDUSessionReactivationResultErrorCause{}
	pDUSessionReactivationResultErrorCause.SetIei(iei)
	return pDUSessionReactivationResultErrorCause
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// Iei Row, sBit, len = [], 8, 8
func (a *PDUSessionReactivationResultErrorCause) GetIei() (iei uint8) {
	return a.Iei
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// Iei Row, sBit, len = [], 8, 8
func (a *PDUSessionReactivationResultErrorCause) SetIei(iei uint8) {
	a.Iei = iei
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// Len Row, sBit, len = [], 8, 16
func (a *PDUSessionReactivationResultErrorCause) GetLen() (len uint16) {
	return a.Len
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// Len Row, sBit, len = [], 8, 16
func (a *PDUSessionReactivationResultErrorCause) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// PDUSessionIDAndCauseValue Row, sBit, len = [0, 0], 8 , INF
func (a *PDUSessionReactivationResultErrorCause) GetPDUSessionIDAndCauseValue() (pDUSessionIDAndCauseValue []uint8) {
	pDUSessionIDAndCauseValue = make([]uint8, len(a.Buffer))
	copy(pDUSessionIDAndCauseValue, a.Buffer)
	return pDUSessionIDAndCauseValue
}

// PDUSessionReactivationResultErrorCause 9.11.3.43
// PDUSessionIDAndCauseValue Row, sBit, len = [0, 0], 8 , INF
func (a *PDUSessionReactivationResultErrorCause) SetPDUSessionIDAndCauseValue(pDUSessionIDAndCauseValue []uint8) {
	copy(a.Buffer, pDUSessionIDAndCauseValue)
}
