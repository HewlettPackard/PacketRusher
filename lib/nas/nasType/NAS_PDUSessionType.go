package nasType

// PDUSessionType 9.11.4.11
// Iei Row, sBit, len = [0, 0], 8 , 4
// Spare Row, sBit, len = [0, 0], 4 , 1
// PDUSessionTypeValue Row, sBit, len = [0, 0], 3 , 3
type PDUSessionType struct {
	Octet uint8
}

func NewPDUSessionType(iei uint8) (pDUSessionType *PDUSessionType) {
	pDUSessionType = &PDUSessionType{}
	pDUSessionType.SetIei(iei)
	return pDUSessionType
}

// PDUSessionType 9.11.4.11
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *PDUSessionType) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// PDUSessionType 9.11.4.11
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *PDUSessionType) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// PDUSessionType 9.11.4.11
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *PDUSessionType) GetSpare() (spare uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// PDUSessionType 9.11.4.11
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *PDUSessionType) SetSpare(spare uint8) {
	a.Octet = (a.Octet & 247) + ((spare & 1) << 3)
}

// PDUSessionType 9.11.4.11
// PDUSessionTypeValue Row, sBit, len = [0, 0], 3 , 3
func (a *PDUSessionType) GetPDUSessionTypeValue() (pDUSessionTypeValue uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// PDUSessionType 9.11.4.11
// PDUSessionTypeValue Row, sBit, len = [0, 0], 3 , 3
func (a *PDUSessionType) SetPDUSessionTypeValue(pDUSessionTypeValue uint8) {
	a.Octet = (a.Octet & 248) + (pDUSessionTypeValue & 7)
}
