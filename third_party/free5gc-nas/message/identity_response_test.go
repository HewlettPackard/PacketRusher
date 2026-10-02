package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestIdRspUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x5c, 0x00, 0x0d, 0x01, 0x64, 0xf6, 0x66, 0xf0, 0xff, 0x00, 0x00, 0x16, 0x00, 0x00,
				0x10, 0x79,
			},
			expected: &IdRsp{
				MobileId: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_SUCI,
					PlmnId: ie.PlmnId{
						MCC: "466",
						MNC: "66",
					},
					RoutingIndDigit:    [4]byte{0, 15, 15, 15},
					ProtectionSchemeId: ie.NullScheme,
					HomeNwPubKeyId:     0x00,
					MSINLength:         10,
					MSINDigits:         [10]uint8{6, 1, 0, 0, 0, 0, 0, 1, 9, 7},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(IdRsp)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestIdRspMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x5c, 0x00, 0x0d, 0x01, 0x64, 0xf6, 0x66, 0xf0, 0xff, 0x00, 0x00, 0x16, 0x00, 0x00,
				0x10, 0x79,
			},
			input: &IdRsp{
				MobileId: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_SUCI,
					PlmnId: ie.PlmnId{
						MCC: "466",
						MNC: "66",
					},
					RoutingIndDigit:    [4]byte{0, 15, 15, 15},
					ProtectionSchemeId: ie.NullScheme,
					HomeNwPubKeyId:     0x00,
					MSINLength:         10,
					MSINDigits:         [10]uint8{6, 1, 0, 0, 0, 0, 0, 1, 9, 7},
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
