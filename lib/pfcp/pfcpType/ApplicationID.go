package pfcpType

type ApplicationID struct {
	ApplicationIdentifier []byte
}

func (a *ApplicationID) MarshalBinary() (data []byte, err error) {
	// Octet 5 to (n+4)
	data = a.ApplicationIdentifier

	return data, nil
}

func (a *ApplicationID) UnmarshalBinary(data []byte) error {
	// Octet 5 to (n+4)
	a.ApplicationIdentifier = data

	return nil
}
