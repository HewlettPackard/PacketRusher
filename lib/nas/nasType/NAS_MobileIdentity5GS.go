package nasType

// MobileIdentity5GS 9.11.3.4
// MobileIdentity5GSContents Row, sBit, len = [0, 0], 8 , INF
type MobileIdentity5GS struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewMobileIdentity5GS(iei uint8) (mobileIdentity5GS *MobileIdentity5GS) {
	mobileIdentity5GS = &MobileIdentity5GS{}
	mobileIdentity5GS.SetIei(iei)
	return mobileIdentity5GS
}

// MobileIdentity5GS 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *MobileIdentity5GS) GetIei() (iei uint8) {
	return a.Iei
}

// MobileIdentity5GS 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *MobileIdentity5GS) SetIei(iei uint8) {
	a.Iei = iei
}

// MobileIdentity5GS 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *MobileIdentity5GS) GetLen() (len uint16) {
	return a.Len
}

// MobileIdentity5GS 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *MobileIdentity5GS) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// MobileIdentity5GS 9.11.3.4
// MobileIdentity5GSContents Row, sBit, len = [0, 0], 8 , INF
func (a *MobileIdentity5GS) GetMobileIdentity5GSContents() (mobileIdentity5GSContents []uint8) {
	mobileIdentity5GSContents = make([]uint8, len(a.Buffer))
	copy(mobileIdentity5GSContents, a.Buffer)
	return mobileIdentity5GSContents
}

// MobileIdentity5GS 9.11.3.4
// MobileIdentity5GSContents Row, sBit, len = [0, 0], 8 , INF
func (a *MobileIdentity5GS) SetMobileIdentity5GSContents(mobileIdentity5GSContents []uint8) {
	copy(a.Buffer, mobileIdentity5GSContents)
}
