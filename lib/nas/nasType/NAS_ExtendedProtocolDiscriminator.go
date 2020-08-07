package nasType

// ExtendedProtocolDiscriminator 9.2
// ExtendedProtocolDiscriminator Row, sBit, len = [0, 0], 8 , 8
type ExtendedProtocolDiscriminator struct {
	Octet uint8
}

func NewExtendedProtocolDiscriminator() (extendedProtocolDiscriminator *ExtendedProtocolDiscriminator) {
	extendedProtocolDiscriminator = &ExtendedProtocolDiscriminator{}
	return extendedProtocolDiscriminator
}

// ExtendedProtocolDiscriminator 9.2
// ExtendedProtocolDiscriminator Row, sBit, len = [0, 0], 8 , 8
func (a *ExtendedProtocolDiscriminator) GetExtendedProtocolDiscriminator() (extendedProtocolDiscriminator uint8) {
	return a.Octet
}

// ExtendedProtocolDiscriminator 9.2
// ExtendedProtocolDiscriminator Row, sBit, len = [0, 0], 8 , 8
func (a *ExtendedProtocolDiscriminator) SetExtendedProtocolDiscriminator(extendedProtocolDiscriminator uint8) {
	a.Octet = extendedProtocolDiscriminator
}
