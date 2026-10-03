package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxNumOfSupportedPktFiltersUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *MaxNumOfSupportedPktFilters
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x71},
			expected: &MaxNumOfSupportedPktFilters{
				MaxNumOfSupportedPktFilters: 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(MaxNumOfSupportedPktFilters)
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

func TestMaxNumOfSupportedPktFiltersMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *MaxNumOfSupportedPktFilters
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &MaxNumOfSupportedPktFilters{
				MaxNumOfSupportedPktFilters: 1,
			},
			expected: []byte{0x20},
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
