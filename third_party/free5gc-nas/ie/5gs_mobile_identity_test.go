package ie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMobileId5GS_UnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *MobileId5GS
	}{
		{
			name:  "Positive Case Type TMSI",
			input: []byte{0xf4, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: &MobileId5GS{
				TypeOfId:   IdType_5GS_TMSI,
				AllOneBits: 0x0F,
				AMFSetID:   1016,
				AMFPointer: 0,
				TMSI5G:     [4]byte{0x00, 0x00, 0x00, 0x01},
			},
		},
		{
			name:  "Negative Case Type TMSI - buffer is 1 octet less",
			input: []byte{0xf4, 0xfe, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:  "Negative Case Type TMSI - buffer is 1 octet more",
			input: []byte{0xf4, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01},
		},
		{
			name:  "Positive Case Type GUTI",
			input: []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: &MobileId5GS{
				TypeOfId:   IdType_5GS_GUTI,
				AllOneBits: 0x0F,
				PlmnId: PlmnId{
					MCC: "310",
					MNC: "103",
				},
				AMFRegionID: 202,
				AMFSetID:    1016,
				AMFPointer:  0,
				AMFId:       "cafe00",
				TMSI5G:      [4]byte{0x00, 0x00, 0x00, 0x01},
			},
		},
		{
			name:  "Negative Case Type GUTI - buffer 1 octet less",
			input: []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:  "Negative Case Type GUTI - buffer too long (1 octet more)",
			input: []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02},
		},
		{
			name:  "Positive Case Type IMEI(Even)",
			input: []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0xF4},
			expected: &MobileId5GS{
				TypeOfId:     IdType_5GS_IMEI,
				OddEvenIndic: EvenNumOfIdDigit,
				IMEI:         [15]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 0},
			},
		},
		{
			name:  "Negative Case Type IMEI(Even) buffer 1 octet less",
			input: []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32},
		},
		{
			name:  "Negative Case Type IMEI(Even) buffer 1 octet more",
			input: []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0xF4, 0x22},
		},
		{
			name:  "Positive Case Type IMEI(Odd)",
			input: []byte{0x1B, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54},
			expected: &MobileId5GS{
				TypeOfId:     IdType_5GS_IMEI,
				OddEvenIndic: OddNumOfIdDigit,
				IMEI:         [15]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5},
			},
		},
		{
			name:  "Negative Case Type IMEISV - buffer 1 byte less",
			input: []byte{0x15, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54},
		},
		{
			name:  "Negative Case Type IMEISV - buffer 1 byte more",
			input: []byte{0x15, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54, 0x66, 0xf6},
		},
		{
			name:  "Positive Case Type IMEISV",
			input: []byte{0x15, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54, 0xF6},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_IMEISV,
				IMEISV:   [16]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			},
		},
		{
			name:  "Negative Case Type MAC Address - buffer 1 byte less",
			input: []byte{0x06, 0x01, 0x02, 0x03, 0x04, 0x05},
		},
		{
			name:  "Negative Case Type MAC Address - buffer 1 byte more",
			input: []byte{0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
		{
			name:  "Positive Case Type MAC Address",
			input: []byte{0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_MACAddress,
				MAURI:    0,
				MACAddr:  [6]uint8{1, 2, 3, 4, 5, 6},
			},
		},
		{
			name:  "Negative Case Type EUI64 - buffer 1 byte less",
			input: []byte{0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
		{
			name:  "Negative Case Type EUI64 - buffer 1 byte more",
			input: []byte{0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
		},
		{
			name:  "Positive Case Type EUI64",
			input: []byte{0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_EUI64,
				EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
		{
			name:  "Positive Case Type SUCI - SUPI==IMSI - Null scheme",
			input: []byte{0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34, 0x12, 0x21},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "214",
					MNC: "653",
				},
				RoutingIndDigit:    [4]uint8{8, 7, 0, 9},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x12,
				MSINLength:         6,
				MSINDigits:         [10]uint8{4, 3, 2, 1, 1, 2, 0, 0, 0, 0},
			},
		},
		{
			name:  "Positive Case Type SUCI - SUPI==IMSI - Null scheme - MSINLength=9 (Odd Length with Padding)",
			input: []byte{0x1, 0x99, 0x29, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf2},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "999",
					MNC: "002",
				},
				SUPIFormat:         uint8(SupiIMSI),
				RoutingIndDigit:    [4]uint8{0, 0, 0, 0},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x00,
				MSINLength:         9,
				MSINDigits:         [10]uint8{0, 0, 0, 0, 0, 0, 0, 0, 2, 0xf},
			},
		},
		{
			name:  "PPositive Case Type SUCI - SUPI==IMSI - Null scheme - MSINLength=10",
			input: []byte{0x01, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x00, 0x00, 0x10, 0x32, 0x54, 0x76, 0x98},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "208",
					MNC: "93",
				},
				SUPIFormat:         uint8(SupiIMSI),
				RoutingIndDigit:    [4]uint8{0, 0, 0, 0},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x00,
				MSINLength:         10,
				MSINDigits:         [10]uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI-ECIESSchemeProfileA",
			input: []byte{
				0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x01, 0x12,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71, 0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef, 0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a,
			},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "214",
					MNC: "653",
				},
				RoutingIndDigit:    [4]uint8{8, 7, 0, 9},
				ProtectionSchemeId: 1,
				HomeNwPubKeyId:     0x12,
				ECCEphPubKey: []byte{
					0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71, 0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
					0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef, 0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				},
				CipherVal: []byte{0xb8, 0x6f, 0xb0, 0xfe, 0x25},
				MACTag:    []byte{0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a},
			},
		},
		{
			name:  "Positive Case Type SUCI - SUPI==GCI",
			input: []byte{0x21, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			expected: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiGCI),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
		},
		{
			name:  "Positive Case Type SUCI - SUPI==GLI",
			input: []byte{0x31, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			expected: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiGLI),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
		},
		{
			name:  "Positive Case Type SUCI - SUPI==Network Specific Identifier",
			input: []byte{0x11, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			expected: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiNwSpecificIdentifier),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
		},
		{
			name:  "Positive Case Type None",
			input: []byte{0x00},
			expected: &MobileId5GS{
				TypeOfId: IdType_5GS_None,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(MobileId5GS)
			err := ie.UnmarshalBinary(tc.input)
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			}
		})
	}
}

func TestMobileId5GS_MarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *MobileId5GS
		expected []byte
	}{
		{
			name: "Positive Case Type 5GS TMSI",
			input: &MobileId5GS{
				TypeOfId:   IdType_5GS_TMSI,
				AllOneBits: 0x0F,
				AMFSetID:   1016,
				AMFPointer: 0,
				TMSI5G:     [4]byte{0x00, 0x00, 0x00, 0x01},
			},
			expected: []byte{0xf4, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "Positive Case Type 5G-GUTI",
			input: &MobileId5GS{
				TypeOfId:   IdType_5GS_GUTI,
				AllOneBits: 0x0F,
				PlmnId: PlmnId{
					MCC: "310",
					MNC: "103",
				},
				AMFRegionID: 202,
				AMFSetID:    1016,
				AMFPointer:  0,
				AMFId:       "cafe00",
				TMSI5G:      [4]byte{0x00, 0x00, 0x00, 0x01},
			},
			expected: []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "Positive Case Type 5GS IMEI(even)",
			input: &MobileId5GS{
				TypeOfId:     IdType_5GS_IMEI,
				OddEvenIndic: EvenNumOfIdDigit,
				IMEI:         [15]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 0},
			},
			expected: []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0xF4},
		},
		{
			name: "Positive Case Type 5GS IMEI(odd)",
			input: &MobileId5GS{
				TypeOfId:     IdType_5GS_IMEI,
				OddEvenIndic: OddNumOfIdDigit,
				IMEI:         [15]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5},
			},
			expected: []byte{0x1b, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54},
		},
		{
			name: "Positive Case Type 5GS IMEISV",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_IMEISV,
				IMEISV:   [16]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			},
			expected: []byte{0x15, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54, 0xf6},
		},
		{
			name: "Positive Case Type MAC Address",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_MACAddress,
				MAURI:    0,
				MACAddr:  [6]uint8{1, 2, 3, 4, 5, 6},
			},
			expected: []byte{0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		},
		{
			name: "Positive Case Type EUI64",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_EUI64,
				EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
			},
			expected: []byte{0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		},
		{
			name: "Positive Case Type None",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_None,
			},
			expected: []byte{0x00},
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI - Null scheme",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "214",
					MNC: "653",
				},
				SUPIFormat:         uint8(SupiIMSI),
				RoutingIndDigit:    [4]uint8{8, 7, 0, 9},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x12,
				MSINLength:         6,
				MSINDigits:         [10]uint8{4, 3, 2, 1, 5, 6, 0, 0, 0, 0},
			},
			expected: []byte{0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34, 0x12, 0x65},
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI - Null scheme - MSINLength=9",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "999",
					MNC: "002",
				},
				SUPIFormat:         uint8(SupiIMSI),
				RoutingIndDigit:    [4]uint8{0, 0, 0, 0},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x00,
				MSINLength:         9,
				MSINDigits:         [10]uint8{0, 0, 0, 0, 0, 0, 0, 0, 2, 0xf},
			},
			expected: []byte{0x1, 0x99, 0x29, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf2},
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI - Null scheme - MSINLength=10",
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "208",
					MNC: "93",
				},
				SUPIFormat:         uint8(SupiIMSI),
				RoutingIndDigit:    [4]uint8{0, 0, 0, 0},
				ProtectionSchemeId: 0,
				HomeNwPubKeyId:     0x00,
				MSINLength:         10,
				MSINDigits:         [10]uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
			expected: []byte{0x01, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x00, 0x00, 0x10, 0x32, 0x54, 0x76, 0x98},
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI-ECIESSchemeProfileA",
			expected: []byte{
				0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x01, 0x12,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71, 0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef, 0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a,
			},
			input: &MobileId5GS{
				TypeOfId: IdType_5GS_SUCI,
				PlmnId: PlmnId{
					MCC: "214",
					MNC: "653",
				},
				RoutingIndDigit:    [4]uint8{8, 7, 0, 9},
				ProtectionSchemeId: 1,
				HomeNwPubKeyId:     0x12,
				ECCEphPubKey: []byte{
					0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71, 0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
					0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef, 0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				},
				CipherVal: []byte{0xb8, 0x6f, 0xb0, 0xfe, 0x25},
				MACTag:    []byte{0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a},
			},
		},
		{
			name: "Positive Case Type SUCI - SUPI==GCI",
			input: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiGCI),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
			expected: []byte{0x21, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
		},
		{
			name: "Positive Case Type SUCI - SUPI==GLI",
			input: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiGLI),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
			expected: []byte{0x31, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
		},
		{
			name: "Positive Case Type SUCI - SUPI==Network Specific Identifier",
			input: &MobileId5GS{
				TypeOfId:   IdType_5GS_SUCI,
				SUPIFormat: uint8(SupiNwSpecificIdentifier),
				SUCINAI:    []uint8{0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
			},
			expected: []byte{0x11, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34},
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

func TestMobileId5GS_IdTypeStr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    uint8
		expected string
	}{
		{
			name:     "case 1",
			input:    IdType_5GS_None,
			expected: "None",
		},
		{
			name:     "negative case 1",
			input:    IdType_5GS_EUI64 + 1,
			expected: "unknown ID type(8)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := IdType5GSStr(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestMobileId5GS_IdStr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "GUTI String",
			input:    []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: "310103cafe0000000001",
		},
		{
			name:     "PEIStr IMEI(Even)",
			input:    []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0xF4},
			expected: "imei-12345678901234",
		},
		{
			name:     "Positive Case Type 5G-S-TMSI",
			input:    []byte{0xf4, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: "fe0000000001",
		},
		{
			name:     "Positive Case Type SUCI - SUPI==IMSI - Null scheme",
			input:    []byte{0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x00, 0x12, 0x34, 0x12, 0x21},
			expected: "suci-0-214-653-8709-0-18-432112",
		},
		{
			name:     "SUCI, ProtectionSchemeId == Null, no routing ind",
			input:    []byte{0x01, 0x12, 0x34, 0x56, 0xff, 0xff, 0x00, 0x12, 0x34, 0x12, 0x21},
			expected: "suci-0-214-653-0-0-18-432112",
		},
		{
			name: "Positive Case Type SUCI - SUPI==IMSI-ECIESSchemeProfileA",
			input: []byte{
				0x01, 0x12, 0x34, 0x56, 0x78, 0x90, 0x01, 0x12,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71, 0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef, 0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a,
			},
			expected: "suci-0-214-653-8709-1-18-" +
				"2bc3c1baeff3cb7158ca0e3570359fdf9b11cce059a90cefefda6b9da17a5f0bb86fb0fe2529d72d14200a1d8a",
		},
		{
			name:     "Positive Case Type EUI64",
			input:    []byte{0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expected: "0102030405060708",
		},
		{
			name:     "Positive Case Type MAC Address",
			input:    []byte{0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			expected: "010203040506",
		},
		{
			name:     "PEIStr IMEI(Even)",
			input:    []byte{0x13, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0xF4},
			expected: "imei-12345678901234",
		},
		{
			name:     "PEIStr IMEI(Odd)",
			input:    []byte{0x1B, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54},
			expected: "imei-123456789012345",
		},
		{
			name:     "PEIStr IMEISV",
			input:    []byte{0x15, 0x32, 0x54, 0x76, 0x98, 0x10, 0x32, 0x54, 0xF6},
			expected: "imeisv-1234567890123456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(MobileId5GS)
			if err := ie.UnmarshalBinary(tc.input); err != nil {
				require.Equal(t, err, nil)
			}
			out := ie.IdStr()
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestMobileId5GS_FromGUTIStr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "GUTI String",
			input:    []byte{0xf2, 0x13, 0x30, 0x01, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: "310103cafe0000000001",
		},
		{
			name:     "GUTI String",
			input:    []byte{0xf2, 0x02, 0xf8, 0x39, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: "20893cafe0000000001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(MobileId5GS)
			if err := ie.UnmarshalBinary(tc.input); err != nil {
				require.Equal(t, err, nil)
			}
			out := ie.IdStr()
			require.Equal(t, tc.expected, out)

			ie2 := new(MobileId5GS)
			err := ie2.FromGUTIStr(tc.expected)

			fmt.Printf("ie2 %+v %v\n", ie2, err)
			fmt.Printf("ie %+v\n", ie)
			require.Equal(t, ie, ie2)
		})
	}
}
