package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeregAcceptUEOrigUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x46,
			},
			expected: &DeregAcceptUEOrig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(DeregAcceptUEOrig)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestDeregAcceptUEOrigMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x46,
			},
			input: &DeregAcceptUEOrig{},
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
