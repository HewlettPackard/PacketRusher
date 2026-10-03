package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNNUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *DNN
	}{
		{
			name:  "Positive Case 1 ",
			input: []byte{0x05, 0x41, 0x42, 0x43, 0x2d, 0x31, 0x1, 0x32, 0x3, 0x63, 0x6f, 0x6d},
			expected: &DNN{
				Value: "abc-1.2.com",
			},
		},
		{
			name:  "Positive Case 2 ",
			input: []byte{0x00},
			expected: &DNN{
				Value: "",
			},
		},
		{
			name:  "Negative Case 1: Bad label length (exceed real buf len)",
			input: []byte{0xff, 0x41, 0x42, 0x43, 0x3, 0x63, 0x6f, 0x6d},
			expected: &DNN{
				Value: "",
			},
		},
		{
			name:  "Valid with space",
			input: []byte{0x4, 0x2d, 0x42, 0x43, 0x20},
			expected: &DNN{
				Value: "-bc ",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(DNN)
			err := ie.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, ie)
		})
	}
}

func TestDNNMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *DNN
		expected []byte
	}{
		{
			name:     "Positive Case 1 ",
			expected: []byte{0x05, 0x61, 0x62, 0x63, 0x2d, 0x31, 0x1, 0x32, 0x3, 0x63, 0x6f, 0x6d},
			input: &DNN{
				Value: "ABC-1.2.com",
			},
		},
		{
			name:     "Positive Case 2 ",
			expected: []byte{0x00},
			input: &DNN{
				Value: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}
