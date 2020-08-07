package nasType

// RequestedQosRules 9.11.4.13
// QoSRules Row, sBit, len = [0, 0], 8 , INF
type RequestedQosRules struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewRequestedQosRules(iei uint8) (requestedQosRules *RequestedQosRules) {
	requestedQosRules = &RequestedQosRules{}
	requestedQosRules.SetIei(iei)
	return requestedQosRules
}

// RequestedQosRules 9.11.4.13
// Iei Row, sBit, len = [], 8, 8
func (a *RequestedQosRules) GetIei() (iei uint8) {
	return a.Iei
}

// RequestedQosRules 9.11.4.13
// Iei Row, sBit, len = [], 8, 8
func (a *RequestedQosRules) SetIei(iei uint8) {
	a.Iei = iei
}

// RequestedQosRules 9.11.4.13
// Len Row, sBit, len = [], 8, 8
func (a *RequestedQosRules) GetLen() (len uint8) {
	return a.Len
}

// RequestedQosRules 9.11.4.13
// Len Row, sBit, len = [], 8, 8
func (a *RequestedQosRules) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// RequestedQosRules 9.11.4.13
// QoSRules Row, sBit, len = [0, 0], 8 , INF
func (a *RequestedQosRules) GetQoSRules() (qoSRules []uint8) {
	qoSRules = make([]uint8, len(a.Buffer))
	copy(qoSRules, a.Buffer)
	return qoSRules
}

// RequestedQosRules 9.11.4.13
// QoSRules Row, sBit, len = [0, 0], 8 , INF
func (a *RequestedQosRules) SetQoSRules(qoSRules []uint8) {
	copy(a.Buffer, qoSRules)
}
