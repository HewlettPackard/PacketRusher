package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNwNameUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NwName
	}{
		{
			name: "empty string",
			input: []byte{
				0x80,
			},
			expected: &NwName{
				Ext:          1, // Ext always be 1
				CodeScheme:   CodeScheme_Default,
				AddCi:        0,
				NumSpareBits: 0,
				TextStr:      "",
			},
		},
		{
			name: "hellohello",
			input: []byte{
				0x82, 0xe8, 0x32, 0x9b, 0xfd, 0x46, 0x97, 0xd9, 0xec, 0x37,
			},
			expected: &NwName{
				Ext:          1, // Ext always be 1
				CodeScheme:   CodeScheme_Default,
				AddCi:        0,
				NumSpareBits: 2,
				TextStr:      "hellohello",
			},
		},
		{
			name: "12345678",
			input: []byte{
				0x80,
				0x31, 0xd9, 0x8c, 0x56, 0xb3, 0xdd, 0x70,
			},
			expected: &NwName{
				Ext:          1, // Ext always be 1
				CodeScheme:   CodeScheme_Default,
				AddCi:        0,
				NumSpareBits: 0,
				TextStr:      "12345678",
			},
		},
		{
			name: "1234567890abcdefghijklmnopqrstuzwxyx",
			input: []byte{
				0x84,
				0x31, 0xd9, 0x8c, 0x56, 0xb3, 0xdd, 0x70, 0x39, 0x58,
				0x58, 0x3c, 0x26, 0x97, 0xcd, 0x67, 0x74, 0x5a, 0xbd,
				0x66, 0xb7, 0xdd, 0x6f, 0x78, 0x5c, 0x3e, 0xa7, 0xd7,
				0xf5, 0x77, 0x7c, 0x1e, 0x0f,
			},
			expected: &NwName{
				Ext:          1, // Ext always be 1
				CodeScheme:   CodeScheme_Default,
				AddCi:        0,
				NumSpareBits: 4,
				TextStr:      "1234567890abcdefghijklmnopqrstuzwxyx",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NwName)
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

func TestNwNameMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NwName
		expected []byte
	}{
		{
			name: "empty string",
			input: &NwName{
				Ext:        1, // Ext always be 1
				CodeScheme: CodeScheme_Default,
				AddCi:      0,
				// NumSpareBits: 0, // calculated on MarshalBinary()
				TextStr: "",
			},
			expected: []byte{
				0x80,
			},
		},
		{
			name: "hellohello",
			input: &NwName{
				Ext:        1, // Ext always be 1
				CodeScheme: CodeScheme_Default,
				AddCi:      0,
				// NumSpareBits: 2, // calculated on MarshalBinary()
				TextStr: "hellohello",
			},
			expected: []byte{
				0x82, 0xe8, 0x32, 0x9b, 0xfd, 0x46, 0x97, 0xd9, 0xec, 0x37,
			},
		},
		{
			name: "12345678",
			input: &NwName{
				Ext:        1, // Ext always be 1
				CodeScheme: CodeScheme_Default,
				AddCi:      0,
				// NumSpareBits: 0, // calculated on MarshalBinary()
				TextStr: "12345678",
			},
			expected: []byte{
				0x80,
				0x31, 0xd9, 0x8c, 0x56, 0xb3, 0xdd, 0x70,
			},
		},
		{
			name: "1234567890abcdefghijklmnopqrstuzwxyx",
			input: &NwName{
				Ext:        1, // Ext always be 1
				CodeScheme: CodeScheme_Default,
				AddCi:      0,
				// NumSpareBits: 4, // calculated on MarshalBinary()
				TextStr: "1234567890abcdefghijklmnopqrstuzwxyx",
			},
			expected: []byte{
				0x84,
				0x31, 0xd9, 0x8c, 0x56, 0xb3, 0xdd, 0x70, 0x39, 0x58,
				0x58, 0x3c, 0x26, 0x97, 0xcd, 0x67, 0x74, 0x5a, 0xbd,
				0x66, 0xb7, 0xdd, 0x6f, 0x78, 0x5c, 0x3e, 0xa7, 0xd7,
				0xf5, 0x77, 0x7c, 0x1e, 0x0f,
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
