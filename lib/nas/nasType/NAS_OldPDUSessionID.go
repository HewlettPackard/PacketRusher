package nasType

// OldPDUSessionID 9.11.3.41
// OldPDUSessionID Row, sBit, len = [0, 0], 8 , 8
type OldPDUSessionID struct {
	Iei   uint8
	Octet uint8
}

func NewOldPDUSessionID(iei uint8) (oldPDUSessionID *OldPDUSessionID) {
	oldPDUSessionID = &OldPDUSessionID{}
	oldPDUSessionID.SetIei(iei)
	return oldPDUSessionID
}

// OldPDUSessionID 9.11.3.41
// Iei Row, sBit, len = [], 8, 8
func (a *OldPDUSessionID) GetIei() (iei uint8) {
	return a.Iei
}

// OldPDUSessionID 9.11.3.41
// Iei Row, sBit, len = [], 8, 8
func (a *OldPDUSessionID) SetIei(iei uint8) {
	a.Iei = iei
}

// OldPDUSessionID 9.11.3.41
// OldPDUSessionID Row, sBit, len = [0, 0], 8 , 8
func (a *OldPDUSessionID) GetOldPDUSessionID() (oldPDUSessionID uint8) {
	return a.Octet
}

// OldPDUSessionID 9.11.3.41
// OldPDUSessionID Row, sBit, len = [0, 0], 8 , 8
func (a *OldPDUSessionID) SetOldPDUSessionID(oldPDUSessionID uint8) {
	a.Octet = oldPDUSessionID
}
