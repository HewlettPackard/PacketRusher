package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayloadCntrTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PayloadCntrType
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x81},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_N1SMInfo,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x02},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_SMS,
			},
		},
		{
			name:  "Positive Case 3",
			input: []byte{0x03},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_LPP,
			},
		},
		{
			name:  "Positive Case 4",
			input: []byte{0x04},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_SOR,
			},
		},
		{
			name:  "Positive Case 5",
			input: []byte{0x05},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_UEPolicy,
			},
		},
		{
			name:  "Positive Case 6",
			input: []byte{0x06},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_UEParamsUpdate,
			},
		},
		{
			name:  "Positive Case 7",
			input: []byte{0x07},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_LocationSvcMsg,
			},
		},
		{
			name:  "Positive Case 8",
			input: []byte{0x08},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_CiotUserData,
			},
		},
		{
			name:  "Positive Case 9",
			input: []byte{0x09},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_SvcLvlAACntr,
			},
		},
		{
			name:  "Positive Case 10",
			input: []byte{0x0a},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_EventNotif,
			},
		},
		{
			name:  "Positive Case f",
			input: []byte{0x0f},
			expected: &PayloadCntrType{
				Value: PayloadCntrType_MultiplePayloads,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PayloadCntrType)
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

func TestPayloadCntrTypeMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PayloadCntrType
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &PayloadCntrType{
				Value: PayloadCntrType_N1SMInfo,
			},
			expected: []byte{0x01},
		},
		{
			name: "Positive Case 2",
			input: &PayloadCntrType{
				Value: PayloadCntrType_SMS,
			},
			expected: []byte{0x02},
		},
		{
			name: "Positive Case 3",
			input: &PayloadCntrType{
				Value: PayloadCntrType_LPP,
			},
			expected: []byte{0x03},
		},
		{
			name: "Positive Case 4",
			input: &PayloadCntrType{
				Value: PayloadCntrType_SOR,
			},
			expected: []byte{0x04},
		},
		{
			name: "Positive Case 5",
			input: &PayloadCntrType{
				Value: PayloadCntrType_UEPolicy,
			},
			expected: []byte{0x05},
		},
		{
			name: "Positive Case 6",
			input: &PayloadCntrType{
				Value: PayloadCntrType_UEParamsUpdate,
			},
			expected: []byte{0x06},
		},
		{
			name: "Positive Case 7",
			input: &PayloadCntrType{
				Value: PayloadCntrType_LocationSvcMsg,
			},
			expected: []byte{0x07},
		},
		{
			name: "Positive Case 8",
			input: &PayloadCntrType{
				Value: PayloadCntrType_CiotUserData,
			},
			expected: []byte{0x08},
		},
		{
			name: "Positive Case 9",
			input: &PayloadCntrType{
				Value: PayloadCntrType_SvcLvlAACntr,
			},
			expected: []byte{0x09},
		},
		{
			name: "Positive Case 10",
			input: &PayloadCntrType{
				Value: PayloadCntrType_EventNotif,
			},
			expected: []byte{0x0a},
		},
		{
			name: "Positive Case f",
			input: &PayloadCntrType{
				Value: PayloadCntrType_MultiplePayloads,
			},
			expected: []byte{0x0f},
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
