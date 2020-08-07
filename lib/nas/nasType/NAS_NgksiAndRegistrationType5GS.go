package nasType

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
// FOR  Row, sBit, len = [0, 0], 4 , 1
// RegistrationType5GS Row, sBit, len = [0, 0], 3 , 3
type NgksiAndRegistrationType5GS struct {
	Octet uint8
}

func NewNgksiAndRegistrationType5GS() (ngksiAndRegistrationType5GS *NgksiAndRegistrationType5GS) {
	ngksiAndRegistrationType5GS = &NgksiAndRegistrationType5GS{}
	return ngksiAndRegistrationType5GS
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
func (a *NgksiAndRegistrationType5GS) GetTSC() (tSC uint8) {
	return a.Octet & GetBitMask(8, 7) >> (7)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// TSC Row, sBit, len = [0, 0], 8 , 1
func (a *NgksiAndRegistrationType5GS) SetTSC(tSC uint8) {
	a.Octet = (a.Octet & 127) + ((tSC & 1) << 7)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
func (a *NgksiAndRegistrationType5GS) GetNasKeySetIdentifiler() (nasKeySetIdentifiler uint8) {
	return a.Octet & GetBitMask(7, 4) >> (4)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 7 , 3
func (a *NgksiAndRegistrationType5GS) SetNasKeySetIdentifiler(nasKeySetIdentifiler uint8) {
	a.Octet = (a.Octet & 143) + ((nasKeySetIdentifiler & 7) << 4)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// FOR Row, sBit, len = [0, 0], 4 , 1
func (a *NgksiAndRegistrationType5GS) GetFOR() (fOR uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// FOR Row, sBit, len = [0, 0], 4 , 1
func (a *NgksiAndRegistrationType5GS) SetFOR(fOR uint8) {
	a.Octet = (a.Octet & 247) + ((fOR & 1) << 3)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// RegistrationType5GS Row, sBit, len = [0, 0], 3 , 3
func (a *NgksiAndRegistrationType5GS) GetRegistrationType5GS() (registrationType5GS uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// NgksiAndRegistrationType5GS 9.11.3.7 9.11.3.32
// RegistrationType5GS Row, sBit, len = [0, 0], 3 , 3
func (a *NgksiAndRegistrationType5GS) SetRegistrationType5GS(registrationType5GS uint8) {
	a.Octet = (a.Octet & 248) + (registrationType5GS & 7)
}
