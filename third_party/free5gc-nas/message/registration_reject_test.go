package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestRegRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x44, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
				// ie.GPRSTimer2                        0x5F   TLV,       3B, 9.11.2.5
				0x5f, 0x01, 0x06,
				// ie.GPRSTimer2                        0x16   TLV,       3B, 9.11.2.4
				0x16, 0x01, 0x2c,
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// RejectedNSSAI : &ie.RejectedNSSAI 0x69   TLV,    4-42B, 9.11.3.46
				// CAGInfoList   : &ie.CAGInfoList   0x75 TLV-E,     3-nB, 9.11.3.18A
			},
			expected: &RegRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_2Secs,
					Value: 6,
				},
				T3502Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
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
			msg := new(RegRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestRegRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x44, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
				// ie.GPRSTimer2                        0x5F   TLV,       3B, 9.11.2.5
				0x5f, 0x01, 0x06,
				// ie.GPRSTimer2                        0x16   TLV,       3B, 9.11.2.4
				0x16, 0x01, 0x2c,
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// RejectedNSSAI : &ie.RejectedNSSAI 0x69   TLV,    4-42B, 9.11.3.46
				// CAGInfoList   : &ie.CAGInfoList   0x75 TLV-E,     3-nB, 9.11.3.18A
			},
			input: &RegRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_2Secs,
					Value: 6,
				},
				T3502Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
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
