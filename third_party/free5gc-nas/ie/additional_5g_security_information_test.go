package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdditional5GSecInfoUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *Additional5GSecInfo
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xa2},
			expected: &Additional5GSecInfo{
				RINMR: true,
				HDP:   false,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(Additional5GSecInfo)
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

func TestAdditional5GSecInfoMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Additional5GSecInfo
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x02},
			input: &Additional5GSecInfo{
				RINMR: true,
				HDP:   false,
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
