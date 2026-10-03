package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestSecModeCmdUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "5GS SecModeCmd",
			input: []byte{
				0x7e, 0x00, 0x5d, // EPD, SecHdr/Spare, Msg Type
				// SelectedNASSecAlgos *ie.NASSecAlgos V, 1B, 9.11.3.34
				0x02,
				// Ngksi *ie.NASKeySetId  V, 1/2B, 9.11.3.32
				0x00,
				// ReplayedUESecCapabilities   *ie.UESecCapability LV, 3-9B, 9.11.3.54
				0x04, 0x80, 0x20, 0xe0, 0xe0,
				// IMEISVReq *ie.IMEISVReq 0xE0  TV, 1B, 9.11.3.28
				0xe1,
				// Additional5GSecInfo *ie.Additional5GSecInfo 0x36 TLV, 3B, 9.11.3.12
				0x36, 0x01, 0x00,
				// SelectedEPSNASSecAlgos       *ie.EPSNASSecAlgos       0x57    TV,       2B, 9.11.3.25
				// EAPMsg                      *ie.EAPMsg              0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ABBA                        *ie.ABBA                0x38   TLV,     4-nB, 9.11.3.10
				0x38, 0x02, 0x00, 0x00,
				// ReplayedS1UESecCapabilities *ie.S1UESecCapability   0x19   TLV,     4-7B, 9.11.3.48A
			},
			expected: &SecModeCmd{
				SelectedNASSecAlgos: &ie.NASSecAlgos{
					CipheringAlgo: ie.EncAlgo_5GEA0,
					MsgIntAlgo:    ie.IntegrityAlgo_1285GIA2,
				},
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 0,
				},
				ReplayedUESecCapabilities: &ie.UESecCapability{
					Length: 4,
					EA05G:  true, EA1_128_5G: false, EA2_128_5G: false, EA3_128_5G: false,
					EA45G: false, EA55G: false, EA65G: false, EA75G: false,
					IA05G: false, IA1_128_5G: false, IA2_128_5G: true, IA3_128_5G: false,
					IA45G: false, IA55G: false, IA65G: false, IA75G: false,
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
				},
				IMEISVReq: &ie.IMEISVReq{
					Value: 1,
				},
				Additional5GSecInfo: &ie.Additional5GSecInfo{
					RINMR: false,
					HDP:   false,
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{0x01, 0x02, 0x03, 0x04},
				},
				ABBA: &ie.ABBA{
					Abba: []byte{0x00, 0x00},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(SecModeCmd)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestSecModeCmdMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name:     "5GS SecModeCmd",
			expected: []byte{0x7e, 0x00, 0x5d, 0x02, 0x00, 0x04, 0x80, 0x20, 0xe0, 0xe0, 0xe1, 0x36, 0x01, 0x00},
			input: &SecModeCmd{
				SelectedNASSecAlgos: &ie.NASSecAlgos{
					CipheringAlgo: ie.EncAlgo_5GEA0,
					MsgIntAlgo:    ie.IntegrityAlgo_1285GIA2,
				},
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 0,
				},
				ReplayedUESecCapabilities: &ie.UESecCapability{
					Length: 4,
					EA05G:  true, EA1_128_5G: false, EA2_128_5G: false, EA3_128_5G: false,
					EA45G: false, EA55G: false, EA65G: false, EA75G: false,
					IA05G: false, IA1_128_5G: false, IA2_128_5G: true, IA3_128_5G: false,
					IA45G: false, IA55G: false, IA65G: false, IA75G: false,
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
				},
				IMEISVReq: &ie.IMEISVReq{
					Value: 1,
				},
				Additional5GSecInfo: &ie.Additional5GSecInfo{
					RINMR: false,
					HDP:   false,
				},
			},
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
