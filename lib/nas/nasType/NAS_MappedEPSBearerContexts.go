package nasType

// MappedEPSBearerContexts 9.11.4.8
// MappedEPSBearerContext Row, sBit, len = [0, 0], 8 , INF
type MappedEPSBearerContexts struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewMappedEPSBearerContexts(iei uint8) (mappedEPSBearerContexts *MappedEPSBearerContexts) {
	mappedEPSBearerContexts = &MappedEPSBearerContexts{}
	mappedEPSBearerContexts.SetIei(iei)
	return mappedEPSBearerContexts
}

// MappedEPSBearerContexts 9.11.4.8
// Iei Row, sBit, len = [], 8, 8
func (a *MappedEPSBearerContexts) GetIei() (iei uint8) {
	return a.Iei
}

// MappedEPSBearerContexts 9.11.4.8
// Iei Row, sBit, len = [], 8, 8
func (a *MappedEPSBearerContexts) SetIei(iei uint8) {
	a.Iei = iei
}

// MappedEPSBearerContexts 9.11.4.8
// Len Row, sBit, len = [], 8, 16
func (a *MappedEPSBearerContexts) GetLen() (len uint16) {
	return a.Len
}

// MappedEPSBearerContexts 9.11.4.8
// Len Row, sBit, len = [], 8, 16
func (a *MappedEPSBearerContexts) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// MappedEPSBearerContexts 9.11.4.8
// MappedEPSBearerContext Row, sBit, len = [0, 0], 8 , INF
func (a *MappedEPSBearerContexts) GetMappedEPSBearerContext() (mappedEPSBearerContext []uint8) {
	mappedEPSBearerContext = make([]uint8, len(a.Buffer))
	copy(mappedEPSBearerContext, a.Buffer)
	return mappedEPSBearerContext
}

// MappedEPSBearerContexts 9.11.4.8
// MappedEPSBearerContext Row, sBit, len = [0, 0], 8 , INF
func (a *MappedEPSBearerContexts) SetMappedEPSBearerContext(mappedEPSBearerContext []uint8) {
	copy(a.Buffer, mappedEPSBearerContext)
}
