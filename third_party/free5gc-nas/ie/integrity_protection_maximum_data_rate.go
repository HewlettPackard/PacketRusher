package ie

// IntegrityProtectionMaxDataRate is detailed in
// 9.11.4.7 Integrity protection maximum data rate, 24.501
type IntegrityProtectionMaxDataRate struct {
	Uplink   uint8
	Downlink uint8
}

func (i *IntegrityProtectionMaxDataRate) UnmarshalBinary(b []byte) error {
	i.Uplink = b[0]
	i.Downlink = b[1]
	return nil
}

func (i *IntegrityProtectionMaxDataRate) MarshalBinary() ([]byte, error) {
	return []byte{i.Uplink, i.Downlink}, nil
}
