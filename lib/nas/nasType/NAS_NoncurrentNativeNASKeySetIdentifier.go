package nasType

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// Iei  Row, sBit, len = [0, 0], 8 , 4
// Tsc  Row, sBit, len = [0, 0], 4 , 1
// NasKeySetIdentifiler  Row, sBit, len = [0, 0], 3 , 3
type NoncurrentNativeNASKeySetIdentifier struct {
	Octet uint8
}

func NewNoncurrentNativeNASKeySetIdentifier(iei uint8) (noncurrentNativeNASKeySetIdentifier *NoncurrentNativeNASKeySetIdentifier) {
	noncurrentNativeNASKeySetIdentifier = &NoncurrentNativeNASKeySetIdentifier{}
	noncurrentNativeNASKeySetIdentifier.SetIei(iei)
	return noncurrentNativeNASKeySetIdentifier
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NoncurrentNativeNASKeySetIdentifier) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NoncurrentNativeNASKeySetIdentifier) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// Tsc Row, sBit, len = [0, 0], 4 , 1
func (a *NoncurrentNativeNASKeySetIdentifier) GetTsc() (tsc uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// Tsc Row, sBit, len = [0, 0], 4 , 1
func (a *NoncurrentNativeNASKeySetIdentifier) SetTsc(tsc uint8) {
	a.Octet = (a.Octet & 247) + ((tsc & 1) << 3)
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *NoncurrentNativeNASKeySetIdentifier) GetNasKeySetIdentifiler() (nasKeySetIdentifiler uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// NoncurrentNativeNASKeySetIdentifier 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *NoncurrentNativeNASKeySetIdentifier) SetNasKeySetIdentifiler(nasKeySetIdentifiler uint8) {
	a.Octet = (a.Octet & 248) + (nasKeySetIdentifiler & 7)
}
