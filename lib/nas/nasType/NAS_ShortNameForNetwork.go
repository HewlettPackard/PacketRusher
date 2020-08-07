package nasType

// ShortNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 , 1
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
// AddCI Row, sBit, len = [0, 0], 4 , 1
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
// TextString Row, sBit, len = [1, 1], 4 , INF
type ShortNameForNetwork struct {
	Iei    uint8
	Len    uint8
	Buffer []uint8
}

func NewShortNameForNetwork(iei uint8) (shortNameForNetwork *ShortNameForNetwork) {
	shortNameForNetwork = &ShortNameForNetwork{}
	shortNameForNetwork.SetIei(iei)
	return shortNameForNetwork
}

// ShortNameForNetwork 9.11.3.35
// Iei Row, sBit, len = [], 8, 8
func (a *ShortNameForNetwork) GetIei() (iei uint8) {
	return a.Iei
}

// ShortNameForNetwork 9.11.3.35
// Iei Row, sBit, len = [], 8, 8
func (a *ShortNameForNetwork) SetIei(iei uint8) {
	a.Iei = iei
}

// ShortNameForNetwork 9.11.3.35
// Len Row, sBit, len = [], 8, 8
func (a *ShortNameForNetwork) GetLen() (len uint8) {
	return a.Len
}

// ShortNameForNetwork 9.11.3.35
// Len Row, sBit, len = [], 8, 8
func (a *ShortNameForNetwork) SetLen(len uint8) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// ShortNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 , 1
func (a *ShortNameForNetwork) GetExt() (ext uint8) {
	return a.Buffer[0] & GetBitMask(8, 7) >> (7)
}

// ShortNameForNetwork 9.11.3.35
// Ext Row, sBit, len = [0, 0], 8 , 1
func (a *ShortNameForNetwork) SetExt(ext uint8) {
	a.Buffer[0] = (a.Buffer[0] & 127) + ((ext & 1) << 7)
}

// ShortNameForNetwork 9.11.3.35
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
func (a *ShortNameForNetwork) GetCodingScheme() (codingScheme uint8) {
	return a.Buffer[0] & GetBitMask(7, 4) >> (4)
}

// ShortNameForNetwork 9.11.3.35
// CodingScheme Row, sBit, len = [0, 0], 7 , 3
func (a *ShortNameForNetwork) SetCodingScheme(codingScheme uint8) {
	a.Buffer[0] = (a.Buffer[0] & 143) + ((codingScheme & 7) << 4)
}

// ShortNameForNetwork 9.11.3.35
// AddCI Row, sBit, len = [0, 0], 4 , 1
func (a *ShortNameForNetwork) GetAddCI() (addCI uint8) {
	return a.Buffer[0] & GetBitMask(4, 3) >> (3)
}

// ShortNameForNetwork 9.11.3.35
// AddCI Row, sBit, len = [0, 0], 4 , 1
func (a *ShortNameForNetwork) SetAddCI(addCI uint8) {
	a.Buffer[0] = (a.Buffer[0] & 247) + ((addCI & 1) << 3)
}

// ShortNameForNetwork 9.11.3.35
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
func (a *ShortNameForNetwork) GetNumberOfSpareBitsInLastOctet() (numberOfSpareBitsInLastOctet uint8) {
	return a.Buffer[0] & GetBitMask(3, 0)
}

// ShortNameForNetwork 9.11.3.35
// NumberOfSpareBitsInLastOctet Row, sBit, len = [0, 0], 3 , 3
func (a *ShortNameForNetwork) SetNumberOfSpareBitsInLastOctet(numberOfSpareBitsInLastOctet uint8) {
	a.Buffer[0] = (a.Buffer[0] & 248) + (numberOfSpareBitsInLastOctet & 7)
}

// ShortNameForNetwork 9.11.3.35
// TextString Row, sBit, len = [1, 1], 4 , INF
func (a *ShortNameForNetwork) GetTextString() (textString []uint8) {
	textString = make([]uint8, len(a.Buffer)-1)
	copy(textString, a.Buffer[1:])
	return textString
}

// ShortNameForNetwork 9.11.3.35
// TextString Row, sBit, len = [1, 1], 4 , INF
func (a *ShortNameForNetwork) SetTextString(textString []uint8) {
	copy(a.Buffer[1:], textString)
}
