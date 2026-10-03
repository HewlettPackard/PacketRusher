package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapability5GMMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *Capability5GMM
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xaa},
			expected: &Capability5GMM{
				Length:       1,
				SGC:          true,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   true,
				LPP:          false,
				HOAttach:     true,
				S1Mode:       false,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0xaa, 0xaa},
			expected: &Capability5GMM{
				Length:       2,
				SGC:          true,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   true,
				LPP:          false,
				HOAttach:     true,
				S1Mode:       false,
				RACS:         true,
				NSSAA:        false,
				LCS5G:        true,
				V2XCNPC5:     false,
				V2XCEPC5:     true,
				V2X:          false,
				UPCiot5G:     true,
				SRVCC5G:      false,
			},
		},
		{
			name:  "Positive Case 3",
			input: []byte{0xaa, 0xaa, 0x0a},
			expected: &Capability5GMM{
				Length:       3,
				SGC:          true,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   true,
				LPP:          false,
				HOAttach:     true,
				S1Mode:       false,
				RACS:         true,
				NSSAA:        false,
				LCS5G:        true,
				V2XCNPC5:     false,
				V2XCEPC5:     true,
				V2X:          false,
				UPCiot5G:     true,
				SRVCC5G:      false,
				EHCCPCiot5G:  true,
				Multipleup:   false,
				WUSA:         true,
				CAG:          false,
			},
		},
		{
			name:  "Positive Case 4",
			input: []byte{0xaa, 0xaa, 0x0a, 0x00, 0x00},
			expected: &Capability5GMM{
				Length:       5,
				SGC:          true,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   true,
				LPP:          false,
				HOAttach:     true,
				S1Mode:       false,
				RACS:         true,
				NSSAA:        false,
				LCS5G:        true,
				V2XCNPC5:     false,
				V2XCEPC5:     true,
				V2X:          false,
				UPCiot5G:     true,
				SRVCC5G:      false,
				EHCCPCiot5G:  true,
				Multipleup:   false,
				WUSA:         true,
				CAG:          false,
			},
		},
		{
			name:  "Positive Case 5",
			input: []byte{0xaa, 0xaa, 0x0a, 0x00, 0x53, 0x55},
			expected: &Capability5GMM{
				Length:       6,
				SGC:          true,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   true,
				LPP:          false,
				HOAttach:     true,
				S1Mode:       false,
				RACS:         true,
				NSSAA:        false,
				LCS5G:        true,
				V2XCNPC5:     false,
				V2XCEPC5:     true,
				V2X:          false,
				UPCiot5G:     true,
				SRVCC5G:      false,
				EHCCPCiot5G:  true,
				Multipleup:   false,
				WUSA:         true,
				CAG:          false,
				UAS:          true,
				ExCAG:        true,
				MINT:         true,
				NSSRG:        true,
				RCMAP:        true,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{},
		},
		{
			name: "Negative Case 2",
			input: []byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(Capability5GMM)
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

func TestCapability5GMMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Capability5GMM
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &Capability5GMM{
				Length:       1,
				SGC:          false,
				IPHCCPCiot5G: true,
				N3Data:       true,
				CPCiot5G:     true,
				RestrictEC:   false,
				LPP:          true,
				HOAttach:     false,
				S1Mode:       true,
			},
			expected: []byte{0x55},
		},
		{
			name: "Positive Case 2",
			input: &Capability5GMM{
				Length:       2,
				SGC:          false,
				IPHCCPCiot5G: true,
				N3Data:       true,
				CPCiot5G:     true,
				RestrictEC:   false,
				LPP:          true,
				HOAttach:     false,
				S1Mode:       true,
				RACS:         false,
				NSSAA:        true,
				LCS5G:        false,
				V2XCNPC5:     true,
				V2XCEPC5:     false,
				V2X:          true,
				UPCiot5G:     false,
				SRVCC5G:      true,
			},
			expected: []byte{0x55, 0x55},
		},
		{
			name: "Positive Case 3",
			input: &Capability5GMM{
				Length:       3,
				SGC:          false,
				IPHCCPCiot5G: true,
				N3Data:       true,
				CPCiot5G:     true,
				RestrictEC:   false,
				LPP:          true,
				HOAttach:     false,
				S1Mode:       true,
				RACS:         false,
				NSSAA:        true,
				LCS5G:        false,
				V2XCNPC5:     true,
				V2XCEPC5:     false,
				V2X:          true,
				UPCiot5G:     false,
				SRVCC5G:      true,
				EHCCPCiot5G:  false,
				Multipleup:   true,
				WUSA:         false,
				CAG:          true,
			},
			expected: []byte{0x55, 0x55, 0x05},
		},
		{
			name: "Positive Case 4",
			input: &Capability5GMM{
				Length:       5,
				SGC:          false,
				IPHCCPCiot5G: true,
				N3Data:       true,
				CPCiot5G:     true,
				RestrictEC:   false,
				LPP:          true,
				HOAttach:     false,
				S1Mode:       true,
				RACS:         false,
				NSSAA:        true,
				LCS5G:        false,
				V2XCNPC5:     true,
				V2XCEPC5:     false,
				V2X:          true,
				UPCiot5G:     false,
				SRVCC5G:      true,
				EHCCPCiot5G:  false,
				Multipleup:   true,
				WUSA:         false,
				CAG:          true,
			},
			expected: []byte{0x55, 0x55, 0x05, 0x00, 0x00},
		},
		{
			name: "Positive Case 5",
			input: &Capability5GMM{
				Length:       6,
				SGC:          false,
				IPHCCPCiot5G: true,
				N3Data:       true,
				CPCiot5G:     true,
				RestrictEC:   false,
				LPP:          true,
				HOAttach:     false,
				S1Mode:       true,
				RACS:         false,
				NSSAA:        true,
				LCS5G:        false,
				V2XCNPC5:     true,
				V2XCEPC5:     false,
				V2X:          true,
				UPCiot5G:     false,
				SRVCC5G:      true,
				EHCCPCiot5G:  false,
				Multipleup:   true,
				WUSA:         false,
				CAG:          true,
				UAS:          true,
				ExCAG:        true,
				MINT:         true,
				NSSRG:        true,
				RCMAP:        true,
			},
			expected: []byte{0x55, 0x55, 0x05, 0x00, 0x53, 0x01},
		},
		{
			name: "Negative Case 1",
			input: &Capability5GMM{
				Length: 0,
			},
		},
		{
			name: "Negative Case 2",
			input: &Capability5GMM{
				Length: 14,
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
