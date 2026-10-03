package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestAuthRspUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "5GS AuthRsp",
			input: []byte{
				0x7e, 0x00, 0x57, // EPD, SecHdr/Spare, Msg Type
				// ie.AuthRspParam TLV
				0x2d, 0x10,
				0x9b, 0x56, 0x1f, 0x8c, 0x61, 0x72, 0xc3, 0xa8,
				0x8f, 0xa0, 0xc2, 0x41, 0xc2, 0x1e, 0xc1, 0x36,
				// ie.EAPMsg TLV-E
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			expected: &AuthRsp{
				AuthRspParam: &ie.AuthRspParam{
					Res: []byte{
						0x9b, 0x56, 0x1f, 0x8c, 0x61, 0x72, 0xc3, 0xa8,
						0x8f, 0xa0, 0xc2, 0x41, 0xc2, 0x1e, 0xc1, 0x36,
					},
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(AuthRsp)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestAuthRspMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "5GS AuthRsp",
			expected: []byte{
				0x7e, 0x00, 0x57, // EPD, SecHdr/Spare, Msg Type
				// ie.AuthRspParam TLV
				0x2d, 0x10,
				0x9b, 0x56, 0x1f, 0x8c, 0x61, 0x72, 0xc3, 0xa8,
				0x8f, 0xa0, 0xc2, 0x41, 0xc2, 0x1e, 0xc1, 0x36,
				// ie.EAPMsg TLV-E
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			input: &AuthRsp{
				AuthRspParam: &ie.AuthRspParam{
					Res: []byte{
						0x9b, 0x56, 0x1f, 0x8c, 0x61, 0x72, 0xc3, 0xa8,
						0x8f, 0xa0, 0xc2, 0x41, 0xc2, 0x1e, 0xc1, 0x36,
					},
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
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
