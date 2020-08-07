package nasType

// AuthorizedQosFlowDescriptions 9.11.4.12
// QoSFlowDescriptions Row, sBit, len = [0, 0], 8 , INF
type AuthorizedQosFlowDescriptions struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewAuthorizedQosFlowDescriptions(iei uint8) (authorizedQosFlowDescriptions *AuthorizedQosFlowDescriptions) {
	authorizedQosFlowDescriptions = &AuthorizedQosFlowDescriptions{}
	authorizedQosFlowDescriptions.SetIei(iei)
	return authorizedQosFlowDescriptions
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// Iei Row, sBit, len = [], 8, 8
func (a *AuthorizedQosFlowDescriptions) GetIei() (iei uint8) {
	return a.Iei
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// Iei Row, sBit, len = [], 8, 8
func (a *AuthorizedQosFlowDescriptions) SetIei(iei uint8) {
	a.Iei = iei
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// Len Row, sBit, len = [], 8, 16
func (a *AuthorizedQosFlowDescriptions) GetLen() (len uint16) {
	return a.Len
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// Len Row, sBit, len = [], 8, 16
func (a *AuthorizedQosFlowDescriptions) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// QoSFlowDescriptions Row, sBit, len = [0, 0], 8 , INF
func (a *AuthorizedQosFlowDescriptions) GetQoSFlowDescriptions() (qoSFlowDescriptions []uint8) {
	qoSFlowDescriptions = make([]uint8, len(a.Buffer))
	copy(qoSFlowDescriptions, a.Buffer)
	return qoSFlowDescriptions
}

// AuthorizedQosFlowDescriptions 9.11.4.12
// QoSFlowDescriptions Row, sBit, len = [0, 0], 8 , INF
func (a *AuthorizedQosFlowDescriptions) SetQoSFlowDescriptions(qoSFlowDescriptions []uint8) {
	copy(a.Buffer, qoSFlowDescriptions)
}
