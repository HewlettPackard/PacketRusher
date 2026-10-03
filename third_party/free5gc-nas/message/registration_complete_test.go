package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegCompleteUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name:     "5GS RegComplete",
			input:    []byte{0x7e, 0x00, 0x43},
			expected: &RegComplete{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(RegComplete)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestRegCompleteMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name:     "5GS RegComplete",
			input:    &RegComplete{},
			expected: []byte{0x7e, 0x00, 0x43},
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
