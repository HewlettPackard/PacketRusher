package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPDUSessReactivationResultUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PDUSessReactivationResult
	}{
		{
			name:  "Case 1",
			input: []byte{0x76, 0x84},
			expected: &PDUSessReactivationResult{
				Psi: Psi{PSI: [16]bool{
					false, true, true, false, true, true, true, false,
					false, false, true, false, false, false, false, true,
				}},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PDUSessReactivationResult)
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

func TestPDUSessReactivationResultMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PDUSessReactivationResult
		expected []byte
	}{
		{
			name: "Case 1",
			input: &PDUSessReactivationResult{
				Psi: Psi{PSI: [16]bool{
					false, true, true, false, true, true, true, false,
					false, false, true, false, false, false, false, true,
				}},
			},
			expected: []byte{0x76, 0x84},
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
