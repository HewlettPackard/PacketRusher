//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/cilium/ebpf"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"net/netip"
	"os"
	"testing"
)

func ipv6WirePacket(c Config, inner []byte) []byte {
	p := append(wirePacket(c)[:58], inner...)
	binary.BigEndian.PutUint16(p[16:18], uint16(len(p)-14))
	binary.BigEndian.PutUint16(p[38:40], uint16(len(p)-34))
	binary.BigEndian.PutUint16(p[44:46], uint16(len(inner)+8))
	p[24], p[25] = 0, 0
	binary.BigEndian.PutUint16(p[24:26], testpeer.Checksum(p[14:34]))
	return p
}
func TestActualKernelIPv6PrefixIIDAndHiddenNDAdmission(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	r := NewRegistry()
	require.NoError(t, r.load())
	defer r.collection.Close()
	_, _, c := isolatedRegistry()
	c.AllowIPv6 = true
	c.IPv6InterfaceID[7] = 7
	c.IPv6 = netip.MustParseAddr("2001:db8:1234::7")
	require.NoError(t, r.state.put("locals", ipv4(c.Local), uint32(1), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("downlinks", c.downKey(), c.identity(), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("sessions", c.identity(), c.binding(), ebpf.UpdateNoExist))
	inner := make([]byte, 52)
	inner[0] = 0x60
	inner[6] = 17
	inner[7] = 64
	binary.BigEndian.PutUint16(inner[4:6], 12)
	source := netip.MustParseAddr("2001:db8:ffff::9").As16()
	destination := c.IPv6.As16()
	copy(inner[8:24], source[:])
	copy(inner[24:40], destination[:])
	binary.BigEndian.PutUint16(inner[40:42], 9000)
	binary.BigEndian.PutUint16(inner[42:44], 50000)
	binary.BigEndian.PutUint16(inner[44:46], 12)
	copy(inner[48:], "wire")
	tests := []struct {
		name   string
		packet []byte
		result uint32
	}{
		{"ownedIPv6", append([]byte(nil), inner...), 7},
		{"wrongIID", append([]byte(nil), inner...), 2},
		{"wrongPrefix", append([]byte(nil), inner...), 2},
	}
	tests[1].packet[39]++
	tests[2].packet[28]++
	ra, _ := hex.DecodeString("6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000")
	hidden := append(append(append([]byte(nil), ra[:40]...), 58, 0, 0, 0, 0, 0, 0, 0), ra[40:]...)
	hidden[6] = 0
	binary.BigEndian.PutUint16(hidden[4:6], uint16(len(hidden)-40))
	malformed := append([]byte(nil), hidden...)
	malformed[41] = 255
	fragmented := append([]byte(nil), hidden...)
	fragmented[6] = 44
	tests = append(tests, struct {
		name   string
		packet []byte
		result uint32
	}{"directRAtoControl", ra, ^uint32(0)}, struct {
		name   string
		packet []byte
		result uint32
	}{"hiddenRAtoControl", hidden, ^uint32(0)}, struct {
		name   string
		packet []byte
		result uint32
	}{"malformedExtensionNeverHost", malformed, ^uint32(0)}, struct {
		name   string
		packet []byte
		result uint32
	}{"fragmentedNDNeverHost", fragmented, ^uint32(0)})
	ctx := [48]uint32{}
	ctx[9] = 1
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := ipv6WirePacket(c, tc.packet)
			opts := &ebpf.RunOptions{Data: p, DataOut: make([]byte, len(p)+256), Context: ctx}
			result, err := r.collection.Programs["decap"].Run(opts)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
			if result == 7 {
				require.Equal(t, tc.packet, opts.DataOut[14:])
				require.Equal(t, []byte{0x86, 0xdd}, opts.DataOut[12:14])
			}
		})
	}
	c.IPv6 = netip.Addr{}
	require.NoError(t, r.state.put("sessions", c.identity(), c.binding(), ebpf.UpdateExist))
	result, err := r.collection.Programs["decap"].Run(&ebpf.RunOptions{Data: ipv6WirePacket(c, inner), Context: ctx})
	require.NoError(t, err)
	require.Equal(t, uint32(2), result, "expired prefix no longer authorizes a kernel downlink")
}

// A truncated variable IPv4 header is not an allocated-source packet.
func TestActualKernelUplinkRejectsTruncatedIPv4Options(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	r := NewRegistry()
	require.NoError(t, r.load())
	defer r.collection.Close()
	_, _, c := isolatedRegistry()
	c.EndpointIfIndex = 1
	require.NoError(t, r.state.put("sessions", c.identity(), c.binding(), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("locals", ipv4(c.Local), uint32(1), ebpf.UpdateNoExist))
	c.stageTX = 2
	require.NoError(t, r.state.put("sessions", c.identity(), c.binding(), ebpf.UpdateExist))
	packet := make([]byte, 20)
	packet[0] = 0x4f
	packet[8] = 64
	packet[9] = 17
	binary.BigEndian.PutUint16(packet[2:4], 20)
	copy(packet[12:16], c.IPv4.AsSlice())
	ctx := [48]uint32{}
	ctx[10] = uint32(c.EndpointIfIndex)
	result, err := r.collection.Programs["encap"].Run(&ebpf.RunOptions{Data: packet, Context: ctx})
	require.NoError(t, err)
	require.Equal(t, uint32(2), result)
}

func TestActualKernelChecksumRelayRequiresCurrentOwner(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	r := NewRegistry()
	require.NoError(t, r.load())
	defer r.collection.Close()
	_, _, c := isolatedRegistry()
	c.stageTX = 2
	require.NoError(t, r.state.put("stages", uint32(1), c.identity(), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("locals", ipv4(c.Local), uint32(1), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("sessions", c.identity(), c.binding(), ebpf.UpdateNoExist))
	packet := wirePacket(c)
	copy(packet[26:30], c.Local.AsSlice())
	copy(packet[30:34], c.Remote.AsSlice())
	binary.BigEndian.PutUint32(packet[46:50], c.UplinkTEID)
	packet[55] = 0x10
	copy(packet[70:74], c.IPv4.AsSlice())
	copy(packet[74:78], []byte{192, 0, 2, 1})
	for _, tc := range []struct {
		name   string
		mutate func([]byte)
		result uint32
	}{
		{"allocatedSource", func([]byte) {}, 7},
		{"wrongInnerSource", func(p []byte) { p[73]++ }, 2},
		{"wrongInnerFamily", func(p []byte) { p[58] = 0x60 }, 2},
		{"truncatedIPv4Options", func(p []byte) { p[58] = 0x4f }, 2},
		{"wrongQFI", func(p []byte) { p[56]++ }, 2},
		{"wrongContainerDirection", func(p []byte) { p[55] = 0 }, 2},
		{"retiredTEID", func(p []byte) { p[49]++ }, 2},
		{"wrongOuterPeer", func(p []byte) { p[33]++ }, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := append([]byte(nil), packet...)
			tc.mutate(p)
			ctx := [48]uint32{}
			ctx[9] = 1
			result, err := r.collection.Programs["relay"].Run(&ebpf.RunOptions{Data: p, Context: ctx})
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
		})
	}
	ctx := [48]uint32{}
	ctx[9] = 99
	result, err := r.collection.Programs["relay"].Run(&ebpf.RunOptions{Data: packet, Context: ctx})
	require.NoError(t, err)
	require.Equal(t, uint32(2), result, "unowned staging ingress cannot borrow an allocation")
}
