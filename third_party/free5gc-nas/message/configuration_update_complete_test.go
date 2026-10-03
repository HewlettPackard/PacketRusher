package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCfgUpdateCompleteUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       []byte
		expected    Message
		expectedErr bool
	}{
		{
			name: "Configuration Update Complete",
			input: []byte{
				0x7e, 0x00, 0x44, // EPD, SecHdr/Spare, Msg Type
			},
			expected:    &CfgUpdateComplete{},
			expectedErr: false,
		},
		{
			name: "Invalid Configuration Update Complete",
			input: []byte{
				0x7e, 0x00, // Missing Msg Type
			},
			expected:    &CfgUpdateComplete{},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(CfgUpdateComplete)
			err := msg.UnmarshalBinary(tc.input)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, msg)
			}
		})
	}
}

func TestCfgUpdateCompleteMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       Message
		expected    []byte
		expectedErr bool
	}{
		{
			name:  "Configuration Update Complete",
			input: &CfgUpdateComplete{},
			expected: []byte{
				0x7e, 0x00, 0x55, // EPD, SecHdr/Spare, Msg Type
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			}
		})
	}
}
