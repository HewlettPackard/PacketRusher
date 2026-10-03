package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNwSlicingIndUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NwSlicingInd
	}{
		{
			name:  "Case 1",
			input: []byte{0x76},
			expected: &NwSlicingInd{
				DCNI:  true,
				NSSCI: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NwSlicingInd)
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

func TestNwSlicingIndMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NwSlicingInd
		expected []byte
	}{
		{
			name: "Case 1",
			input: &NwSlicingInd{
				DCNI:  true,
				NSSCI: false,
			},
			expected: []byte{0x02},
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
