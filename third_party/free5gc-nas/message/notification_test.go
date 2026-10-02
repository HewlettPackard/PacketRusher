package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestNotifUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name:  "Notif",
			input: []byte{0x7e, 0x00, 0x65, 0x00},
			expected: &Notif{
				AccessType: &ie.AccessType{Value: 0},
			},
		},
		{
			name:  "Notif",
			input: []byte{0x7e, 0x00, 0x65, 0x01},
			expected: &Notif{
				AccessType: &ie.AccessType{Value: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(Notif)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestNotifMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name:     "Notif",
			expected: []byte{0x7e, 0x00, 0x65, 0x00},
			input: &Notif{
				AccessType: &ie.AccessType{Value: 0},
			},
		},
		{
			name:     "Notif",
			expected: []byte{0x7e, 0x00, 0x65, 0x01},
			input: &Notif{
				AccessType: &ie.AccessType{Value: 1},
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
