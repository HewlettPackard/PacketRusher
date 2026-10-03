package ie

// MappedEPSBearerCtxs preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type MappedEPSBearerCtxs struct {
	Contents []byte
}

func (i *MappedEPSBearerCtxs) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("MappedEPSBearerCtxs", b, 4, 65535)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *MappedEPSBearerCtxs) MarshalBinary() ([]byte, error) {
	return copyIEValue("MappedEPSBearerCtxs", i.Contents, 4, 65535)
}
