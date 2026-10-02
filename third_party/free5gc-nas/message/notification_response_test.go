package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestNotifRspUnmarshalBinary(t *testing.T) {
	t.Parallel()

	pduSessStatus := &ie.PDUSessStatus{
		Psi: ie.Psi{
			PSI: [16]bool{
				true, false, true, false, true, false, true, false,
				true, false, true, false, true, false, true, false,
			},
		},
	}

	testCases := []struct {
		name      string
		input     []byte
		expected  *NotifRsp
		expectErr bool
	}{
		{
			name: "Valid Notification Response with PDU Session Status",
			input: []byte{
				byte(Epd5GSMobilityMgmtMsg), // EPD
				byte(SecHdrTypePlainNas),    // Security Header
				byte(MsgTypeNotifRsp),       // Message Type
				NotifRspIEIPDUSessStatus,    // IEI
				0x02,                        // Length
				0x55, 0x55,                  // PDU Session Status content
			},
			expected: &NotifRsp{
				PDUSessStatus: pduSessStatus,
			},
			expectErr: false,
		},
		{
			name: "Valid Notification Response without optional IEs",
			input: []byte{
				byte(Epd5GSMobilityMgmtMsg),
				byte(SecHdrTypePlainNas),
				byte(MsgTypeNotifRsp),
			},
			expected:  &NotifRsp{},
			expectErr: false,
		},
		{
			name: "Invalid input - too short",
			input: []byte{
				byte(Epd5GSMobilityMgmtMsg),
				byte(SecHdrTypePlainNas),
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "Invalid input - unknown IEI",
			input: []byte{
				byte(Epd5GSMobilityMgmtMsg),
				byte(SecHdrTypePlainNas),
				byte(MsgTypeNotifRsp),
				0x99, // Unknown IEI
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var notifRsp NotifRsp
			err := notifRsp.UnmarshalBinary(tc.input)

			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, &notifRsp)
		})
	}
}

func TestNotifRspMarshalBinary(t *testing.T) {
	t.Parallel()

	pduSessStatus := &ie.PDUSessStatus{
		Psi: ie.Psi{
			PSI: [16]bool{
				true, false, true, false, true, false, true, false,
				true, false, true, false, true, false, true, false,
			},
		},
	}

	testCases := []struct {
		name      string
		input     *NotifRsp
		expected  []byte
		expectErr bool
	}{
		{
			name: "Marshal Notification Response with PDU Session Status",
			input: &NotifRsp{
				PDUSessStatus: pduSessStatus,
			},
			expected: []byte{
				byte(Epd5GSMobilityMgmtMsg), // EPD
				byte(SecHdrTypePlainNas),    // Security Header
				byte(MsgTypeNotifRsp),       // Message Type
				NotifRspIEIPDUSessStatus,    // IEI
				0x02,                        // Length
				0x55, 0x55,                  // PDU Session Status content
			},
			expectErr: false,
		},
		{
			name:  "Marshal Notification Response without optional IEs",
			input: &NotifRsp{},
			expected: []byte{
				byte(Epd5GSMobilityMgmtMsg),
				byte(SecHdrTypePlainNas),
				byte(MsgTypeNotifRsp),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := tc.input.MarshalBinary()

			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, output)
		})
	}
}
