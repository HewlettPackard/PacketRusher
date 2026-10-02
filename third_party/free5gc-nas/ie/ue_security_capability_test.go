package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUESecCapabilityUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *UESecCapability
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x80, 0x20, 0xe0, 0xe0},
			expected: &UESecCapability{
				Length:     4,
				EA05G:      true,
				EA1_128_5G: false,
				EA2_128_5G: false,
				EA3_128_5G: false,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: false,
				IA2_128_5G: true,
				IA3_128_5G: false,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0xf0, 0x70},
			expected: &UESecCapability{
				Length:     2,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       false,
				EEA1_128:   false,
				EEA2_128:   false,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       false,
				EIA1_128:   false,
				EIA2_128:   false,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:  "Positive Case 3",
			input: []byte{0xf0, 0x70, 0xe0, 0xe0, 0x00, 0x00},
			expected: &UESecCapability{
				Length:     6,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:  "Positive Case 4",
			input: []byte{0xf0, 0x70, 0xe0, 0xe0, 0x00, 0x00, 0x00, 0x00},
			expected: &UESecCapability{
				Length:     8,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54, 0x32},
		},
		{
			name:  "Negative Case 2",
			input: []byte{0x76, 0x54, 0x32, 0x76, 0x54, 0x32, 0x76, 0x54, 0x32, 0x76, 0x54, 0x32},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(UESecCapability)
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

func TestUESecCapabilityMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *UESecCapability
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x80, 0x20, 0xe0, 0xe0},
			input: &UESecCapability{
				Length:     4,
				EA05G:      true,
				EA1_128_5G: false,
				EA2_128_5G: false,
				EA3_128_5G: false,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: false,
				IA2_128_5G: true,
				IA3_128_5G: false,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:     "Positive Case 2",
			expected: []byte{0xf0, 0x70},
			input: &UESecCapability{
				Length:     2,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
			},
		},
		{
			name:     "Positive Case 3",
			expected: []byte{0xf0, 0x70, 0xe0, 0xe0, 0x00, 0x00},
			input: &UESecCapability{
				Length:     6,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name:     "Positive Case 3",
			expected: []byte{0xf0, 0x70, 0xe0, 0xe0, 0x00, 0x00, 0x00, 0x00},
			input: &UESecCapability{
				Length:     8,
				EA05G:      true,
				EA1_128_5G: true,
				EA2_128_5G: true,
				EA3_128_5G: true,
				EA45G:      false,
				EA55G:      false,
				EA65G:      false,
				EA75G:      false,
				IA05G:      false,
				IA1_128_5G: true,
				IA2_128_5G: true,
				IA3_128_5G: true,
				IA45G:      false,
				IA55G:      false,
				IA65G:      false,
				IA75G:      false,
				EEA0:       true,
				EEA1_128:   true,
				EEA2_128:   true,
				EEA3_128:   false,
				EEA4:       false,
				EEA5:       false,
				EEA6:       false,
				EEA7:       false,
				EIA0:       true,
				EIA1_128:   true,
				EIA2_128:   true,
				EIA3_128:   false,
				EIA4:       false,
				EIA5:       false,
				EIA6:       false,
				EIA7:       false,
			},
		},
		{
			name: "Negative Case 1",
			input: &UESecCapability{
				Length: 1,
			},
		},
		{
			name: "Negative Case 2",
			input: &UESecCapability{
				Length: 7,
			},
		},
		{
			name: "Negative Case 3",
			input: &UESecCapability{
				Length: 9,
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
