package ie

// EmergNumList preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type EmergNumList struct {
	Contents []byte
}

func (i *EmergNumList) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("EmergNumList", b, 3, 48)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *EmergNumList) MarshalBinary() ([]byte, error) {
	return copyIEValue("EmergNumList", i.Contents, 3, 48)
}
