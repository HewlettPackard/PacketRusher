package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUplinkDataStatusUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *UplinkDataStatus
	}{
		{
			name:  "Case 1",
			input: []byte{0x76, 0x84},
			expected: &UplinkDataStatus{
				Psi: Psi{PSI: [16]bool{
					false, true, true, false, true, true, true, false,
					false, false, true, false, false, false, false, true,
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(UplinkDataStatus)
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

func TestUplinkDataStatusMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *UplinkDataStatus
		expected []byte
	}{
		{
			name: "Case 1",
			input: &UplinkDataStatus{
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
