package nasType

// TMSI5GS 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
// AMFSetID Row, sBit, len = [1, 1], 8 , 8
// AMFSetID Row, sBit, len = [1, 2], 8 , 10
// AMFPointer Row, sBit, len = [2, 2], 6 , 6
// TMSI5G Row, sBit, len = [3, 6], 8 , 32
type TMSI5GS struct {
	Iei   uint8
	Len   uint16
	Octet [7]uint8
}

func NewTMSI5GS(iei uint8) (tMSI5GS *TMSI5GS) {
	tMSI5GS = &TMSI5GS{}
	tMSI5GS.SetIei(iei)
	return tMSI5GS
}

// TMSI5GS 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *TMSI5GS) GetIei() (iei uint8) {
	return a.Iei
}

// TMSI5GS 9.11.3.4
// Iei Row, sBit, len = [], 8, 8
func (a *TMSI5GS) SetIei(iei uint8) {
	a.Iei = iei
}

// TMSI5GS 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *TMSI5GS) GetLen() (len uint16) {
	return a.Len
}

// TMSI5GS 9.11.3.4
// Len Row, sBit, len = [], 8, 16
func (a *TMSI5GS) SetLen(len uint16) {
	a.Len = len
}

// TMSI5GS 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *TMSI5GS) GetSpare() (spare uint8) {
	return a.Octet[0] & GetBitMask(4, 3) >> (3)
}

// TMSI5GS 9.11.3.4
// Spare Row, sBit, len = [0, 0], 4 , 1
func (a *TMSI5GS) SetSpare(spare uint8) {
	a.Octet[0] = (a.Octet[0] & 247) + ((spare & 1) << 3)
}

// TMSI5GS 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *TMSI5GS) GetTypeOfIdentity() (typeOfIdentity uint8) {
	return a.Octet[0] & GetBitMask(3, 0)
}

// TMSI5GS 9.11.3.4
// TypeOfIdentity Row, sBit, len = [0, 0], 3 , 3
func (a *TMSI5GS) SetTypeOfIdentity(typeOfIdentity uint8) {
	a.Octet[0] = (a.Octet[0] & 248) + (typeOfIdentity & 7)
}

// TMSI5GS 9.11.3.4
// AMFSetID Row, sBit, len = [1, 2], 8 , 10
func (a *TMSI5GS) GetAMFSetID() (aMFSetID uint16) {
	return (uint16(a.Octet[1])<<2 + uint16((a.Octet[2])&GetBitMask(8, 2))>>6)
}

// TMSI5GS 9.11.3.4
// AMFSetID Row, sBit, len = [1, 2], 8 , 10
func (a *TMSI5GS) SetAMFSetID(aMFSetID uint16) {
	a.Octet[1] = uint8((aMFSetID)>>2) & 255
	a.Octet[2] = a.Octet[2]&GetBitMask(6, 6) + uint8(aMFSetID&3)<<6
}

// TMSI5GS 9.11.3.4
// AMFPointer Row, sBit, len = [2, 2], 6 , 6
func (a *TMSI5GS) GetAMFPointer() (aMFPointer uint8) {
	return a.Octet[2] & GetBitMask(6, 0)
}

// TMSI5GS 9.11.3.4
// AMFPointer Row, sBit, len = [2, 2], 6 , 6
func (a *TMSI5GS) SetAMFPointer(aMFPointer uint8) {
	a.Octet[2] = (a.Octet[2] & 192) + (aMFPointer & 63)
}

// TMSI5GS 9.11.3.4
// TMSI5G Row, sBit, len = [3, 6], 8 , 32
func (a *TMSI5GS) GetTMSI5G() (tMSI5G [4]uint8) {
	copy(tMSI5G[:], a.Octet[3:7])
	return tMSI5G
}

// TMSI5GS 9.11.3.4
// TMSI5G Row, sBit, len = [3, 6], 8 , 32
func (a *TMSI5GS) SetTMSI5G(tMSI5G [4]uint8) {
	copy(a.Octet[3:7], tMSI5G[:])
}
