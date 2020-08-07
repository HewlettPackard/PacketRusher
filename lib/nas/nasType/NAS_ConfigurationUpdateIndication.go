package nasType

// ConfigurationUpdateIndication 9.11.3.18
// Iei Row, sBit, len = [0, 0], 8 , 4
// RED Row, sBit, len = [0, 0], 2 , 1
// ACK Row, sBit, len = [0, 0], 1 , 1
type ConfigurationUpdateIndication struct {
	Octet uint8
}

func NewConfigurationUpdateIndication(iei uint8) (configurationUpdateIndication *ConfigurationUpdateIndication) {
	configurationUpdateIndication = &ConfigurationUpdateIndication{}
	configurationUpdateIndication.SetIei(iei)
	return configurationUpdateIndication
}

// ConfigurationUpdateIndication 9.11.3.18
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *ConfigurationUpdateIndication) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// ConfigurationUpdateIndication 9.11.3.18
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *ConfigurationUpdateIndication) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// ConfigurationUpdateIndication 9.11.3.18
// RED Row, sBit, len = [0, 0], 2 , 1
func (a *ConfigurationUpdateIndication) GetRED() (rED uint8) {
	return a.Octet & GetBitMask(2, 1) >> (1)
}

// ConfigurationUpdateIndication 9.11.3.18
// RED Row, sBit, len = [0, 0], 2 , 1
func (a *ConfigurationUpdateIndication) SetRED(rED uint8) {
	a.Octet = (a.Octet & 253) + ((rED & 1) << 1)
}

// ConfigurationUpdateIndication 9.11.3.18
// ACK Row, sBit, len = [0, 0], 1 , 1
func (a *ConfigurationUpdateIndication) GetACK() (aCK uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// ConfigurationUpdateIndication 9.11.3.18
// ACK Row, sBit, len = [0, 0], 1 , 1
func (a *ConfigurationUpdateIndication) SetACK(aCK uint8) {
	a.Octet = (a.Octet & 254) + (aCK & 1)
}
