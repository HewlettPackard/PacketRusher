package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestSvcRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x4d, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
				// ie.PDUSessStatus                     0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0xaa,
				// ie.GPRSTimer2                        0x5f   TLV,       3B, 9.11.2.4
				0x5f, 0x01, 0x2c,
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ie.GPRSTimer2                        0x6B   TLV,       3B, 9.11.2.4
				0x6B, 0x01, 0x2c,
				// ie.CAGInfoList                       0x75 TLV-E,     3-nB, 9.11.3.18A
			},
			expected: &SvcRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, true, false, true, false, true, false, true,
					}},
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
				T3448Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				// CAGInfoList                      : &ie.CAGInfoList{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(SvcRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestSvcRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x4d, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
				// ie.PDUSessStatus                     0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0xaa,
				// ie.GPRSTimer2                        0x5f   TLV,       3B, 9.11.2.4
				0x5f, 0x01, 0x2c,
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ie.GPRSTimer2                        0x6B   TLV,       3B, 9.11.2.4
				0x6B, 0x01, 0x2c,
				// ie.CAGInfoList                       0x75 TLV-E,     3-nB, 9.11.3.18A
			},
			input: &SvcRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, true, false, true, false, true, false, true,
					}},
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
				T3448Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				// CAGInfoList                      : &ie.CAGInfoList{},
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
