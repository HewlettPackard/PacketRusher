package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDSTTEthPortMACAddrUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *DSTTEthPortMACAddr
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			expected: &DSTTEthPortMACAddr{
				MacAddr: [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			},
		},
		{
			name:  "Negative Case bad length",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(DSTTEthPortMACAddr)
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

func TestDSTTEthPortMACAddrMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *DSTTEthPortMACAddr
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			input: &DSTTEthPortMACAddr{
				MacAddr: [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
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
