package nasType

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// OperatorDefinedAccessCategoryDefintiion Row, sBit, len = [0, 0], 8 , INF
type OperatordefinedAccessCategoryDefinitions struct {
	Iei    uint8
	Len    uint16
	Buffer []uint8
}

func NewOperatordefinedAccessCategoryDefinitions(iei uint8) (operatordefinedAccessCategoryDefinitions *OperatordefinedAccessCategoryDefinitions) {
	operatordefinedAccessCategoryDefinitions = &OperatordefinedAccessCategoryDefinitions{}
	operatordefinedAccessCategoryDefinitions.SetIei(iei)
	return operatordefinedAccessCategoryDefinitions
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// Iei Row, sBit, len = [], 8, 8
func (a *OperatordefinedAccessCategoryDefinitions) GetIei() (iei uint8) {
	return a.Iei
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// Iei Row, sBit, len = [], 8, 8
func (a *OperatordefinedAccessCategoryDefinitions) SetIei(iei uint8) {
	a.Iei = iei
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// Len Row, sBit, len = [], 8, 16
func (a *OperatordefinedAccessCategoryDefinitions) GetLen() (len uint16) {
	return a.Len
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// Len Row, sBit, len = [], 8, 16
func (a *OperatordefinedAccessCategoryDefinitions) SetLen(len uint16) {
	a.Len = len
	a.Buffer = make([]uint8, a.Len)
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// OperatorDefinedAccessCategoryDefintiion Row, sBit, len = [0, 0], 8 , INF
func (a *OperatordefinedAccessCategoryDefinitions) GetOperatorDefinedAccessCategoryDefintiion() (operatorDefinedAccessCategoryDefintiion []uint8) {
	operatorDefinedAccessCategoryDefintiion = make([]uint8, len(a.Buffer))
	copy(operatorDefinedAccessCategoryDefintiion, a.Buffer)
	return operatorDefinedAccessCategoryDefintiion
}

// OperatordefinedAccessCategoryDefinitions 9.11.3.38
// OperatorDefinedAccessCategoryDefintiion Row, sBit, len = [0, 0], 8 , INF
func (a *OperatordefinedAccessCategoryDefinitions) SetOperatorDefinedAccessCategoryDefintiion(operatorDefinedAccessCategoryDefintiion []uint8) {
	copy(a.Buffer, operatorDefinedAccessCategoryDefintiion)
}
