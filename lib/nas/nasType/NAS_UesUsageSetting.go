package nasType

// UesUsageSetting 9.11.3.55
// UesUsageSetting Row, sBit, len = [0, 0], 1 , 1
type UesUsageSetting struct {
	Iei   uint8
	Len   uint8
	Octet uint8
}

func NewUesUsageSetting(iei uint8) (uesUsageSetting *UesUsageSetting) {
	uesUsageSetting = &UesUsageSetting{}
	uesUsageSetting.SetIei(iei)
	return uesUsageSetting
}

// UesUsageSetting 9.11.3.55
// Iei Row, sBit, len = [], 8, 8
func (a *UesUsageSetting) GetIei() (iei uint8) {
	return a.Iei
}

// UesUsageSetting 9.11.3.55
// Iei Row, sBit, len = [], 8, 8
func (a *UesUsageSetting) SetIei(iei uint8) {
	a.Iei = iei
}

// UesUsageSetting 9.11.3.55
// Len Row, sBit, len = [], 8, 8
func (a *UesUsageSetting) GetLen() (len uint8) {
	return a.Len
}

// UesUsageSetting 9.11.3.55
// Len Row, sBit, len = [], 8, 8
func (a *UesUsageSetting) SetLen(len uint8) {
	a.Len = len
}

// UesUsageSetting 9.11.3.55
// UesUsageSetting Row, sBit, len = [0, 0], 1 , 1
func (a *UesUsageSetting) GetUesUsageSetting() (uesUsageSetting uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// UesUsageSetting 9.11.3.55
// UesUsageSetting Row, sBit, len = [0, 0], 1 , 1
func (a *UesUsageSetting) SetUesUsageSetting(uesUsageSetting uint8) {
	a.Octet = (a.Octet & 254) + (uesUsageSetting & 1)
}
