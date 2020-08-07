package pfcpType

import (
	"fmt"
)

// Acceptance in a response
const (
	CauseRequestAccepted uint8 = 1
)

// Rejection in a response
const (
	CauseRequestRejected uint8 = iota + 64
	CauseSessionContextNotFound
	CauseMandatoryIeMissing
	CauseConditionalIeMissing
	CauseInvalidLength
	CauseMandatoryIeIncorrect
	CauseInvalidForwardingPolicy
	CauseInvalidFTeidAllocationOption
	CauseNoEstablishedPfcpAssociation
	CauseRuleCreationModificationFailure
	CausePfcpEntityInCongestion
	CauseNoResourcesAvailable
	CauseServiceNotSupported
	CauseSystemFailure
)

type Cause struct {
	CauseValue uint8
}

func (c *Cause) MarshalBinary() (data []byte, err error) {
	// Octet 5
	data = append([]byte(""), byte(c.CauseValue))

	return data, nil
}

func (c *Cause) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	c.CauseValue = uint8(data[idx])
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
