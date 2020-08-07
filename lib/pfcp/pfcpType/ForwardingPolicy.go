package pfcpType

import (
	"fmt"
)

type ForwardingPolicy struct {
	ForwardingPolicyIdentifierLength uint8
	ForwardingPolicyIdentifier       []byte
}

func (f *ForwardingPolicy) MarshalBinary() (data []byte, err error) {
	// Octet 5
	data = append([]byte(""), byte(f.ForwardingPolicyIdentifierLength))

	// Octet 6 to (6+a)
	if len(f.ForwardingPolicyIdentifier) != int(f.ForwardingPolicyIdentifierLength) {
		return []byte(""), fmt.Errorf("Unmatch length of forwarding policy identifier: Expect %d, got %d", f.ForwardingPolicyIdentifierLength, len(f.ForwardingPolicyIdentifier))
	}
	data = append(data, f.ForwardingPolicyIdentifier...)

	return data, nil
}

func (f *ForwardingPolicy) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	f.ForwardingPolicyIdentifierLength = uint8(data[idx])
	idx = idx + 1

	// Octet 6 to (6+a)
	if length < idx+uint16(f.ForwardingPolicyIdentifierLength) {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	f.ForwardingPolicyIdentifier = data[idx : idx+uint16(f.ForwardingPolicyIdentifierLength)]
	idx = idx + uint16(f.ForwardingPolicyIdentifierLength)

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
