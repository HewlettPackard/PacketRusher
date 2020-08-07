package nasType

// Cause5GMM 9.11.3.2
// CauseValue Row, sBit, len = [0, 0], 8 , 8
type Cause5GMM struct {
	Iei   uint8
	Octet uint8
}

func NewCause5GMM(iei uint8) (cause5GMM *Cause5GMM) {
	cause5GMM = &Cause5GMM{}
	cause5GMM.SetIei(iei)
	return cause5GMM
}

// Cause5GMM 9.11.3.2
// Iei Row, sBit, len = [], 8, 8
func (a *Cause5GMM) GetIei() (iei uint8) {
	return a.Iei
}

// Cause5GMM 9.11.3.2
// Iei Row, sBit, len = [], 8, 8
func (a *Cause5GMM) SetIei(iei uint8) {
	a.Iei = iei
}

// Cause5GMM 9.11.3.2
// CauseValue Row, sBit, len = [0, 0], 8 , 8
func (a *Cause5GMM) GetCauseValue() (causeValue uint8) {
	return a.Octet
}

// Cause5GMM 9.11.3.2
// CauseValue Row, sBit, len = [0, 0], 8 , 8
func (a *Cause5GMM) SetCauseValue(causeValue uint8) {
	a.Octet = causeValue
}
