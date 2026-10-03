package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestParsePlain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name:  "Parse plain PDUSessEstReq",
			input: []byte{0x2E, 0x05, 0x01, 0xC1, 0xFF, 0xFF, 0x91, 0xA1},
			expected: &PDUSessEstReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xFF,
					Downlink: 0xFF,
				},
				PDUSessType: &ie.PDUSessType{
					Value: 1,
				},
				SSCMode: &ie.SSCMode{
					Mode: 1,
				},
			},
		},
		{
			name: "Parse plain AuthReq",
			input: []byte{
				0x7e, 0x00, 0x56, 0x00, 0x02, 0x00, 0x00, 0x21, 0x34, 0xfb, 0x88, 0xaf, 0x54, 0xa9, 0x1a, 0x58,
				0x29, 0xcd, 0x90, 0xc0, 0xa8, 0xfa, 0xd4, 0xda, 0x20, 0x10, 0x0b, 0x32, 0x64, 0x66, 0xfb, 0xc0,
				0x80, 0x00, 0x84, 0x9e, 0x28, 0x49, 0x5f, 0xb1, 0x71, 0xa1,
			},
			expected: &AuthReq{
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 0,
				},
				ABBA: &ie.ABBA{
					Abba: []byte{0x00, 0x00},
				},
				AuthParamRAND5GAuthChlg: &ie.AuthParamRAND{
					Rand: []byte{
						0x34, 0xfb, 0x88, 0xaf, 0x54, 0xa9, 0x1a, 0x58,
						0x29, 0xcd, 0x90, 0xc0, 0xa8, 0xfa, 0xd4, 0xda,
					},
				},
				AuthParamAUTN5GAuthChlg: &ie.AuthParamAUTN{
					Autn: []byte{
						0x0b, 0x32, 0x64, 0x66, 0xfb, 0xc0, 0x80, 0x00,
						0x84, 0x9e, 0x28, 0x49, 0x5f, 0xb1, 0x71, 0xa1,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := parsePlainMsg(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestParseNIA0WithInvalidMac(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			// Should pass since NIA0 means no check for MAC
			name: "Parse NIA0 RegComplete with arbitrary MAC",
			input: []byte{
				0x7e, 0x02, // intefrity protected (NIA0) and ciphered (NEA0)
				// arbitrary MAC
				0x11, 0x22, 0x33, 0x44,
				0x01, 0x7e, 0x00, 0x43,
			},
			expected: &RegComplete{},
		},
	}

	for _, tc := range testCases {
		// fake a secCtx
		sc := NewSecCtx(
			CoreNetworkSide, Bearer3GPP, AlgCiphering128NEA0, AlgIntegrity128NIA0, nil, nil)
		t.Run(tc.name, func(t *testing.T) {
			msg, err := Parse(tc.input, sc)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestParseNEA0EncWithNoSecCtx(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Parse NEA0 encrypted PDUSessEstReq without security header",
			input: []byte{
				0x7e, 0x03, 0x01, 0x02, 0x03, 0x04, 0x00,
				0x2E, 0x05, 0x01, 0xC1, 0xFF, 0xFF, 0x91, 0xA1,
			},
			expected: &PDUSessEstReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xFF,
					Downlink: 0xFF,
				},
				PDUSessType: &ie.PDUSessType{
					Value: 1,
				},
				SSCMode: &ie.SSCMode{
					Mode: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := Parse(tc.input, nil)
			macErr := &Error{
				MACFailure: &MACFailure{
					Expected: []byte{0x00, 0x00, 0x00, 0x00},
					Received: []byte{0x01, 0x02, 0x03, 0x04},
				},
			}
			require.Equal(t, macErr, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestMarshalPlain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Marshal plain AuthReq",
			input: &AuthReq{
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 0,
				},
				ABBA: &ie.ABBA{
					Abba: []byte{0x00, 0x00},
				},
				AuthParamRAND5GAuthChlg: &ie.AuthParamRAND{
					Rand: []byte{
						0x34, 0xfb, 0x88, 0xaf, 0x54, 0xa9, 0x1a, 0x58,
						0x29, 0xcd, 0x90, 0xc0, 0xa8, 0xfa, 0xd4, 0xda,
					},
				},
				AuthParamAUTN5GAuthChlg: &ie.AuthParamAUTN{
					Autn: []byte{
						0x0b, 0x32, 0x64, 0x66, 0xfb, 0xc0, 0x80, 0x00,
						0x84, 0x9e, 0x28, 0x49, 0x5f, 0xb1, 0x71, 0xa1,
					},
				},
			},
			expected: []byte{
				0x7e, 0x00, 0x56, 0x00, 0x02, 0x00, 0x00, 0x21, 0x34, 0xfb, 0x88, 0xaf, 0x54, 0xa9, 0x1a, 0x58,
				0x29, 0xcd, 0x90, 0xc0, 0xa8, 0xfa, 0xd4, 0xda, 0x20, 0x10, 0x0b, 0x32, 0x64, 0x66, 0xfb, 0xc0,
				0x80, 0x00, 0x84, 0x9e, 0x28, 0x49, 0x5f, 0xb1, 0x71, 0xa1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := Marshal(tc.input, nil, SecHdrTypePlainNas)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}
