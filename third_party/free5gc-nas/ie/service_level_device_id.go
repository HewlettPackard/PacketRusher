package ie

import "github.com/pkg/errors"

// SvcLvlDeviceID is detailed in 9.11.2.11 Service-level device ID, 24.501
type SvcLvlDeviceID struct {
	// Name, uint8, Bits, Octet
	deviceID string
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlDeviceID) UnmarshalBinary(b []byte) error {
	if len(b) < 1 || len(b) > 255 {
		return errors.Errorf("bad len(SvcLvlDeviceID) (%d)", len(b))
	}
	i.deviceID = string(b)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlDeviceID) MarshalBinary() ([]byte, error) {
	length := len(i.deviceID)
	if length < 1 || length > 255 {
		return nil, errors.Errorf("bad len(deviceID) (%d)", length)
	}

	b := make([]byte, length)
	copy(b, i.deviceID)

	return b, nil
}

func (i *SvcLvlDeviceID) SetDeviceID(deviceID string) error {
	length := len(deviceID)
	if length < 1 || length > 255 {
		return errors.Errorf("bad len(DeviceID) (%d), max=255", length)
	}
	i.deviceID = deviceID
	return nil
}

func (i *SvcLvlDeviceID) GetDeviceID() string {
	return i.deviceID
}
