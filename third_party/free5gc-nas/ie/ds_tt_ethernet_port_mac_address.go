package ie

import (
	"github.com/pkg/errors"
)

// DSTTEthPortMACAddr is detailed in 9.11.4.25 DS-TT Ethernet port MAC address, 24.501
type DSTTEthPortMACAddr struct {
	// Name, uint8, Bits, Octet
	MacAddr [6]byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DSTTEthPortMACAddr) UnmarshalBinary(b []byte) error {
	if len(b) != 6 {
		return errors.Errorf("Bad DSTTEthPortMACAddr IE length(%d)", len(b))
	}
	copy(i.MacAddr[:], b)

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DSTTEthPortMACAddr) MarshalBinary() ([]byte, error) {
	b := make([]byte, 6)
	copy(b, i.MacAddr[:])

	return b, nil
}
