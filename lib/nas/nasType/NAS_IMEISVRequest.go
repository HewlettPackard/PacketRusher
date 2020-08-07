package nasType

// IMEISVRequest 9.11.3.28
// Iei Row, sBit, len = [0, 0], 8 , 4
// IMEISVRequestValue Row, sBit, len = [0, 0], 3 , 3
type IMEISVRequest struct {
	Octet uint8
}

func NewIMEISVRequest(iei uint8) (iMEISVRequest *IMEISVRequest) {
	iMEISVRequest = &IMEISVRequest{}
	iMEISVRequest.SetIei(iei)
	return iMEISVRequest
}

// IMEISVRequest 9.11.3.28
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *IMEISVRequest) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// IMEISVRequest 9.11.3.28
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *IMEISVRequest) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// IMEISVRequest 9.11.3.28
// IMEISVRequestValue Row, sBit, len = [0, 0], 3 , 3
func (a *IMEISVRequest) GetIMEISVRequestValue() (iMEISVRequestValue uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// IMEISVRequest 9.11.3.28
// IMEISVRequestValue Row, sBit, len = [0, 0], 3 , 3
func (a *IMEISVRequest) SetIMEISVRequestValue(iMEISVRequestValue uint8) {
	a.Octet = (a.Octet & 248) + (iMEISVRequestValue & 7)
}
