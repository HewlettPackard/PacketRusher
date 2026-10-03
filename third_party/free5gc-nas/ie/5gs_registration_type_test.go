package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegType5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *RegType5GS
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x74},
			expected: &RegType5GS{
				FOR_Pending: false,
				Value:       4,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(RegType5GS)
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

func TestRegType5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *RegType5GS
		expected []byte
	}{
		{
			name: "Case 1",
			input: &RegType5GS{
				FOR_Pending: false,
				Value:       4,
			},
			expected: []byte{0x04},
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
