package nasType

// NetworkDaylightSavingTime 9.11.3.19
// value Row, sBit, len = [0, 0], 2 , 2
type NetworkDaylightSavingTime struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewNetworkDaylightSavingTime(iei uint8) (networkDaylightSavingTime *NetworkDaylightSavingTime) {
	networkDaylightSavingTime = &NetworkDaylightSavingTime{}
	networkDaylightSavingTime.SetIei(iei)
	return networkDaylightSavingTime
}

// NetworkDaylightSavingTime 9.11.3.19
// Iei Row, sBit, len = [], 8, 8
func (a *NetworkDaylightSavingTime) GetIei() (iei uint8) {
	return a.Iei
}

// NetworkDaylightSavingTime 9.11.3.19
// Iei Row, sBit, len = [], 8, 8
func (a *NetworkDaylightSavingTime) SetIei(iei uint8) {
	a.Iei = iei
}

// NetworkDaylightSavingTime 9.11.3.19
// Len Row, sBit, len = [], 8, 8
func (a *NetworkDaylightSavingTime) GetLen() (len uint8) {
	return a.Len
}

// NetworkDaylightSavingTime 9.11.3.19
// Len Row, sBit, len = [], 8, 8
func (a *NetworkDaylightSavingTime) SetLen(len uint8) {
	a.Len = len
}

// NetworkDaylightSavingTime 9.11.3.19
// value Row, sBit, len = [0, 0], 2 , 2
func (a *NetworkDaylightSavingTime) Getvalue() (value uint8) {
	return a.Octet & GetBitMask(2, 0)
}

// NetworkDaylightSavingTime 9.11.3.19
// value Row, sBit, len = [0, 0], 2 , 2
func (a *NetworkDaylightSavingTime) Setvalue(value uint8) {
	a.Octet = (a.Octet & 252) + (value & 3)
}
