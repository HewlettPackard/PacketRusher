package nasType

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
// SwitchOff Row, sBit, len = [0, 0], 4 , 1
// ReRegistrationRequired Row, sBit, len = [0, 0], 3 , 1
// AccessType Row, sBit, len = [0, 0], 2 , 2
type NgksiAndDeregistrationType struct {
	Octet uint8
}

func NewNgksiAndDeregistrationType() (ngksiAndDeregistrationType *NgksiAndDeregistrationType) {
	ngksiAndDeregistrationType = &NgksiAndDeregistrationType{}
	return ngksiAndDeregistrationType
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
func (a *NgksiAndDeregistrationType) GetTSC() (tSC uint8) {
	return a.Octet & GetBitMask(8, 7) >> (7)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
func (a *NgksiAndDeregistrationType) SetTSC(tSC uint8) {
	a.Octet = (a.Octet & 127) + ((tSC & 1) << 7)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
func (a *NgksiAndDeregistrationType) GetNasKeySetIdentifiler() (nasKeySetIdentifiler uint8) {
	return a.Octet & GetBitMask(7, 4) >> (4)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
func (a *NgksiAndDeregistrationType) SetNasKeySetIdentifiler(nasKeySetIdentifiler uint8) {
	a.Octet = (a.Octet & 143) + ((nasKeySetIdentifiler & 7) << 4)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// SwitchOff Row, sBit, len = [0, 0], 4 , 1
func (a *NgksiAndDeregistrationType) GetSwitchOff() (switchOff uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// SwitchOff Row, sBit, len = [0, 0], 4 , 1
func (a *NgksiAndDeregistrationType) SetSwitchOff(switchOff uint8) {
	a.Octet = (a.Octet & 247) + ((switchOff & 1) << 3)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// ReRegistrationRequired Row, sBit, len = [0, 0], 3 , 1
func (a *NgksiAndDeregistrationType) GetReRegistrationRequired() (reRegistrationRequired uint8) {
	return a.Octet & GetBitMask(3, 2) >> (2)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// ReRegistrationRequired Row, sBit, len = [0, 0], 3 , 1
func (a *NgksiAndDeregistrationType) SetReRegistrationRequired(reRegistrationRequired uint8) {
	a.Octet = (a.Octet & 251) + ((reRegistrationRequired & 1) << 2)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// AccessType Row, sBit, len = [0, 0], 2 , 2
func (a *NgksiAndDeregistrationType) GetAccessType() (accessType uint8) {
	return a.Octet & GetBitMask(2, 0)
}

// NgksiAndDeregistrationType 9.11.3.20 9.11.3.32
// AccessType Row, sBit, len = [0, 0], 2 , 2
func (a *NgksiAndDeregistrationType) SetAccessType(accessType uint8) {
	a.Octet = (a.Octet & 252) + (accessType & 3)
}
