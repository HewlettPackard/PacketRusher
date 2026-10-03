package ie

// S1UESecCapability preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type S1UESecCapability struct {
	Contents []byte
}

func (i *S1UESecCapability) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("S1UESecCapability", b, 2, 5)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *S1UESecCapability) MarshalBinary() ([]byte, error) {
	return copyIEValue("S1UESecCapability", i.Contents, 2, 5)
}
