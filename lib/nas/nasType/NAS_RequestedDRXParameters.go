package nasType

// RequestedDRXParameters 9.11.3.2A
// DRXValue Row, sBit, len = [0, 0], 4 , 4
type RequestedDRXParameters struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewRequestedDRXParameters(iei uint8) (requestedDRXParameters *RequestedDRXParameters) {
	requestedDRXParameters = &RequestedDRXParameters{}
	requestedDRXParameters.SetIei(iei)
	return requestedDRXParameters
}

// RequestedDRXParameters 9.11.3.2A
// Iei Row, sBit, len = [], 8, 8
func (a *RequestedDRXParameters) GetIei() (iei uint8) {
	return a.Iei
}

// RequestedDRXParameters 9.11.3.2A
// Iei Row, sBit, len = [], 8, 8
func (a *RequestedDRXParameters) SetIei(iei uint8) {
	a.Iei = iei
}

// RequestedDRXParameters 9.11.3.2A
// Len Row, sBit, len = [], 8, 8
func (a *RequestedDRXParameters) GetLen() (len uint8) {
	return a.Len
}

// RequestedDRXParameters 9.11.3.2A
// Len Row, sBit, len = [], 8, 8
func (a *RequestedDRXParameters) SetLen(len uint8) {
	a.Len = len
}

// RequestedDRXParameters 9.11.3.2A
// DRXValue Row, sBit, len = [0, 0], 4 , 4
func (a *RequestedDRXParameters) GetDRXValue() (dRXValue uint8) {
	return a.Octet & GetBitMask(4, 0)
}

// RequestedDRXParameters 9.11.3.2A
// DRXValue Row, sBit, len = [0, 0], 4 , 4
func (a *RequestedDRXParameters) SetDRXValue(dRXValue uint8) {
	a.Octet = (a.Octet & 240) + (dRXValue & 15)
}
