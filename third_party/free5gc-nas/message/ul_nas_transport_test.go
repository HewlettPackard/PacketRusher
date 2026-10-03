package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestULNASTransportUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x67, // EPD, SecHdr/Spare, Msg Type
				// PayloadCntrType  *ie.PayloadCntrType  //     V,     1/2B, 9.11.3.40
				0x01, //
				// PayloadCntr      *ie.PayloadCntr      //  LV-E, 3-65537B, 9.11.3.39
				0x00, 0x08,
				0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1,
				// PDUSessID        *ie.PDUSessId2       0x12    TV,       2B, 9.11.3.41
				0x12, 0x05,
				// ReqType          *ie.ReqType          0x80    TV,       1B, 9.11.3.47
				0x81,
				// DNN              *ie.DNN              0x25   TLV,   3-102B, 9.11.2.1B
				0x25, 0x09, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74,
				// OldPDUSessID     : &ie.PDUSessId2       0x59    TV,       2B, 9.11.3.41
				0x59, 0x02,
				// SNSSAI           : &ie.SNSSAI           0x22   TLV,    3-10B, 9.11.2.8
				0x22, 0x08,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10,
				// AdditionalInfo   : &ie.AdditionalInfo   0x24   TLV,     3-nB, 9.11.2.1
				// MAPDUSessInfo    : &ie.MAPDUSessInfo    0xA0    TV,       1B, 9.11.3.31A
				// RelAssistanceInd : &ie.RelAssistanceInd 0xF0    TV,       1B, 9.11.3.46A
			},
			expected: &ULNASTransport{
				PayloadCntrType: &ie.PayloadCntrType{
					Value: ie.PayloadCntrType_N1SMInfo,
				},
				PayloadCntr: &ie.PayloadCntr{
					Pct:      ie.PayloadCntrType_N1SMInfo,
					Contents: []byte{0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
				},
				PDUSessID: &ie.PDUSessId2{
					Value: 0x5,
				},
				ReqType: &ie.ReqType{
					Value: ie.ReqType_InitialReq,
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
				OldPDUSessID: &ie.PDUSessId2{
					Value: 0x2,
				},
				SNSSAI: &ie.SNSSAI{
					SST:            0x12,
					SD:             "345678",
					MappedHPLMNSST: 0x90,
					MappedHPLMNSD:  "543210",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(ULNASTransport)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestULNASTransportMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x67, // EPD, SecHdr/Spare, Msg Type
				// PayloadCntrType  *ie.PayloadCntrType  //     V,     1/2B, 9.11.3.40
				0x01, //
				// PayloadCntr      *ie.PayloadCntr      //  LV-E, 3-65537B, 9.11.3.39
				0x00, 0x08,
				0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1,
				// PDUSessID        *ie.PDUSessId2       0x12    TV,       2B, 9.11.3.41
				0x12, 0x05,
				// OldPDUSessID     : &ie.PDUSessId2       0x59    TV,       2B, 9.11.3.41
				0x59, 0x02,
				// ReqType          *ie.ReqType          0x80    TV,       1B, 9.11.3.47
				0x81,
				// SNSSAI           : &ie.SNSSAI           0x22   TLV,    3-10B, 9.11.2.8
				0x22, 0x08,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x54, 0x32, 0x10,
				// DNN              *ie.DNN              0x25   TLV,   3-102B, 9.11.2.1B
				0x25, 0x09, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74,
				// AdditionalInfo   : &ie.AdditionalInfo   0x24   TLV,     3-nB, 9.11.2.1
				// MAPDUSessInfo    : &ie.MAPDUSessInfo    0xA0    TV,       1B, 9.11.3.31A
				// RelAssistanceInd : &ie.RelAssistanceInd 0xF0    TV,       1B, 9.11.3.46A
			},
			input: &ULNASTransport{
				PayloadCntrType: &ie.PayloadCntrType{
					Value: ie.PayloadCntrType_N1SMInfo,
				},
				PayloadCntr: &ie.PayloadCntr{
					Pct:      ie.PayloadCntrType_N1SMInfo,
					Contents: []byte{0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
				},
				PDUSessID: &ie.PDUSessId2{
					Value: 0x5,
				},
				ReqType: &ie.ReqType{
					Value: ie.ReqType_InitialReq,
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
				OldPDUSessID: &ie.PDUSessId2{
					Value: 0x2,
				},
				SNSSAI: &ie.SNSSAI{
					SST:            0x12,
					SD:             "345678",
					MappedHPLMNSST: 0x90,
					MappedHPLMNSD:  "543210",
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
