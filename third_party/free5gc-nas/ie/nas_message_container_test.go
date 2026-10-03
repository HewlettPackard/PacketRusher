package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNASMsgCntrUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NASMsgCntr
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x43},
			expected: &NASMsgCntr{
				Contents: []byte{0x43},
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			expected: &NASMsgCntr{
				Contents: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			},
		},
		{
			name:  "NegativeCase 1",
			input: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NASMsgCntr)
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

func TestNASMsgCntrMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NASMsgCntr
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x72},
			input: &NASMsgCntr{
				Contents: []byte{0x72},
			},
		},
		{
			name:     "Positive Case 2",
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			input: &NASMsgCntr{
				Contents: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
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
