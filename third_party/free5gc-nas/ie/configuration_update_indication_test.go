package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCfgUpdateIndUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *CfgUpdateInd
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xa2},
			expected: &CfgUpdateInd{
				RED: true,
				ACK: false,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(CfgUpdateInd)
			err := ie.UnmarshalBinary(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCfgUpdateIndMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *CfgUpdateInd
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x02},
			input: &CfgUpdateInd{
				RED: true,
				ACK: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}
