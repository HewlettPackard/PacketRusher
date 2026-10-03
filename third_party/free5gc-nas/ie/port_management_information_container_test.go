package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPmicCntrUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PortMgmtInfoCntr
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x06, 0x00, 0x02, 0x00, 0x00},
			expected: &PortMgmtInfoCntr{
				PmicCntr: []byte{0x06, 0x00, 0x02, 0x00, 0x00},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PortMgmtInfoCntr)
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

func TestPmicCntrMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PortMgmtInfoCntr
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x06, 0x00, 0x02, 0x00, 0x00},
			input: &PortMgmtInfoCntr{
				PmicCntr: []byte{0x06, 0x00, 0x02, 0x00, 0x00},
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
