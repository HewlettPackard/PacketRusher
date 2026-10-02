package ie

// AdditionalInfo preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type AdditionalInfo struct {
	Contents []byte
}

func (i *AdditionalInfo) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("AdditionalInfo", b, 1, 255)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *AdditionalInfo) MarshalBinary() ([]byte, error) {
	return copyIEValue("AdditionalInfo", i.Contents, 1, 255)
}
