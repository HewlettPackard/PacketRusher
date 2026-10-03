package ie

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendedProtCfgOptsUnmarshalFromMs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *ExtendedProtCfgOpts
	}{
		{
			name: "Positive case 1",
			input: []byte{
				0x80,
				0x80, 0x21, 0x10, // Internet Protocol Control Protocol, 16B
				0x01, 0x00, 0x00, 0x10, // 16B
				0x81, 0x06, 0x00, 0x00, 0x00, 0x00, // Primary DNS server IP addr, 6B, 0.0.0.0
				0x83, 0x06, 0x00, 0x00, 0x00, 0x00, // 2nd DNS server IP addr, 6B, 0.0.0.0
				0x00, 0x0d, 0x00, // DNS IPv4 Req, 0B
				0x00, 0x0a, 0x00, // IP Addr Allocation via NAS Signaling, 0B
				0x00, 0x05, 0x00, // MS Support of Nw Req Bearer Ctrl Indicator, 0B
				0x00, 0x10, 0x00, // IPv4 Link MTU Req, 0B
				0x00, 0x11, 0x00, // MS Support of Local Addr in TFT Indicator, 0B
				0x00, 0x17, 0x01, 0x01, // 3GPP PS data off UE status, 1B
				0x00, 0x23, 0x00, // QoS Rules w/ the len of 2B support indicator, 0B
				0x00, 0x24, 0x00, // QoS flow desc w/ the len of 2B support indicator, 0B
			},
			expected: &ExtendedProtCfgOpts{
				FromMs: &ExtCfgOptFromMs{
					IPv4LinkMTUReq:        true,
					DNSV4Req:              true,
					UEStatus3GPPPSDataOff: UEStatus3GPPPSDataOff_Deactivate,
				},
			},
		},
		{
			name: "Positive case 2",
			input: []byte{
				0x80,
				0x00, 0x03, 0x00, // ipv6 dns
				0x00, 0x0d, 0x00, // ipv4 dns
				0x00, 0x10, 0x00, // mtu
			},
			expected: &ExtendedProtCfgOpts{
				FromMs: &ExtCfgOptFromMs{
					IPv4LinkMTUReq:        true,
					DNSV4Req:              true,
					DNSV6Req:              true,
					UEStatus3GPPPSDataOff: UEStatus3GPPPSDataOff_NotPresent,
				},
			},
		},
		{
			name: "Negative case - bad total length 1",
			input: []byte{
				0x80,
				0x80, 0x21, 0x10, // Internet Protocol Control Protocol, 16B
				0x01, 0x00, 0x00, 0x10, // 16B
				0x81, 0x06, 0x00, 0x00, 0x00, 0x00, // Primary DNS server IP addr, 6B, 0.0.0.0
				0x83, 0x06, 0x00, 0x00, 0x00, 0x00, // 2nd DNS server IP addr, 6B, 0.0.0.0
				0x00, 0x0d, 0x08, // DNS IPv4 Req, 8B <-- 0x00, 0x0d, 0x00
				0x00, 0x0a, 0x00, // IP Addr Allocation via NAS Signaling, 0B
				0x00, 0x05, 0x00, // MS Support of Nw Req Bearer Ctrl Indicator, 0B
				0x00, 0x10, 0x00, // IPv4 Link MTU Req, 0B
				0x00, 0x11, 0x00, // MS Support of Local Addr in TFT Indicator, 0B
				0x00, 0x17, 0x01, 0x01, // 3GPP PS data off UE status, 1B
				0x00, 0x23, 0x00, // QoS Rules w/ the len of 2B support indicator, 0B
				0x00, 0x24, 0x00, // QoS flow desc w/ the len of 2B support indicator, 0B
			},
		},
		{
			name: "Negative case - bad total length 2",
			input: []byte{
				0x80,
				0x80, 0x21, 0x10, // Internet Protocol Control Protocol, 16B
				0x01, 0x00, 0x00, 0x10, // 16B
				0x81, 0x06, 0x00, 0x00, 0x00, 0x00, // Primary DNS server IP addr, 6B, 0.0.0.0
				0x83, 0x06, 0x00, 0x00, 0x00, 0x00, // 2nd DNS server IP addr, 6B, 0.0.0.0
				0x00, 0x0d, 0x00, // DNS IPv4 Req, 0B
				0x00, 0x0a, 0x00, // IP Addr Allocation via NAS Signaling, 0B
				0x00, 0x05, 0x00, // MS Support of Nw Req Bearer Ctrl Indicator, 0B
				0x00, 0x10, 0x09, // IPv4 Link MTU Req, 9B  <-- 0x00, 0x10, 0x00
				0x00, 0x11, 0x00, // MS Support of Local Addr in TFT Indicator, 0B
				0x00, 0x17, 0x01, 0x01, // 3GPP PS data off UE status, 1B
				0x00, 0x23, 0x00, // QoS Rules w/ the len of 2B support indicator, 0B
				0x00, 0x24, 0x00, // QoS flow desc w/ the len of 2B support indicator, 0B
			},
		},
		{
			name: "Positive case - 3GPP PS data off UE status",
			input: []byte{
				0x80,
				0x00, 0x17, 0x01, // 3GPP PS data off UE status, 1B
				0x02, // UEStatus3GPPPSDataOff_Activate
			},
			expected: &ExtendedProtCfgOpts{
				FromMs: &ExtCfgOptFromMs{
					UEStatus3GPPPSDataOff: UEStatus3GPPPSDataOff_Activate,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(ExtendedProtCfgOpts)
			err := ie.UnmarshalFromMs(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestExtendedProtCfgOptsUnmarshalFromNw(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *ExtendedProtCfgOpts
	}{
		{
			name: "Positive case 1",
			input: []byte{
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 Req, 0B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU Req, 0B
				0x05, 0x78, // 1400
			},
			expected: &ExtendedProtCfgOpts{
				FromNw: &ExtCfgOptFromNw{
					IPv4LinkMTU: 1400,
					DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
				},
			},
		},
		{
			name: "Positive case 2 - 3GPP PS Data Off Support Indicate",
			input: []byte{
				0x80,
				0x00, 0x17, 0x00, // 3GPP PS Data Off Support Indication, 0B
			},
			expected: &ExtendedProtCfgOpts{
				FromNw: &ExtCfgOptFromNw{
					Indication3GPPPSDataOffSupport: true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(ExtendedProtCfgOpts)
			err := ie.UnmarshalFromNw(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestExtendedProtCfgOptsMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *ExtendedProtCfgOpts
		expected []byte
	}{
		{
			name: "Positive case 1",
			expected: []byte{
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 Req, 0B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU Req, 0B
				0x05, 0x78, // 1400
			},
			input: &ExtendedProtCfgOpts{
				FromNw: &ExtCfgOptFromNw{
					IPv4LinkMTU: 1400,
					DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
				},
			},
		},
		{
			name: "Positive case 2",
			expected: []byte{
				0x80,
				0x00, 0x03, 0x00, // ipv6 dns
				0x00, 0x0d, 0x00, // ipv4 dns
				0x00, 0x10, 0x00, // mtu
			},
			input: &ExtendedProtCfgOpts{
				FromMs: &ExtCfgOptFromMs{
					IPv4LinkMTUReq: true,
					DNSV4Req:       true,
					DNSV6Req:       true,
				},
			},
		},
		{
			name: "Positive case 3",
			expected: []byte{
				0x80,
				0x00, 0x17, 0x01, // UE 3GPP PS Data Off Status
				0x02, // UEStatus3GPPPSDataOff_Deactivate
			},
			input: &ExtendedProtCfgOpts{
				FromMs: &ExtCfgOptFromMs{
					UEStatus3GPPPSDataOff: UEStatus3GPPPSDataOff_Activate,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}
