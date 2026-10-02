package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNSSAIUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NSSAI
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x08, 0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
			expected: &NSSAI{
				SNSSAIs: []SNSSAI{
					{
						SST:            0x12,
						SD:             "345678",
						MappedHPLMNSST: 0x90,
						MappedHPLMNSD:  "543210",
					},
				},
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0x01, 0x12, 0x08, 0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
			expected: &NSSAI{
				SNSSAIs: []SNSSAI{
					{
						SST: 0x12,
					},
					{
						SST:            0x12,
						SD:             "345678",
						MappedHPLMNSST: 0x90,
						MappedHPLMNSD:  "543210",
					},
				},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x02, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
		{
			name:  "Negative Case 2",
			input: []byte{0x32, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NSSAI)
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

func TestNSSAIMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NSSAI
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x08, 0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
			input: &NSSAI{
				SNSSAIs: []SNSSAI{
					{
						SST:            0x12,
						SD:             "345678",
						MappedHPLMNSST: 0x90,
						MappedHPLMNSD:  "543210",
					},
				},
			},
		},
		{
			name:     "Positive Case 2",
			expected: []byte{0x01, 0x12, 0x08, 0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10},
			input: &NSSAI{
				SNSSAIs: []SNSSAI{
					{
						SST: 0x12,
					},
					{
						SST:            0x12,
						SD:             "345678",
						MappedHPLMNSST: 0x90,
						MappedHPLMNSD:  "543210",
					},
				},
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
