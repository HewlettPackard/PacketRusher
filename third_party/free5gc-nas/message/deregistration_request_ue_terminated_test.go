package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestDeregReqUETermUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1 - Mandatory only",
			input: []byte{
				0x7e, 0x00, 0x47, 0x09,
			},
			expected: &DeregReqUETerm{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
			},
		},
		{
			name: "Case 2 - Mandatory + 5GMM cause + T3346 Value",
			input: []byte{
				0x7e, 0x00, 0x47, 0x09,
				// 5GMM cause
				0x58, 0x05,
				// T3346 Value
				0x5F, 0x01, 0x56,
				// Rejected NSSAI
				// CAG Information List
			},
			expected: &DeregReqUETerm{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_PEINotAccepted,
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_Decihours,
					Value: 0x16,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(DeregReqUETerm)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestDeregReqUETermMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1 - mandatory only",
			expected: []byte{
				0x7e, 0x00, 0x47, 0x09,
			},
			input: &DeregReqUETerm{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
			},
		},
		{
			name: "Case 2 - Mandatory + 5GMM cause + T3346 Value",
			expected: []byte{
				0x7e, 0x00, 0x47, 0x09,
				// 5GMM cause
				0x58, 0x05,
				// T3346 Value
				0x5F, 0x01, 0x56,
				// Rejected NSSAI
				// CAG Information List
			},
			input: &DeregReqUETerm{
				DeregType: &ie.DeregType{
					Switchoff:     true,
					ReregRequired: false,
					AccessType:    1,
				},
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_PEINotAccepted,
				},
				T3346Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_Decihours,
					Value: 0x16,
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
