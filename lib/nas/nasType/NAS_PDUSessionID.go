package nasType

// PDUSessionID 9.4
// PDUSessionID Row, sBit, len = [0, 0], 8 , 8
type PDUSessionID struct {
	Octet uint8
}

func NewPDUSessionID() (pDUSessionID *PDUSessionID) {
	pDUSessionID = &PDUSessionID{}
	return pDUSessionID
}

// PDUSessionID 9.4
// PDUSessionID Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSessionID) GetPDUSessionID() (pDUSessionID uint8) {
	return a.Octet
}

// PDUSessionID 9.4
// PDUSessionID Row, sBit, len = [0, 0], 8 , 8
func (a *PDUSessionID) SetPDUSessionID(pDUSessionID uint8) {
	a.Octet = pDUSessionID
}
