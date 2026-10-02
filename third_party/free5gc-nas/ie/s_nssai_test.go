package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSNSSAIUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *SNSSAI
	}{
		{
			name:  "SST",
			input: []byte{0x12},
			expected: &SNSSAI{
				SST: 0x12,
			},
		},
		{
			name:  "SST MappedHPLMNSST",
			input: []byte{0x12, 0x10},
			expected: &SNSSAI{
				SST:            0x12,
				MappedHPLMNSST: 0x10,
			},
		},
		{
			name:  "SST SD",
			input: []byte{0x12, 0x11, 0x22, 0x33},
			expected: &SNSSAI{
				SST: 0x12,
				SD:  "112233",
			},
		},
		{
			name:  "SST SD MappedHPLMNSST",
			input: []byte{0x12, 0x11, 0x22, 0x33, 0x10},
			expected: &SNSSAI{
				SST:            0x12,
				SD:             "112233",
				MappedHPLMNSST: 0x10,
			},
		},
		{
			name:  "SST SD MappedHPLMNSST MappedHPLMNSD",
			input: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
			expected: &SNSSAI{
				SST:            0x12,
				SD:             "345678",
				MappedHPLMNSST: 0x90,
				MappedHPLMNSD:  "543210",
			},
		},
		{
			name:  "ERROR - length too long",
			input: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(SNSSAI)
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

func TestSNSSAIMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *SNSSAI
		expected []byte
	}{
		{
			name: "SST",
			input: &SNSSAI{
				SST: 0x12,
			},
			expected: []byte{0x12},
		},
		{
			name: "SST MappedHPLMNSST",
			input: &SNSSAI{
				SST:            0x12,
				MappedHPLMNSST: 0x10,
			},
			expected: []byte{0x12, 0x10},
		},
		{
			name: "SST SD",
			input: &SNSSAI{
				SST: 0x12,
				SD:  "112233",
			},
			expected: []byte{0x12, 0x11, 0x22, 0x33},
		},
		{
			name: "SST SD MappedHPLMNSST",
			input: &SNSSAI{
				SST:            0x12,
				SD:             "112233",
				MappedHPLMNSST: 0x10,
			},
			expected: []byte{0x12, 0x11, 0x22, 0x33, 0x10},
		},
		{
			name: "SST SD MappedHPLMNSST MappedHPLMNSD",
			input: &SNSSAI{
				SST:            0x12,
				SD:             "345678",
				MappedHPLMNSST: 0x90,
				MappedHPLMNSD:  "543210",
			},
			expected: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
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
