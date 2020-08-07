package nasType

// SNSSAI 9.11.2.8
// SST Row, sBit, len = [0, 0], 8 , 8
// SD Row, sBit, len = [1, 3], 8 , 24
// MappedHPLMNSST Row, sBit, len = [4, 4], 8 , 8
// MappedHPLMNSD Row, sBit, len = [5, 7], 8 , 24
type SNSSAI struct {
	Iei   uint8
	Len   uint8
	Octet [8]uint8
}

func NewSNSSAI(iei uint8) (sNSSAI *SNSSAI) {
	sNSSAI = &SNSSAI{}
	sNSSAI.SetIei(iei)
	return sNSSAI
}

// SNSSAI 9.11.2.8
// Iei Row, sBit, len = [], 8, 8
func (a *SNSSAI) GetIei() (iei uint8) {
	return a.Iei
}

// SNSSAI 9.11.2.8
// Iei Row, sBit, len = [], 8, 8
func (a *SNSSAI) SetIei(iei uint8) {
	a.Iei = iei
}

// SNSSAI 9.11.2.8
// Len Row, sBit, len = [], 8, 8
func (a *SNSSAI) GetLen() (len uint8) {
	return a.Len
}

// SNSSAI 9.11.2.8
// Len Row, sBit, len = [], 8, 8
func (a *SNSSAI) SetLen(len uint8) {
	a.Len = len
}

// SNSSAI 9.11.2.8
// SST Row, sBit, len = [0, 0], 8 , 8
func (a *SNSSAI) GetSST() (sST uint8) {
	return a.Octet[0]
}

// SNSSAI 9.11.2.8
// SST Row, sBit, len = [0, 0], 8 , 8
func (a *SNSSAI) SetSST(sST uint8) {
	a.Octet[0] = sST
}

// SNSSAI 9.11.2.8
// SD Row, sBit, len = [1, 3], 8 , 24
func (a *SNSSAI) GetSD() (sD [3]uint8) {
	copy(sD[:], a.Octet[1:4])
	return sD
}

// SNSSAI 9.11.2.8
// SD Row, sBit, len = [1, 3], 8 , 24
func (a *SNSSAI) SetSD(sD [3]uint8) {
	copy(a.Octet[1:4], sD[:])
}

// SNSSAI 9.11.2.8
// MappedHPLMNSST Row, sBit, len = [4, 4], 8 , 8
func (a *SNSSAI) GetMappedHPLMNSST() (mappedHPLMNSST uint8) {
	return a.Octet[4]
}

// SNSSAI 9.11.2.8
// MappedHPLMNSST Row, sBit, len = [4, 4], 8 , 8
func (a *SNSSAI) SetMappedHPLMNSST(mappedHPLMNSST uint8) {
	a.Octet[4] = mappedHPLMNSST
}

// SNSSAI 9.11.2.8
// MappedHPLMNSD Row, sBit, len = [5, 7], 8 , 24
func (a *SNSSAI) GetMappedHPLMNSD() (mappedHPLMNSD [3]uint8) {
	copy(mappedHPLMNSD[:], a.Octet[5:8])
	return mappedHPLMNSD
}

// SNSSAI 9.11.2.8
// MappedHPLMNSD Row, sBit, len = [5, 7], 8 , 24
func (a *SNSSAI) SetMappedHPLMNSD(mappedHPLMNSD [3]uint8) {
	copy(a.Octet[5:8], mappedHPLMNSD[:])
}
