package ie

import "github.com/pkg/errors"

// Additional5GSecInfo is detailed in 9.11.3.12 Additional 5G security information, 24.501
type Additional5GSecInfo struct {
	// Name, uint8, Bits, Octet
	RINMR bool // 2 -> 2 , 3 -> 3 Retransmission of initial NAS message requested
	HDP   bool // 1 -> 1 , 3 -> 3 Horizontal derivation parameter
}

const (
	// HDP
	KamfDerivationNotRequired bool = false
	KamfDerivationRequired    bool = true

	// RINMR
	ReXmitOfInitNasMsgNotRequested bool = false
	ReXmitOfInitNasMsgRequested    bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Additional5GSecInfo) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The Additional5GSecInfo  IE length(%d) is incorrect", len(b))
	}
	i.RINMR = GetBit2(b[0]) == 1
	i.HDP = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Additional5GSecInfo) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit2(b[0], bool2uint8(i.RINMR))
	b[0] = SetBit1(b[0], bool2uint8(i.HDP))

	return b, nil
}
