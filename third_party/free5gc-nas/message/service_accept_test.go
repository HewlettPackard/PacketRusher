package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestSvcAcceptUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x4e, // EPD, SecHdr/Spare, Msg Type
				// ie.PDUSessStatus, 0x50, TLV, 9.11.3.44
				0x50, 0x02, 0x02, 0x00,
				// ie.PDUSessReactivationResult, 0x26, TLV, 9.11.3.42
				0x26, 0x02, 0x00, 0x00,
				// PDUSessReactivationResultErrCause : &ie.PDUSessReactivationResultErrCause 0x72 TLV-E,   5-515B, 9.11.3.43
				// EAPMsg     : &ie.EAPMsg      0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// T3448Value : &ie.GPRSTimer2  0x6B   TLV,       3B, 9.11.2.4
				0x6B, 0x01, 0x56,
			},
			expected: &SvcAccept{
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				PDUSessReactivationResult: &ie.PDUSessReactivationResult{
					Psi: ie.Psi{PSI: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
				T3448Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_Decihours,
					Value: 0x16,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(SvcAccept)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestSvcAcceptMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x4e, 0x50, 0x02, 0x02, 0x00, 0x26, 0x02, 0x00, 0x00,
			},
			input: &SvcAccept{
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				PDUSessReactivationResult: &ie.PDUSessReactivationResult{
					Psi: ie.Psi{PSI: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
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
