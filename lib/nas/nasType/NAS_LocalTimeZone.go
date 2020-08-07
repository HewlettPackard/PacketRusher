package nasType

// LocalTimeZone 9.11.3.52
// TimeZone Row, sBit, len = [0, 0], 8 , 8
type LocalTimeZone struct {
	Iei   uint8
	Octet uint8
}

func NewLocalTimeZone(iei uint8) (localTimeZone *LocalTimeZone) {
	localTimeZone = &LocalTimeZone{}
	localTimeZone.SetIei(iei)
	return localTimeZone
}

// LocalTimeZone 9.11.3.52
// Iei Row, sBit, len = [], 8, 8
func (a *LocalTimeZone) GetIei() (iei uint8) {
	return a.Iei
}

// LocalTimeZone 9.11.3.52
// Iei Row, sBit, len = [], 8, 8
func (a *LocalTimeZone) SetIei(iei uint8) {
	a.Iei = iei
}

// LocalTimeZone 9.11.3.52
// TimeZone Row, sBit, len = [0, 0], 8 , 8
func (a *LocalTimeZone) GetTimeZone() (timeZone uint8) {
	return a.Octet
}

// LocalTimeZone 9.11.3.52
// TimeZone Row, sBit, len = [0, 0], 8 , 8
func (a *LocalTimeZone) SetTimeZone(timeZone uint8) {
	a.Octet = timeZone
}
