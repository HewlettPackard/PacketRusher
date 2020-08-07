package nasType

// MICOIndication 9.11.3.31
// Iei Row, sBit, len = [0, 0], 8 , 4
// RAAI Row, sBit, len = [0, 0], 1 , 1
type MICOIndication struct {
	Octet uint8
}

func NewMICOIndication(iei uint8) (mICOIndication *MICOIndication) {
	mICOIndication = &MICOIndication{}
	mICOIndication.SetIei(iei)
	return mICOIndication
}

// MICOIndication 9.11.3.31
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *MICOIndication) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// MICOIndication 9.11.3.31
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *MICOIndication) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// MICOIndication 9.11.3.31
// RAAI Row, sBit, len = [0, 0], 1 , 1
func (a *MICOIndication) GetRAAI() (rAAI uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// MICOIndication 9.11.3.31
// RAAI Row, sBit, len = [0, 0], 1 , 1
func (a *MICOIndication) SetRAAI(rAAI uint8) {
	a.Octet = (a.Octet & 254) + (rAAI & 1)
}
