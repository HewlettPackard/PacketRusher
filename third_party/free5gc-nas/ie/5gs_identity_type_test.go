package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdType5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *IdType5GS
	}{
		{
			name:  "All other values are unused and shall be interpreted as SUCI, if received by the UE.",
			input: []byte{0x00},
			expected: &IdType5GS{
				IdType: IdType_5GS_SUCI,
			},
		},
		{
			name:  "SUCI",
			input: []byte{0x01},
			expected: &IdType5GS{
				IdType: IdType_5GS_SUCI,
			},
		},
		{
			name:  "5GGUTI",
			input: []byte{0x02},
			expected: &IdType5GS{
				IdType: IdType_5GS_GUTI,
			},
		},
		{
			name:  "IMEI",
			input: []byte{0x03},
			expected: &IdType5GS{
				IdType: IdType_5GS_IMEI,
			},
		},
		{
			name:  "5GSTMSI",
			input: []byte{0x04},
			expected: &IdType5GS{
				IdType: IdType_5GS_TMSI,
			},
		},
		{
			name:  "IMEISV",
			input: []byte{0x05},
			expected: &IdType5GS{
				IdType: IdType_5GS_IMEISV,
			},
		},
		{
			name:  "MACAddress",
			input: []byte{0x06},
			expected: &IdType5GS{
				IdType: IdType_5GS_MACAddress,
			},
		},
		{
			name:  "EUI64",
			input: []byte{0x07},
			expected: &IdType5GS{
				IdType: IdType_5GS_EUI64,
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(IdType5GS)
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

func TestIdType5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *IdType5GS
		expected []byte
	}{
		{
			name: "None",
			input: &IdType5GS{
				IdType: IdType_5GS_None,
			},
			expected: []byte{0x00},
		},
		{
			name: "SUCI",
			input: &IdType5GS{
				IdType: IdType_5GS_SUCI,
			},
			expected: []byte{0x01},
		},
		{
			name: "5GGUTI",
			input: &IdType5GS{
				IdType: IdType_5GS_GUTI,
			},
			expected: []byte{0x02},
		},
		{
			name: "IMEI",
			input: &IdType5GS{
				IdType: IdType_5GS_IMEI,
			},
			expected: []byte{0x03},
		},
		{
			name: "5GSTMSI",
			input: &IdType5GS{
				IdType: IdType_5GS_TMSI,
			},
			expected: []byte{0x04},
		},
		{
			name: "IMEISV",
			input: &IdType5GS{
				IdType: IdType_5GS_IMEISV,
			},
			expected: []byte{0x05},
		},
		{
			name: "MACAddress",
			input: &IdType5GS{
				IdType: IdType_5GS_MACAddress,
			},
			expected: []byte{0x06},
		},
		{
			name: "EUI64",
			input: &IdType5GS{
				IdType: IdType_5GS_EUI64,
			},
			expected: []byte{0x07},
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
