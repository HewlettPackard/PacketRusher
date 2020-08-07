package nasType

// FullNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 ,1
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
// AddCI Row, sBit, len = [0, 0], 4 , 1
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
// TextString Row, sBit, len = [1, 1], 4 , INF
type FullNameForNetwork struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewFullNameForNetwork(iei uint8) (fullNameForNetwork *FullNameForNetwork) {
	fullNameForNetwork = &FullNameForNetwork{}
	fullNameForNetwork.SetIei(iei)
	return fullNameForNetwork
}

// FullNameForNetwork 9.11.3.35
// Iei Row, sBit, len = [], 8, 8
func (a *FullNameForNetwork) GetIei() (iei uint8) {
	return a.Iei
}

// FullNameForNetwork 9.11.3.35
// Iei Row, sBit, len = [], 8, 8
func (a *FullNameForNetwork) SetIei(iei uint8) {
	a.Iei = iei
}

// FullNameForNetwork 9.11.3.35
// Len Row, sBit, len = [], 8, 8
func (a *FullNameForNetwork) GetLen() (len uint8) {
	return a.Len
}

// FullNameForNetwork 9.11.3.35
// Len Row, sBit, len = [], 8, 8
func (a *FullNameForNetwork) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// FullNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 ,1
func (a *FullNameForNetwork) GetExt() (ext uint8) {
	return a.Buffer[0] & GetBitMask(8, 7) >> (7)
}

// FullNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 ,1
func (a *FullNameForNetwork) SetExt(ext uint8) {
	a.Buffer[0] = (a.Buffer[0] & 127) + ((ext & 1) << 7)
}

// FullNameForNetwork 9.11.3.35
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
func (a *FullNameForNetwork) GetCodingScheme() (codingScheme uint8) {
	return a.Buffer[0] & GetBitMask(7, 4) >> (4)
}

// FullNameForNetwork 9.11.3.35
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
func (a *FullNameForNetwork) SetCodingScheme(codingScheme uint8) {
	a.Buffer[0] = (a.Buffer[0] & 143) + ((codingScheme & 7) << 4)
}

// FullNameForNetwork 9.11.3.35
// AddCI Row, sBit, len = [0, 0], 4 , 1
func (a *FullNameForNetwork) GetAddCI() (addCI uint8) {
	return a.Buffer[0] & GetBitMask(4, 3) >> (3)
}

// FullNameForNetwork 9.11.3.35
// AddCI Row, sBit, len = [0, 0], 4 , 1
func (a *FullNameForNetwork) SetAddCI(addCI uint8) {
	a.Buffer[0] = (a.Buffer[0] & 247) + ((addCI & 1) << 3)
}

// FullNameForNetwork 9.11.3.35
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
func (a *FullNameForNetwork) GetNumberOfSpareBitsInLastOctet() (numberOfSpareBitsInLastOctet uint8) {
	return a.Buffer[0] & GetBitMask(3, 0)
}

// FullNameForNetwork 9.11.3.35
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
func (a *FullNameForNetwork) SetNumberOfSpareBitsInLastOctet(numberOfSpareBitsInLastOctet uint8) {
	a.Buffer[0] = (a.Buffer[0] & 248) + (numberOfSpareBitsInLastOctet & 7)
}

// FullNameForNetwork 9.11.3.35
// TextString Row, sBit, len = [1, 1], 4 , INF
func (a *FullNameForNetwork) GetTextString() (textString []uint8) {
	textString = make([]uint8, len(a.Buffer)-1)
	copy(textString, a.Buffer[1:])
	return textString
}

// FullNameForNetwork 9.11.3.35
// TextString Row, sBit, len = [1, 1], 4 , INF
func (a *FullNameForNetwork) SetTextString(textString []uint8) {
	copy(a.Buffer[1:], textString)
}
