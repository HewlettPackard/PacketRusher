package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS1UENwCapabilityUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *S1UENwCapability
	}{
		{
			name:  "Positive Case - 2 mandatory octets only",
			input: []byte{0xAA, 0xCC},
			expected: &S1UENwCapability{
				// 0xaa
				EEA0:     true,
				EEA1_128: false,
				EEA2_128: true,
				EEA3_128: false,
				EEA4:     true,
				EEA5:     false,
				EEA6:     true,
				EEA7:     false,

				// 0xcc
				EIA0:     true,
				EIA1_128: true,
				EIA2_128: false,
				EIA3_128: false,
				EIA4:     true,
				EIA5:     true,
				EIA6:     false,
				EIA7:     false,
			},
		},
		{
			name:  "Positive Case - all",
			input: []byte{0xAA, 0xCC, 0xF0, 0x0f, 0x55, 0x33, 0xe7, 0x18},
			expected: &S1UENwCapability{
				// 0xaa
				EEA0:     true,
				EEA1_128: false,
				EEA2_128: true,
				EEA3_128: false,
				EEA4:     true,
				EEA5:     false,
				EEA6:     true,
				EEA7:     false,

				// 0xcc
				EIA0:     true,
				EIA1_128: true,
				EIA2_128: false,
				EIA3_128: false,
				EIA4:     true,
				EIA5:     true,
				EIA6:     false,
				EIA7:     false,

				// 0xf0
				UEA0: true,
				UEA1: true,
				UEA2: true,
				UEA3: true,
				UEA4: false,
				UEA5: false,
				UEA6: false,
				UEA7: false,

				// 0x0f
				UCS2: false,
				UIA1: false,
				UIA2: false,
				UIA3: false,
				UIA4: true,
				UIA5: true,
				UIA6: true,
				UIA7: true,

				// 0x55
				ProSe_dd: false,
				ProSe:    true,
				H245_ASH: false,
				ACC_CSFB: true,
				LPP:      false,
				LCS:      true,
				SRVCC_1x: false,
				NF:       true,

				// 0x33
				EPCO:       false,
				HC_CP_CIoT: false,
				ERw_oPDN:   true,
				S1U_Data:   true,
				UP_CIoT:    false,
				CP_CIoT:    false,
				ProseRelay: true,
				ProSe_dc:   true,

				// 0x e3
				Bearers15:   true,
				SGC:         true,
				N1mode:      true,
				DCNR:        false,
				CP_Backoff:  false,
				RestrictEC:  true,
				V2X_PC5:     true,
				MultipleDRB: true,

				// 0x18
				V2X_NR_PC5: true,
				UP_MT_EDT:  true,
				CP_MT_EDT:  false,
				WUSA:       false,
				RACS:       false,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(S1UENwCapability)
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

func TestS1UENwCapabilityMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *S1UENwCapability
		expected []byte
	}{
		{
			name:     "Positive Case 2 mandatory octets only",
			expected: []byte{0xAA, 0xCC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			input: &S1UENwCapability{
				// 0xaa
				EEA0:     true,
				EEA1_128: false,
				EEA2_128: true,
				EEA3_128: false,
				EEA4:     true,
				EEA5:     false,
				EEA6:     true,
				EEA7:     false,

				// 0xcc
				EIA0:     true,
				EIA1_128: true,
				EIA2_128: false,
				EIA3_128: false,
				EIA4:     true,
				EIA5:     true,
				EIA6:     false,
				EIA7:     false,
			},
		},
		{
			name:     "Positive Case 1",
			expected: []byte{0xAA, 0xCC, 0xF0, 0x0f, 0x55, 0x33, 0xe7, 0x18},
			input: &S1UENwCapability{
				// 0xaa
				EEA0:     true,
				EEA1_128: false,
				EEA2_128: true,
				EEA3_128: false,
				EEA4:     true,
				EEA5:     false,
				EEA6:     true,
				EEA7:     false,

				// 0xcc
				EIA0:     true,
				EIA1_128: true,
				EIA2_128: false,
				EIA3_128: false,
				EIA4:     true,
				EIA5:     true,
				EIA6:     false,
				EIA7:     false,

				// 0xf0
				UEA0: true,
				UEA1: true,
				UEA2: true,
				UEA3: true,
				UEA4: false,
				UEA5: false,
				UEA6: false,
				UEA7: false,

				// 0x0f
				UCS2: false,
				UIA1: false,
				UIA2: false,
				UIA3: false,
				UIA4: true,
				UIA5: true,
				UIA6: true,
				UIA7: true,

				// 0x55
				ProSe_dd: false,
				ProSe:    true,
				H245_ASH: false,
				ACC_CSFB: true,
				LPP:      false,
				LCS:      true,
				SRVCC_1x: false,
				NF:       true,

				// 0x33
				EPCO:       false,
				HC_CP_CIoT: false,
				ERw_oPDN:   true,
				S1U_Data:   true,
				UP_CIoT:    false,
				CP_CIoT:    false,
				ProseRelay: true,
				ProSe_dc:   true,

				// 0x e7
				Bearers15:   true,
				SGC:         true,
				N1mode:      true,
				DCNR:        false,
				CP_Backoff:  false,
				RestrictEC:  true,
				V2X_PC5:     true,
				MultipleDRB: true,

				// 0x18
				V2X_NR_PC5: true,
				UP_MT_EDT:  true,
				CP_MT_EDT:  false,
				WUSA:       false,
				RACS:       false,
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
