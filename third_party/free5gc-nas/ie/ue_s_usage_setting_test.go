package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUesUsageSettingUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *UesUsageSetting
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &UesUsageSetting{
				UESUsageSetting: uint8(UsageType_VoiceCentric),
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(UesUsageSetting)
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

func TestUesUsageSettingMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *UesUsageSetting
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &UesUsageSetting{
				UESUsageSetting: uint8(UsageType_VoiceCentric),
			},
			expected: []byte{0x00},
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
