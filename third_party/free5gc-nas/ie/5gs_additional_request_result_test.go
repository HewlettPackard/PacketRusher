package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdditionalReqResult5GS_UnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected *AdditionalReqResult5GS
	}{
		{
			name:     "Valid input - 1",
			input:    []byte{0x71},
			expected: &AdditionalReqResult5GS{PRD: PagingRestriction_Accepted},
		},
		{
			name:     "Valid input - 2 ",
			input:    []byte{0x12},
			expected: &AdditionalReqResult5GS{PRD: PagingRestriction_Rejected},
		},
		{
			name:     "Valid input - 3",
			input:    []byte{0x22},
			expected: &AdditionalReqResult5GS{PRD: PagingRestriction_Rejected},
		},
		{
			name:     "Invalid input",
			input:    []byte{0x32, 0x01},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var i AdditionalReqResult5GS
			err := i.UnmarshalBinary(tc.input)
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &i)
			}
		})
	}
}

func TestAdditionalReqResult5GS_MarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    *AdditionalReqResult5GS
		expected []byte
	}{
		{
			name:     "PRD = 0",
			input:    &AdditionalReqResult5GS{PRD: UEReqType_Reserve},
			expected: []byte{0x00},
		},
		{
			name:     "PRD = 1",
			input:    &AdditionalReqResult5GS{PRD: PagingRestriction_Accepted},
			expected: []byte{0x01},
		},
		{
			name:     "PRD = 2",
			input:    &AdditionalReqResult5GS{PRD: PagingRestriction_Rejected},
			expected: []byte{0x02},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}
