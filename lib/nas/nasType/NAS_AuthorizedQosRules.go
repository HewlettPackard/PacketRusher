package nasType

// AuthorizedQosRules 9.11.4.13
// QosRule Row, sBit, len = [0, 0], 3 , INF
type AuthorizedQosRules struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewAuthorizedQosRules(iei uint8) (authorizedQosRules *AuthorizedQosRules) {
	authorizedQosRules = &AuthorizedQosRules{}
	authorizedQosRules.SetIei(iei)
	return authorizedQosRules
}

// AuthorizedQosRules 9.11.4.13
// Iei Row, sBit, len = [], 8, 8
func (a *AuthorizedQosRules) GetIei() (iei uint8) {
	return a.Iei
}

// AuthorizedQosRules 9.11.4.13
// Iei Row, sBit, len = [], 8, 8
func (a *AuthorizedQosRules) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthorizedQosRules 9.11.4.13
// Len Row, sBit, len = [], 8, 16
func (a *AuthorizedQosRules) GetLen() (len uint16) {
	return a.Len
}

// AuthorizedQosRules 9.11.4.13
// Len Row, sBit, len = [], 8, 16
func (a *AuthorizedQosRules) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// AuthorizedQosRules 9.11.4.13
// QosRule Row, sBit, len = [0, 0], 3 , INF
func (a *AuthorizedQosRules) GetQosRule() (qosRule []uint8) {
	qosRule = make([]uint8, len(a.Buffer))
	copy(qosRule, a.Buffer)
	return qosRule
}

// AuthorizedQosRules 9.11.4.13
// QosRule Row, sBit, len = [0, 0], 3 , INF
func (a *AuthorizedQosRules) SetQosRule(qosRule []uint8) {
	copy(a.Buffer, qosRule)
}
