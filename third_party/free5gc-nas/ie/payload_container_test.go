package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayloadCntrUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		pct      uint8
		expected *PayloadCntr
	}{
		{
			name:  "case 1",
			input: []byte{0x00, 0x08, 0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
			pct:   PayloadCntrType_N1SMInfo,
			expected: &PayloadCntr{
				Pct:      PayloadCntrType_N1SMInfo,
				Contents: []byte{0x00, 0x08, 0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
			},
		},
		// Negative case: length > 65535, so omitted here.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PayloadCntr)
			err := ie.UnmarshalBinary(tc.input, tc.pct)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPayloadCntrMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PayloadCntr
		expected []byte
	}{
		{
			name:     "Case 1",
			expected: []byte{0x00, 0x08, 0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
			input: &PayloadCntr{
				Contents: []byte{0x00, 0x08, 0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
			},
		},
		{
			name:     "Data length 1",
			expected: []byte{0x01},
			input: &PayloadCntr{
				Contents: []byte{0x01},
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
