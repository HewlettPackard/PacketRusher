package ie

// OperatorDefinedAccessCategoryDefs preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type OperatorDefinedAccessCategoryDefs struct {
	Contents []byte
}

func (i *OperatorDefinedAccessCategoryDefs) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("OperatorDefinedAccessCategoryDefs", b, 0, 8320)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *OperatorDefinedAccessCategoryDefs) MarshalBinary() ([]byte, error) {
	return copyIEValue("OperatorDefinedAccessCategoryDefs", i.Contents, 0, 8320)
}
