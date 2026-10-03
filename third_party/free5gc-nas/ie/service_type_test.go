package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSvcTypeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *SvcType
		expected string
	}{
		{
			name: "Signalling",
			input: &SvcType{
				Value: SvcType_Signalling,
			},
			expected: "Signaling",
		},
		{
			name: "Data",
			input: &SvcType{
				Value: SvcType_Data,
			},
			expected: "Data",
		},
		{
			name: "MobileTermSvc",
			input: &SvcType{
				Value: SvcType_MobileTermSvc,
			},
			expected: "MobileTermSvc",
		},
		{
			name: "EmergSvc",
			input: &SvcType{
				Value: SvcType_EmergSvc,
			},
			expected: "EmergSvc",
		},
		{
			name: "EmergSvcFlbk",
			input: &SvcType{
				Value: SvcType_EmergSvcFlbk,
			},
			expected: "EmergSvcFlbk",
		},
		{
			name: "HighPriorityAccess",
			input: &SvcType{
				Value: SvcType_HighPriorityAccess,
			},
			expected: "HighPriorityAccess",
		},
		{
			name: "ElevatedSignalling",
			input: &SvcType{
				Value: SvcType_ElevatedSignalling,
			},
			expected: "ElevatedSignalling",
		},
		{
			name: "Reserved01",
			input: &SvcType{
				Value: SvcType_Reserved01,
			},
			expected: "Reserved01 ",
		},
		{
			name: "Reserved02",
			input: &SvcType{
				Value: SvcType_Reserved02,
			},
			expected: "Reserved02 ",
		},
		{
			name: "Reserved03",
			input: &SvcType{
				Value: SvcType_Reserved03,
			},
			expected: "Reserved03 ",
		},
		{
			name: "Reserved04",
			input: &SvcType{
				Value: SvcType_Reserved04,
			},
			expected: "Reserved04 ",
		},
		{
			name:     "NilInput",
			input:    nil,
			expected: "",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			s := tc.input.String()
			require.Equal(t, tc.expected, s)
		})
	}
}

func TestSvcTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *SvcType
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xA0},
			expected: &SvcType{
				Value: SvcType_Signalling,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0xB1},
			expected: &SvcType{
				Value: SvcType_Data,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0xC2},
			expected: &SvcType{
				Value: SvcType_MobileTermSvc,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0xD3},
			expected: &SvcType{
				Value: SvcType_EmergSvc,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0xE4},
			expected: &SvcType{
				Value: SvcType_EmergSvcFlbk,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0xf5},
			expected: &SvcType{
				Value: SvcType_HighPriorityAccess,
			},
		},
		{
			name:  "Positive Case 1",
			input: []byte{0x06},
			expected: &SvcType{
				Value: SvcType_ElevatedSignalling,
			},
		},
		{
			name:  "Positive Case unused 1",
			input: []byte{0x27},
			expected: &SvcType{
				Value: SvcType_Signalling,
			},
		},
		{
			name:  "Positive Case unused 2",
			input: []byte{0x38},
			expected: &SvcType{
				Value: SvcType_Signalling,
			},
		},
		{
			name:  "Positive Case unused 3",
			input: []byte{0x49},
			expected: &SvcType{
				Value: SvcType_Data,
			},
		},
		{
			name:  "Positive Case unused 4",
			input: []byte{0x5A},
			expected: &SvcType{
				Value: SvcType_Data,
			},
		},
		{
			name:  "Positive Case unused 5",
			input: []byte{0x6B},
			expected: &SvcType{
				Value: SvcType_Data,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54, 0x32},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(SvcType)
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

func TestSvcTypeMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *SvcType
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x03},
			input: &SvcType{
				Value: 3,
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
