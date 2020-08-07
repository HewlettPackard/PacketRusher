package nasType

// SpareHalfOctetAndSecurityHeaderType 9.3 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
// SecurityHeaderType Row, sBit, len = [0, 0], 4 , 4
type SpareHalfOctetAndSecurityHeaderType struct {
	Octet uint8
}

func NewSpareHalfOctetAndSecurityHeaderType() (spareHalfOctetAndSecurityHeaderType *SpareHalfOctetAndSecurityHeaderType) {
	spareHalfOctetAndSecurityHeaderType = &SpareHalfOctetAndSecurityHeaderType{}
	return spareHalfOctetAndSecurityHeaderType
}

// SpareHalfOctetAndSecurityHeaderType 9.3 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
func (a *SpareHalfOctetAndSecurityHeaderType) GetSpareHalfOctet() (spareHalfOctet uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// SpareHalfOctetAndSecurityHeaderType 9.3 9.5
// SpareHalfOctet Row, sBit, len = [0, 0], 8 , 4
func (a *SpareHalfOctetAndSecurityHeaderType) SetSpareHalfOctet(spareHalfOctet uint8) {
	a.Octet = (a.Octet & 15) + ((spareHalfOctet & 15) << 4)
}

// SpareHalfOctetAndSecurityHeaderType 9.3 9.5
// SecurityHeaderType Row, sBit, len = [0, 0], 4 , 4
func (a *SpareHalfOctetAndSecurityHeaderType) GetSecurityHeaderType() (securityHeaderType uint8) {
	return a.Octet & GetBitMask(4, 0)
}

// SpareHalfOctetAndSecurityHeaderType 9.3 9.5
// SecurityHeaderType Row, sBit, len = [0, 0], 4 , 4
func (a *SpareHalfOctetAndSecurityHeaderType) SetSecurityHeaderType(securityHeaderType uint8) {
	a.Octet = (a.Octet & 240) + (securityHeaderType & 15)
}
