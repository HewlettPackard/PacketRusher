package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestDeregReqUEOrigUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x45,
				// DeregType   *ie.DeregType        V,     1/2B, 9.11.3.20
				// Ngksi       *ie.NASKeySetId      V,     1/2B, 9.11.3.32
				0x19,
				// MobileId5GS *ie.MobileId5GS   LV-E,     6-nB, 9.11.3.4
				0x00, 0x0b,
				0xf2, 0x64, 0xf6, 0x66, 0xa8, 0x01, 0xb0, 0x00, 0x00, 0x00, 0x01,
			},
			expected: &DeregReqUEOrig{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 1,
				},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId:   ie.IdType_5GS_GUTI,
					AllOneBits: 0x0F,
					PlmnId: ie.PlmnId{
						MCC: "466",
						MNC: "66",
					},
					AMFRegionID: 168,
					AMFSetID:    6,
					AMFPointer:  48,
					AMFId:       "a801b0",
					TMSI5G:      [4]byte{0x00, 0x00, 0x00, 0x01},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(DeregReqUEOrig)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestDeregReqUEOrigMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x45,
				// DeregType   *ie.DeregType        V,     1/2B, 9.11.3.20
				// Ngksi       *ie.NASKeySetId      V,     1/2B, 9.11.3.32
				0x19,
				// MobileId5GS *ie.MobileId5GS   LV-E,     6-nB, 9.11.3.4
				0x00, 0x0b,
				0xf2, 0x64, 0xf6, 0x66, 0xa8, 0x01, 0xb0, 0x00, 0x00, 0x00, 0x01,
			},
			input: &DeregReqUEOrig{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 1,
				},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId:   ie.IdType_5GS_GUTI,
					AllOneBits: 0x0F,
					PlmnId: ie.PlmnId{
						MCC: "466",
						MNC: "66",
					},
					AMFRegionID: 168,
					AMFSetID:    6,
					AMFPointer:  48,
					TMSI5G:      [4]byte{0x00, 0x00, 0x00, 0x01},
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
