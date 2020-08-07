package nasType

// ExtendedProtocolConfigurationOptions 9.11.4.6
// ExtendedProtocolConfigurationOptionsContents Row, sBit, len = [0, 0], 8 , INF
type ExtendedProtocolConfigurationOptions struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewExtendedProtocolConfigurationOptions(iei uint8) (extendedProtocolConfigurationOptions *ExtendedProtocolConfigurationOptions) {
	extendedProtocolConfigurationOptions = &ExtendedProtocolConfigurationOptions{}
	extendedProtocolConfigurationOptions.SetIei(iei)
	return extendedProtocolConfigurationOptions
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// Iei Row, sBit, len = [], 8, 8
func (a *ExtendedProtocolConfigurationOptions) GetIei() (iei uint8) {
	return a.Iei
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// Iei Row, sBit, len = [], 8, 8
func (a *ExtendedProtocolConfigurationOptions) SetIei(iei uint8) {
	a.Iei = iei
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// Len Row, sBit, len = [], 8, 16
func (a *ExtendedProtocolConfigurationOptions) GetLen() (len uint16) {
	return a.Len
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// Len Row, sBit, len = [], 8, 16
func (a *ExtendedProtocolConfigurationOptions) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// ExtendedProtocolConfigurationOptionsContents Row, sBit, len = [0, 0], 8 , INF
func (a *ExtendedProtocolConfigurationOptions) GetExtendedProtocolConfigurationOptionsContents() (extendedProtocolConfigurationOptionsContents []uint8) {
	extendedProtocolConfigurationOptionsContents = make([]uint8, len(a.Buffer))
	copy(extendedProtocolConfigurationOptionsContents, a.Buffer)
	return extendedProtocolConfigurationOptionsContents
}

// ExtendedProtocolConfigurationOptions 9.11.4.6
// ExtendedProtocolConfigurationOptionsContents Row, sBit, len = [0, 0], 8 , INF
func (a *ExtendedProtocolConfigurationOptions) SetExtendedProtocolConfigurationOptionsContents(extendedProtocolConfigurationOptionsContents []uint8) {
	copy(a.Buffer, extendedProtocolConfigurationOptionsContents)
}
