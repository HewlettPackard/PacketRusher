package ie

// EPSBearerCtxStatus preserves an unconsumed standard IE value. See README.packetrusher.md.
// Contents excludes the IEI and length octets.
type EPSBearerCtxStatus struct {
	Contents []byte
}

func (i *EPSBearerCtxStatus) UnmarshalBinary(b []byte) error {
	value, err := copyIEValue("EPSBearerCtxStatus", b, 2, 2)
	if err != nil {
		return err
	}
	i.Contents = value
	return nil
}

func (i *EPSBearerCtxStatus) MarshalBinary() ([]byte, error) {
	return copyIEValue("EPSBearerCtxStatus", i.Contents, 2, 2)
}
