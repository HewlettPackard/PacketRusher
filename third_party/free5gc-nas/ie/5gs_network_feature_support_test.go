package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNwFeatureSupport5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NwFeatureSupport5GS
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &NwFeatureSupport5GS{
				Length:       1,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: false,
				N3Data:       false,
				CPCiot5G:     false,
				RestrictEC:   0x00,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       false,
				LCS5G:        false,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x76, 0x54},
			expected: &NwFeatureSupport5GS{
				Length:       2,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: true,
				N3Data:       false,
				CPCiot5G:     true,
				RestrictEC:   0x01,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       false,
				LCS5G:        false,
			},
		},
		{
			name:  "Positive Case 3",
			input: []byte{0x76, 0x54, 0x32},
			expected: &NwFeatureSupport5GS{
				Length:       3,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: true,
				N3Data:       false,
				CPCiot5G:     true,
				RestrictEC:   0x01,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       true,
				LCS5G:        false,
				PIV:          true,
				RPR:          true,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54, 0x32, 0x10},
		},
		{
			name:  "Negative Case 2",
			input: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NwFeatureSupport5GS)
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

func TestNwFeatureSupport5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NwFeatureSupport5GS
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &NwFeatureSupport5GS{
				Length:       1,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: true,
				N3Data:       false,
				CPCiot5G:     true,
				RestrictEC:   0x01,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       true,
				LCS5G:        false,
			},
			expected: []byte{0x76},
		},
		{
			name: "Positive Case 2",
			input: &NwFeatureSupport5GS{
				Length:       2,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: true,
				N3Data:       false,
				CPCiot5G:     true,
				RestrictEC:   0x01,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       true,
				LCS5G:        false,
			},
			expected: []byte{0x76, 0x54},
		},
		{
			name: "Positive Case 3",
			input: &NwFeatureSupport5GS{
				Length:       3,
				MPSI:         false,
				IWKN26:       true,
				EMF:          0x03,
				EMC:          0x01,
				IMSVoPSN3GPP: true,
				IMSVoPS3GPP:  false,
				UPCiot5G:     false,
				IPHCCPCiot5G: true,
				N3Data:       false,
				CPCiot5G:     true,
				RestrictEC:   0x01,
				MCSI:         false,
				EMCN3:        false,
				EHCCPCiot5G:  false,
				ATSIND:       true,
				LCS5G:        false,
			},
			expected: []byte{0x76, 0x54, 0x02},
		},
		{
			name: "Negative Case 1",
			input: &NwFeatureSupport5GS{
				Length: 4,
			},
		},
		{
			name: "Negative Case 2",
			input: &NwFeatureSupport5GS{
				Length: 0,
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
