package ie

// SORTransparentCntr preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type SORTransparentCntr struct {
	Contents []byte
}

func (i *SORTransparentCntr) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("SORTransparentCntr", b, 17, 65535)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *SORTransparentCntr) MarshalBinary() ([]byte, error) {
	return copyIEValue("SORTransparentCntr", i.Contents, 17, 65535)
}
