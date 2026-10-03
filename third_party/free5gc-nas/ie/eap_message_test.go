package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEAPMsgUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *EAPMsg
	}{
		{
			name:  "Data length 4",
			input: []byte{0x01, 0x02, 0x03, 0x04},
			expected: &EAPMsg{
				Eap: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
		{
			name:     "Data length 1",
			input:    []byte{0x01},
			expected: nil,
		},
		{
			name:     "Data length 1501",
			input:    make([]byte, 1501),
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(EAPMsg)
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

func TestEAPMsgMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *EAPMsg
		expected []byte
	}{
		{
			name: "Data length 4",
			input: &EAPMsg{
				Eap: []byte{0x01, 0x02, 0x03, 0x04},
			},
			expected: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name: "Data length 1",
			input: &EAPMsg{
				Eap: []byte{0x01},
			},
			expected: nil,
		},
		{
			name: "Data length 1501",
			input: &EAPMsg{
				Eap: make([]byte, 1501),
			},
			expected: nil,
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
