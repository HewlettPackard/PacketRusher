package nasType

// ConfiguredNSSAI 9.11.3.37
// SNSSAIValue Row, sBit, len = [0, 0], 0 , INF
type ConfiguredNSSAI struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewConfiguredNSSAI(iei uint8) (configuredNSSAI *ConfiguredNSSAI) {
	configuredNSSAI = &ConfiguredNSSAI{}
	configuredNSSAI.SetIei(iei)
	return configuredNSSAI
}

// ConfiguredNSSAI 9.11.3.37
// Iei Row, sBit, len = [], 8, 8
func (a *ConfiguredNSSAI) GetIei() (iei uint8) {
	return a.Iei
}

// ConfiguredNSSAI 9.11.3.37
// Iei Row, sBit, len = [], 8, 8
func (a *ConfiguredNSSAI) SetIei(iei uint8) {
	a.Iei = iei
}

// ConfiguredNSSAI 9.11.3.37
// Len Row, sBit, len = [], 8, 8
func (a *ConfiguredNSSAI) GetLen() (len uint8) {
	return a.Len
}

// ConfiguredNSSAI 9.11.3.37
// Len Row, sBit, len = [], 8, 8
func (a *ConfiguredNSSAI) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// ConfiguredNSSAI 9.11.3.37
// SNSSAIValue Row, sBit, len = [0, 0], 0 , INF
func (a *ConfiguredNSSAI) GetSNSSAIValue() (sNSSAIValue []uint8) {
	sNSSAIValue = make([]uint8, len(a.Buffer))
	copy(sNSSAIValue, a.Buffer)
	return sNSSAIValue
}

// ConfiguredNSSAI 9.11.3.37
// SNSSAIValue Row, sBit, len = [0, 0], 0 , INF
func (a *ConfiguredNSSAI) SetSNSSAIValue(sNSSAIValue []uint8) {
	copy(a.Buffer, sNSSAIValue)
}
