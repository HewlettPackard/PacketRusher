package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapability5GSMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *Capability5GSM
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xaa, 0x33},
			expected: &Capability5GSM{
				Rqos:    false,
				MH6PDU:  true,
				EPTS1:   false,
				ATSSSST: 5,
				TPMIC:   true,
				APMQF:   true,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(Capability5GSM)
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

func TestCapability5GSMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Capability5GSM
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &Capability5GSM{
				Rqos:    false,
				MH6PDU:  true,
				EPTS1:   false,
				ATSSSST: 5,
				TPMIC:   true,
				APMQF:   true,
			},
			expected: []byte{0xaa, 0x01},
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
