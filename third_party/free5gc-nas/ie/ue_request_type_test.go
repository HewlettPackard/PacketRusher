package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUEReqTypeUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected *UEReqType
	}{
		{
			name:     "Valid input - 1",
			input:    []byte{0x01},
			expected: &UEReqType{ReqType: UEReqType_NasConnRelease},
		},
		{
			name:     "Valid input - 2 ",
			input:    []byte{0x32},
			expected: &UEReqType{ReqType: UEReqType_PagingRej},
		},
		{
			name:     "Valid input - 3",
			input:    []byte{0x82},
			expected: &UEReqType{ReqType: UEReqType_PagingRej},
		},
		{
			name:     "Invalid input",
			input:    []byte{0x32, 0x01},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var i UEReqType
			err := i.UnmarshalBinary(tc.input)
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &i)
			}
		})
	}
}

func TestUEReqTypeMarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    *UEReqType
		expected []byte
	}{
		{
			name:     "ReqType = 0",
			input:    &UEReqType{ReqType: UEReqType_Reserve},
			expected: []byte{0x00},
		},
		{
			name:     "ReqType = 1",
			input:    &UEReqType{ReqType: UEReqType_NasConnRelease},
			expected: []byte{0x01},
		},
		{
			name:     "ReqType = 2",
			input:    &UEReqType{ReqType: UEReqType_PagingRej},
			expected: []byte{0x02},
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
