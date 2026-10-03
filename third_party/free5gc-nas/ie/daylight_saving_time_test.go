package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDSTUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *DST
	}{
		{
			name: "No DST",
			input: []byte{
				0x0,
			},
			expected: &DST{
				Value: uint8(NoAdjustment),
			},
		},
		{
			name: "DST +1",
			input: []byte{
				0x1,
			},
			expected: &DST{
				Value: uint8(HourAdjustment_1),
			},
		},
		{
			name: "DST +2",
			input: []byte{
				0x2,
			},
			expected: &DST{
				Value: uint8(HourAdjustment_2),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(DST)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestDSTMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *DST
		expected []byte
	}{
		{
			name: "No DST",
			expected: []byte{
				0x0,
			},
			input: &DST{
				Value: uint8(NoAdjustment),
			},
		},
		{
			name: "DST +1",
			expected: []byte{
				0x1,
			},
			input: &DST{
				Value: uint8(HourAdjustment_1),
			},
		},
		{
			name: "DST +2",
			expected: []byte{
				0x2,
			},
			input: &DST{
				Value: uint8(HourAdjustment_2),
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
