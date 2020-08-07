package nasType

// MaximumNumberOfSupportedPacketFilters 9.11.4.9
// MaximumNumberOfSupportedPacketFilters Row, sBit, len = [0, 1], 8 , 10
type MaximumNumberOfSupportedPacketFilters struct {
	Iei   uint8
	Octet [2]uint8
}

func NewMaximumNumberOfSupportedPacketFilters(iei uint8) (maximumNumberOfSupportedPacketFilters *MaximumNumberOfSupportedPacketFilters) {
	maximumNumberOfSupportedPacketFilters = &MaximumNumberOfSupportedPacketFilters{}
	maximumNumberOfSupportedPacketFilters.SetIei(iei)
	return maximumNumberOfSupportedPacketFilters
}

// MaximumNumberOfSupportedPacketFilters 9.11.4.9
// Iei Row, sBit, len = [], 8, 8
func (a *MaximumNumberOfSupportedPacketFilters) GetIei() (iei uint8) {
	return a.Iei
}

// MaximumNumberOfSupportedPacketFilters 9.11.4.9
// Iei Row, sBit, len = [], 8, 8
func (a *MaximumNumberOfSupportedPacketFilters) SetIei(iei uint8) {
	a.Iei = iei
}

// MaximumNumberOfSupportedPacketFilters 9.11.4.9
// MaximumNumberOfSupportedPacketFilters Row, sBit, len = [0, 1], 8 , 10
func (a *MaximumNumberOfSupportedPacketFilters) GetMaximumNumberOfSupportedPacketFilters() (maximumNumberOfSupportedPacketFilters uint16) {
	return (uint16(a.Octet[0])<<2 + uint16((a.Octet[1])&GetBitMask(8, 2))>>6)
}

// MaximumNumberOfSupportedPacketFilters 9.11.4.9
// MaximumNumberOfSupportedPacketFilters Row, sBit, len = [0, 1], 8 , 10
func (a *MaximumNumberOfSupportedPacketFilters) SetMaximumNumberOfSupportedPacketFilters(maximumNumberOfSupportedPacketFilters uint16) {
	a.Octet[0] = uint8((maximumNumberOfSupportedPacketFilters)>>2) & 255
	a.Octet[1] = a.Octet[1]&GetBitMask(6, 6) + uint8(maximumNumberOfSupportedPacketFilters&3)<<6
}
