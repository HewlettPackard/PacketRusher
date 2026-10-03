// SPDX-License-Identifier: Apache-2.0
package ipv6

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

// Independently generated IPv6/ICMPv6 bytes, including pseudo-header checksum.
const advertisementFixture = "6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000"

func TestRouterAdvertisementWireFixture(t *testing.T) {
	packet, err := hex.DecodeString(advertisementFixture)
	require.NoError(t, err)
	iid := [8]byte{0, 0, 0, 0, 0, 0, 0, 7}
	advertisement, err := ParseAdvertisement(packet, iid)
	require.NoError(t, err)
	require.Equal(t, netip.MustParsePrefix("2001:db8:1234::/64"), advertisement.Prefix)
	require.Equal(t, netip.MustParseAddr("2001:db8:1234::7"), advertisement.Address)
	require.Equal(t, uint32(3600), advertisement.ValidLifetime)
	require.Equal(t, uint32(1800), advertisement.PreferredLifetime)
	for _, offset := range []int{4, 6, 7, 8, 24, 40, 41, 42, 46, 56, 57, 58, 59, 63, 67} {
		corrupt := append([]byte(nil), packet...)
		corrupt[offset] ^= 1
		_, err := ParseAdvertisement(corrupt, iid)
		require.Error(t, err, "tampered octet %d", offset)
	}
	for length := 0; length < len(packet); length++ {
		_, err := ParseAdvertisement(packet[:length], iid)
		require.Error(t, err)
	}
}

func TestRouterAdvertisementSemanticsWithValidChecksum(t *testing.T) {
	packet, _ := hex.DecodeString(advertisementFixture)
	for _, modify := range []func([]byte){
		func(b []byte) { b[58] = 48 }, // /48 is unsuitable for a 64-bit NAS IID.
		func(b []byte) { b[59] &= ^byte(0x40) },

		func(b []byte) { copy(b[64:68], []byte{0, 0, 0xff, 0xff}) },
		func(b []byte) { b[57] = 0 },
		func(b []byte) { b[57] = 5 },
		func(b []byte) { b[72] = 0xfe; b[73] = 0x80 },
		func(b []byte) { b[72] = 0xff },
		func(b []byte) { b[8] = 0x20; b[9] = 1 },
	} {
		malformed := append([]byte(nil), packet...)
		modify(malformed)
		malformed[42], malformed[43] = 0, 0
		checksumValue := checksum(malformed)
		malformed[42], malformed[43] = byte(checksumValue>>8), byte(checksumValue)
		_, err := ParseAdvertisement(malformed, [8]byte{7})
		require.Error(t, err)
	}
}

func TestRouterSolicitationUsesAllocatedLinkLocalAndAllRouters(t *testing.T) {
	iid := [8]byte{0, 0, 0, 0, 0, 0, 0, 7}
	packet := RouterSolicitation(iid)
	require.Equal(t, 48, len(packet))
	require.Equal(t, byte(255), packet[7])
	require.Equal(t, byte(133), packet[40])
	require.Equal(t, netip.MustParseAddr("fe80::7").AsSlice(), packet[8:24])
	require.Equal(t, netip.MustParseAddr("ff02::2").AsSlice(), packet[24:40])
	require.Equal(t, uint16(0), checksum(packet))
}
