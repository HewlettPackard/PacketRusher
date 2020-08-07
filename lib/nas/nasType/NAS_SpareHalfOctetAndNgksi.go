package nasType

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
// TSC Row, sBit, len = [0, 0], 4 , 1
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
type SpareHalfOctetAndNgksi struct {
	Octet uint8
}

func NewSpareHalfOctetAndNgksi() (spareHalfOctetAndNgksi *SpareHalfOctetAndNgksi) {
	spareHalfOctetAndNgksi = &SpareHalfOctetAndNgksi{}
	return spareHalfOctetAndNgksi
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
func (a *SpareHalfOctetAndNgksi) GetSpareHalfOctet() (spareHalfOctet uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
func (a *SpareHalfOctetAndNgksi) SetSpareHalfOctet(spareHalfOctet uint8) {
	a.Octet = (a.Octet & 15) + ((spareHalfOctet & 15) << 4)
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// TSC Row, sBit, len = [0, 0], 4 , 1
func (a *SpareHalfOctetAndNgksi) GetTSC() (tSC uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// TSC Row, sBit, len = [0, 0], 4 , 1
func (a *SpareHalfOctetAndNgksi) SetTSC(tSC uint8) {
	a.Octet = (a.Octet & 247) + ((tSC & 1) << 3)
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *SpareHalfOctetAndNgksi) GetNasKeySetIdentifiler() (nasKeySetIdentifiler uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// SpareHalfOctetAndNgksi 9.11.3.32 9.5
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *SpareHalfOctetAndNgksi) SetNasKeySetIdentifiler(nasKeySetIdentifiler uint8) {
	a.Octet = (a.Octet & 248) + (nasKeySetIdentifiler & 7)
}
