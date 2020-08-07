package nasType

// MobileIdentity 9.11.3.4
// MobileIdentityContents Row, sBit, len = [0, 0], 8 , INF
type MobileIdentity struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewMobileIdentity(iei uint8) (mobileIdentity *MobileIdentity) {
	mobileIdentity = &MobileIdentity{}
	mobileIdentity.SetIei(iei)
	return mobileIdentity
}

// MobileIdentity 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *MobileIdentity) GetIei() (iei uint8) {
	return a.Iei
}

// MobileIdentity 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *MobileIdentity) SetIei(iei uint8) {
	a.Iei = iei
}

// MobileIdentity 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *MobileIdentity) GetLen() (len uint16) {
	return a.Len
}

// MobileIdentity 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *MobileIdentity) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// MobileIdentity 9.11.3.4
// MobileIdentityContents Row, sBit, len = [0, 0], 8 , INF
func (a *MobileIdentity) GetMobileIdentityContents() (mobileIdentityContents []uint8) {
	mobileIdentityContents = make([]uint8, len(a.Buffer))
	copy(mobileIdentityContents, a.Buffer)
	return mobileIdentityContents
}

// MobileIdentity 9.11.3.4
// MobileIdentityContents Row, sBit, len = [0, 0], 8 , INF
func (a *MobileIdentity) SetMobileIdentityContents(mobileIdentityContents []uint8) {
	copy(a.Buffer, mobileIdentityContents)
}
