package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacketFilterPortBoundsPreserveWireEncoding(t *testing.T) {
	t.Parallel()
	for _, side := range []string{"local", "remote"} {
		t.Run(side, func(t *testing.T) {
			singleType, rangeType := byte(0x40), byte(0x41)
			if side == "remote" {
				singleType, rangeType = 0x50, 0x51
			}
			for _, tc := range []struct {
				name, value, decoded string
				wire                 []byte
			}{
				{"zero", "0", "0", []byte{singleType, 0x00, 0x00}},
				{"maximum", "65535", "65535", []byte{singleType, 0xff, 0xff}},
				{"decimal-leading-zeros", "00080", "80", []byte{singleType, 0x00, 0x50}},
				{"decimal-plus-sign", "+80", "80", []byte{singleType, 0x00, 0x50}},
				{"full-range", "0-65535", "0-65535", []byte{rangeType, 0x00, 0x00, 0xff, 0xff}},
				{"maximum-range", "65535-65535", "65535-65535", []byte{rangeType, 0xff, 0xff, 0xff, 0xff}},
				{"range-plus-signs", "+80-+443", "80-443", []byte{rangeType, 0x00, 0x50, 0x01, 0xbb}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					contents := packetFilterPort(side, tc.value)
					wire, err := contents.MarshalBinary()
					require.NoError(t, err)
					require.Equal(t, tc.wire, wire)
					var decoded PacketFilterContents
					require.NoError(t, decoded.UnmarshalBinary(wire))
					if side == "local" {
						require.Equal(t, tc.decoded, decoded.LocalPortRange)
					} else {
						require.Equal(t, tc.decoded, decoded.RemotePortRange)
					}
				})
			}
		})
	}
}

func TestPacketFilterPortsRejectInvalidEndpoints(t *testing.T) {
	t.Parallel()
	for _, side := range []string{"local", "remote"} {
		t.Run(side, func(t *testing.T) {
			for _, value := range []string{
				"65536", "131071", "+65536", "18446744073709551616",
				"65536-65535", "0-65536", "65536-65536",
				"18446744073709551616-65535", "0-18446744073709551616",
				"-1", "-1-80", "80--1", "80-", "not-a-port", "0-invalid",
			} {
				t.Run(value, func(t *testing.T) {
					contents := packetFilterPort(side, value)
					wire, err := contents.MarshalBinary()
					require.Error(t, err)
					require.Nil(t, wire, "invalid ports must not produce a truncated wire value")
				})
			}
		})
	}
}

func packetFilterPort(side, value string) PacketFilterContents {
	contents := PacketFilterContents{RemoteAddr: "any", LocalAddr: "assigned"}
	if side == "local" {
		contents.LocalPortRange = value
	} else {
		contents.RemotePortRange = value
	}
	return contents
}
