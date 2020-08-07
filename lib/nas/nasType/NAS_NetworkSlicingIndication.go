package nasType

// NetworkSlicingIndication 9.11.3.36
// Iei Row, sBit, len = [0, 0], 8 , 4
// DCNI Row, sBit, len = [0, 0], 2 , 1
// NSSCI Row, sBit, len = [0, 0], 1 , 1
type NetworkSlicingIndication struct {
	Octet uint8
}

func NewNetworkSlicingIndication(iei uint8) (networkSlicingIndication *NetworkSlicingIndication) {
	networkSlicingIndication = &NetworkSlicingIndication{}
	networkSlicingIndication.SetIei(iei)
	return networkSlicingIndication
}

// NetworkSlicingIndication 9.11.3.36
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NetworkSlicingIndication) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// NetworkSlicingIndication 9.11.3.36
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NetworkSlicingIndication) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// NetworkSlicingIndication 9.11.3.36
// DCNI Row, sBit, len = [0, 0], 2 , 1
func (a *NetworkSlicingIndication) GetDCNI() (dCNI uint8) {
	return a.Octet & GetBitMask(2, 1) >> (1)
}

// NetworkSlicingIndication 9.11.3.36
// DCNI Row, sBit, len = [0, 0], 2 , 1
func (a *NetworkSlicingIndication) SetDCNI(dCNI uint8) {
	a.Octet = (a.Octet & 253) + ((dCNI & 1) << 1)
}

// NetworkSlicingIndication 9.11.3.36
// NSSCI Row, sBit, len = [0, 0], 1 , 1
func (a *NetworkSlicingIndication) GetNSSCI() (nSSCI uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// NetworkSlicingIndication 9.11.3.36
// NSSCI Row, sBit, len = [0, 0], 1 , 1
func (a *NetworkSlicingIndication) SetNSSCI(nSSCI uint8) {
	a.Octet = (a.Octet & 254) + (nSSCI & 1)
}
