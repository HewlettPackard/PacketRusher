package nasType

// SMPDUDNRequestContainer 9.11.4.15
// DNSpecificIdentity Row, sBit, len = [0, 0], 8 , INF
type SMPDUDNRequestContainer struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewSMPDUDNRequestContainer(iei uint8) (sMPDUDNRequestContainer *SMPDUDNRequestContainer) {
	sMPDUDNRequestContainer = &SMPDUDNRequestContainer{}
	sMPDUDNRequestContainer.SetIei(iei)
	return sMPDUDNRequestContainer
}

// SMPDUDNRequestContainer 9.11.4.15
// Iei Row, sBit, len = [], 8, 8
func (a *SMPDUDNRequestContainer) GetIei() (iei uint8) {
	return a.Iei
}

// SMPDUDNRequestContainer 9.11.4.15
// Iei Row, sBit, len = [], 8, 8
func (a *SMPDUDNRequestContainer) SetIei(iei uint8) {
	a.Iei = iei
}

// SMPDUDNRequestContainer 9.11.4.15
// Len Row, sBit, len = [], 8, 8
func (a *SMPDUDNRequestContainer) GetLen() (len uint8) {
	return a.Len
}

// SMPDUDNRequestContainer 9.11.4.15
// Len Row, sBit, len = [], 8, 8
func (a *SMPDUDNRequestContainer) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// SMPDUDNRequestContainer 9.11.4.15
// DNSpecificIdentity Row, sBit, len = [0, 0], 8 , INF
func (a *SMPDUDNRequestContainer) GetDNSpecificIdentity() (dNSpecificIdentity []uint8) {
	dNSpecificIdentity = make([]uint8, len(a.Buffer))
	copy(dNSpecificIdentity, a.Buffer)
	return dNSpecificIdentity
}

// SMPDUDNRequestContainer 9.11.4.15
// DNSpecificIdentity Row, sBit, len = [0, 0], 8 , INF
func (a *SMPDUDNRequestContainer) SetDNSpecificIdentity(dNSpecificIdentity []uint8) {
	copy(a.Buffer, dNSpecificIdentity)
}
