package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectedNSSAIUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *RejectedNSSAI
	}{
		{
			name: "Case sst only",
			expected: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 0, sst: 1, sd: ""},
				},
			},
			input: []byte{0x10, 0x01},
		},
		{
			name: "Case 1 set",
			expected: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 0, sst: 1, sd: "010203"},
				},
			},
			input: []byte{0x40, 0x01, 0x01, 0x02, 0x03},
		},
		{
			name: "Case 2 sets",
			expected: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 1, sst: 16, sd: ""},
					{cause: 2, sst: 8, sd: ""},
				},
			},
			input: []byte{
				0x11, 0x10,
				0x12, 0x08,
			},
		},
		{
			name: "Case 3 sets",
			expected: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 1, sst: 16, sd: ""},
					{cause: 0, sst: 1, sd: "010203"},
					{cause: 2, sst: 255, sd: "010203"},
				},
			},
			input: []byte{
				0x11, 0x10,
				0x40, 0x01, 0x01, 0x02, 0x03,
				0x42, 0xff, 0x01, 0x02, 0x03,
			},
		},
		{
			name:  "Negative Case 1 - invalid length",
			input: []byte{0x76, 0x54, 0x32},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(RejectedNSSAI)
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

func TestRejectedNSSAIMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *RejectedNSSAI
		expected []byte
	}{
		{
			name: "Case sst only",
			input: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 0, sst: 1, sd: ""},
				},
			},
			expected: []byte{0x10, 0x01},
		},
		{
			name: "Case 1 set",
			input: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 0, sst: 1, sd: "010203"},
				},
			},
			expected: []byte{0x40, 0x01, 0x01, 0x02, 0x03},
		},
		{
			name: "Case 2 sets",
			input: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 1, sst: 16, sd: ""},
					{cause: 2, sst: 8, sd: ""},
				},
			},
			expected: []byte{
				0x11, 0x10,
				0x12, 0x08,
			},
		},
		{
			name: "Case 3 sets",
			input: &RejectedNSSAI{
				RejSnssai: []rejSNSSAI{
					{cause: 1, sst: 16, sd: ""},
					{cause: 0, sst: 1, sd: "010203"},
					{cause: 2, sst: 255, sd: "010203"},
				},
			},
			expected: []byte{
				0x11, 0x10,
				0x40, 0x01, 0x01, 0x02, 0x03,
				0x42, 0xff, 0x01, 0x02, 0x03,
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
