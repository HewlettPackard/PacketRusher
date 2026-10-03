package ie

// ExtendedEmergNumList preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type ExtendedEmergNumList struct {
	Contents []byte
}

func (i *ExtendedEmergNumList) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("ExtendedEmergNumList", b, 4, 65535)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *ExtendedEmergNumList) MarshalBinary() ([]byte, error) {
	return copyIEValue("ExtendedEmergNumList", i.Contents, 4, 65535)
}
