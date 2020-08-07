package nasType

// NSSAIInclusionMode 9.11.3.37A
// Iei Row, sBit, len = [0, 0], 8 , 4
// NSSAIInclusionMode Row, sBit, len = [0, 0], 2 , 2
type NSSAIInclusionMode struct {
	Octet uint8
}

func NewNSSAIInclusionMode(iei uint8) (nSSAIInclusionMode *NSSAIInclusionMode) {
	nSSAIInclusionMode = &NSSAIInclusionMode{}
	nSSAIInclusionMode.SetIei(iei)
	return nSSAIInclusionMode
}

// NSSAIInclusionMode 9.11.3.37A
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NSSAIInclusionMode) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// NSSAIInclusionMode 9.11.3.37A
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *NSSAIInclusionMode) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// NSSAIInclusionMode 9.11.3.37A
// NSSAIInclusionMode Row, sBit, len = [0, 0], 2 , 2
func (a *NSSAIInclusionMode) GetNSSAIInclusionMode() (nSSAIInclusionMode uint8) {
	return a.Octet & GetBitMask(2, 0)
}

// NSSAIInclusionMode 9.11.3.37A
// NSSAIInclusionMode Row, sBit, len = [0, 0], 2 , 2
func (a *NSSAIInclusionMode) SetNSSAIInclusionMode(nSSAIInclusionMode uint8) {
	a.Octet = (a.Octet & 252) + (nSSAIInclusionMode & 3)
}
