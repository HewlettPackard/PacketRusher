package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestSvcReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x4c, // EPD, SecHdr/Spare, Msg Type
				0x05, // ngKSI / Service Type V
				// ie.MobileId5GS, LV, 9B 9.11.3.4
				0x00, 0x07, 0xf4, 0x01, 0xb0, 0x00, 0x00, 0x00, 0x91,
				// ie.PDUSessStatus, TLV, 9.11.3.44
				0x50, 0x02, 0x02, 0x00,
				// UplinkDataStatus     : &ie.UplinkDataStatus     0x40   TLV,    4-34B, 9.11.3.57
				0x40, 0x02, 0x76, 0x84,
				// AllowedPDUSessStatus : &ie.AllowedPDUSessStatus 0x25   TLV,    4-34B, 9.11.3.13
				0x25, 0x02, 0x76, 0x84,
				// NASMsgCntr           : &ie.NASMsgCntr           0x71 TLV-E,     4-nB, 9.11.3.33
				0x71, 0x00, 0x01, 0x43,
			},
			expected: &SvcReq{
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 5,
				},
				SvcType: &ie.SvcType{
					Value: ie.SvcType_Signalling,
				},
				TMSI5GS: &ie.MobileId5GS{
					TypeOfId:   ie.IdType_5GS_TMSI,
					AllOneBits: 0x0F,
					AMFSetID:   6,
					AMFPointer: 48,
					TMSI5G:     [4]byte{0x00, 0x00, 0x00, 0x91},
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				UplinkDataStatus: &ie.UplinkDataStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				AllowedPDUSessStatus: &ie.AllowedPDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				NASMsgCntr: &ie.NASMsgCntr{
					Contents: []byte{0x43},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(SvcReq)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestSvcReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x4c, // EPD, SecHdr/Spare, Msg Type
				0x05, // ngKSI / Service Type V
				// ie.MobileId5GS, LV, 9B 9.11.3.4
				0x00, 0x07, 0xf4, 0x01, 0xb0, 0x00, 0x00, 0x00, 0x91,
				// UplinkDataStatus     : &ie.UplinkDataStatus     0x40   TLV,    4-34B, 9.11.3.57
				0x40, 0x02, 0x76, 0x84,
				// ie.PDUSessStatus, TLV, 9.11.3.44
				0x50, 0x02, 0x02, 0x00,
				// AllowedPDUSessStatus : &ie.AllowedPDUSessStatus 0x25   TLV,    4-34B, 9.11.3.13
				0x25, 0x02, 0x76, 0x84,
				// NASMsgCntr           : &ie.NASMsgCntr           0x71 TLV-E,     4-nB, 9.11.3.33
				0x71, 0x00, 0x01, 0x43,
			},
			input: &SvcReq{
				Ngksi: &ie.NASKeySetId{
					Tsc: ie.SecCtxTypeNative,
					Ksi: 5,
				},
				SvcType: &ie.SvcType{
					Value: ie.SvcType_Signalling,
				},
				TMSI5GS: &ie.MobileId5GS{
					TypeOfId:   ie.IdType_5GS_TMSI,
					AllOneBits: 0x0F,
					AMFSetID:   6,
					AMFPointer: 48,
					TMSI5G:     [4]byte{0x00, 0x00, 0x00, 0x91},
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				UplinkDataStatus: &ie.UplinkDataStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				AllowedPDUSessStatus: &ie.AllowedPDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				NASMsgCntr: &ie.NASMsgCntr{
					Contents: []byte{0x43},
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
