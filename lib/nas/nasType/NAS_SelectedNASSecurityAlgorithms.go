package nasType

// SelectedNASSecurityAlgorithms 9.11.3.34
// TypeOfCipheringAlgorithm Row, sBit, len = [0, 0], 8 , 4
// TypeOfIntegrityProtectionAlgorithm Row, sBit, len = [0, 0], 4 , 4
type SelectedNASSecurityAlgorithms struct {
	Iei   uint8
	Octet uint8
}

func NewSelectedNASSecurityAlgorithms(iei uint8) (selectedNASSecurityAlgorithms *SelectedNASSecurityAlgorithms) {
	selectedNASSecurityAlgorithms = &SelectedNASSecurityAlgorithms{}
	selectedNASSecurityAlgorithms.SetIei(iei)
	return selectedNASSecurityAlgorithms
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// Iei Row, sBit, len = [], 8, 8
func (a *SelectedNASSecurityAlgorithms) GetIei() (iei uint8) {
	return a.Iei
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// Iei Row, sBit, len = [], 8, 8
func (a *SelectedNASSecurityAlgorithms) SetIei(iei uint8) {
	a.Iei = iei
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// TypeOfCipheringAlgorithm Row, sBit, len = [0, 0], 8 , 4
func (a *SelectedNASSecurityAlgorithms) GetTypeOfCipheringAlgorithm() (typeOfCipheringAlgorithm uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// TypeOfCipheringAlgorithm Row, sBit, len = [0, 0], 8 , 4
func (a *SelectedNASSecurityAlgorithms) SetTypeOfCipheringAlgorithm(typeOfCipheringAlgorithm uint8) {
	a.Octet = (a.Octet & 15) + ((typeOfCipheringAlgorithm & 15) << 4)
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// TypeOfIntegrityProtectionAlgorithm Row, sBit, len = [0, 0], 4 , 4
func (a *SelectedNASSecurityAlgorithms) GetTypeOfIntegrityProtectionAlgorithm() (typeOfIntegrityProtectionAlgorithm uint8) {
	return a.Octet & GetBitMask(4, 0)
}

// SelectedNASSecurityAlgorithms 9.11.3.34
// TypeOfIntegrityProtectionAlgorithm Row, sBit, len = [0, 0], 4 , 4
func (a *SelectedNASSecurityAlgorithms) SetTypeOfIntegrityProtectionAlgorithm(typeOfIntegrityProtectionAlgorithm uint8) {
	a.Octet = (a.Octet & 240) + (typeOfIntegrityProtectionAlgorithm & 15)
}
