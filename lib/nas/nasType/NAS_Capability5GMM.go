package nasType

// Capability5GMM 9.11.3.1
// LPP  Row, sBit, len = [0, 0], 3 , 1
// HOAttach Row, sBit, len = [0, 0], 2 , 1
// S1Mode Row, sBit, len = [0, 0], 1 , 1
// Spare Row, sBit, len = [1, 12], 8 , 96
type Capability5GMM struct {
	Iei   uint8
	Len   uint8
	Octet [13]uint8
}

func NewCapability5GMM(iei uint8) (capability5GMM *Capability5GMM) {
	capability5GMM = &Capability5GMM{}
	capability5GMM.SetIei(iei)
	return capability5GMM
}

// Capability5GMM 9.11.3.1
// Iei Row, sBit, len = [], 8, 8
func (a *Capability5GMM) GetIei() (iei uint8) {
	return a.Iei
}

// Capability5GMM 9.11.3.1
// Iei Row, sBit, len = [], 8, 8
func (a *Capability5GMM) SetIei(iei uint8) {
	a.Iei = iei
}

// Capability5GMM 9.11.3.1
// Len Row, sBit, len = [], 8, 8
func (a *Capability5GMM) GetLen() (len uint8) {
	return a.Len
}

// Capability5GMM 9.11.3.1
// Len Row, sBit, len = [], 8, 8
func (a *Capability5GMM) SetLen(len uint8) {
	a.Len = len
}

// Capability5GMM 9.11.3.1
// LPP Row, sBit, len = [0, 0], 3 , 1
func (a *Capability5GMM) GetLPP() (lPP uint8) {
	return a.Octet[0] & GetBitMask(3, 2) >> (2)
}

// Capability5GMM 9.11.3.1
// LPP Row, sBit, len = [0, 0], 3 , 1
func (a *Capability5GMM) SetLPP(lPP uint8) {
	a.Octet[0] = (a.Octet[0] & 251) + ((lPP & 1) << 2)
}

// Capability5GMM 9.11.3.1
// HOAttach Row, sBit, len = [0, 0], 2 , 1
func (a *Capability5GMM) GetHOAttach() (hOAttach uint8) {
	return a.Octet[0] & GetBitMask(2, 1) >> (1)
}

// Capability5GMM 9.11.3.1
// HOAttach Row, sBit, len = [0, 0], 2 , 1
func (a *Capability5GMM) SetHOAttach(hOAttach uint8) {
	a.Octet[0] = (a.Octet[0] & 253) + ((hOAttach & 1) << 1)
}

// Capability5GMM 9.11.3.1
// S1Mode Row, sBit, len = [0, 0], 1 , 1
func (a *Capability5GMM) GetS1Mode() (s1Mode uint8) {
	return a.Octet[0] & GetBitMask(1, 0)
}

// Capability5GMM 9.11.3.1
// S1Mode Row, sBit, len = [0, 0], 1 , 1
func (a *Capability5GMM) SetS1Mode(s1Mode uint8) {
	a.Octet[0] = (a.Octet[0] & 254) + (s1Mode & 1)
}

// Capability5GMM 9.11.3.1
// Spare Row, sBit, len = [1, 12], 8 , 96
func (a *Capability5GMM) GetSpare() (spare [12]uint8) {
	copy(spare[:], a.Octet[1:13])
	return spare
}

// Capability5GMM 9.11.3.1
// Spare Row, sBit, len = [1, 12], 8 , 96
func (a *Capability5GMM) SetSpare(spare [12]uint8) {
	copy(a.Octet[1:13], spare[:])
}
