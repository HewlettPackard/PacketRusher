package nasType

// UpdateType5GS 9.11.3.9A
// NGRanRcu Row, sBit, len = [0, 0], 2 , 1
// SMSRequested Row, sBit, len = [0, 0], 1 , 1
type UpdateType5GS struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewUpdateType5GS(iei uint8) (updateType5GS *UpdateType5GS) {
	updateType5GS = &UpdateType5GS{}
	updateType5GS.SetIei(iei)
	return updateType5GS
}

// UpdateType5GS 9.11.3.9A
// Iei Row, sBit, len = [], 8, 8
func (a *UpdateType5GS) GetIei() (iei uint8) {
	return a.Iei
}

// UpdateType5GS 9.11.3.9A
// Iei Row, sBit, len = [], 8, 8
func (a *UpdateType5GS) SetIei(iei uint8) {
	a.Iei = iei
}

// UpdateType5GS 9.11.3.9A
// Len Row, sBit, len = [], 8, 8
func (a *UpdateType5GS) GetLen() (len uint8) {
	return a.Len
}

// UpdateType5GS 9.11.3.9A
// Len Row, sBit, len = [], 8, 8
func (a *UpdateType5GS) SetLen(len uint8) {
	a.Len = len
}

// UpdateType5GS 9.11.3.9A
// NGRanRcu Row, sBit, len = [0, 0], 2 , 1
func (a *UpdateType5GS) GetNGRanRcu() (nGRanRcu uint8) {
	return a.Octet & GetBitMask(2, 1) >> (1)
}

// UpdateType5GS 9.11.3.9A
// NGRanRcu Row, sBit, len = [0, 0], 2 , 1
func (a *UpdateType5GS) SetNGRanRcu(nGRanRcu uint8) {
	a.Octet = (a.Octet & 253) + ((nGRanRcu & 1) << 1)
}

// UpdateType5GS 9.11.3.9A
// SMSRequested Row, sBit, len = [0, 0], 1 , 1
func (a *UpdateType5GS) GetSMSRequested() (sMSRequested uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// UpdateType5GS 9.11.3.9A
// SMSRequested Row, sBit, len = [0, 0], 1 , 1
func (a *UpdateType5GS) SetSMSRequested(sMSRequested uint8) {
	a.Octet = (a.Octet & 254) + (sMSRequested & 1)
}
