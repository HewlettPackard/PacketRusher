//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
)

func wirePacket(c Config) []byte {
	inner := make([]byte, 32)
	inner[0] = 0x45
	inner[8] = 64
	inner[9] = 17
	binary.BigEndian.PutUint16(inner[2:4], 32)
	copy(inner[12:16], []byte{192, 0, 2, 1})
	copy(inner[16:20], c.IPv4.AsSlice())
	binary.BigEndian.PutUint16(inner[10:12], testpeer.Checksum(inner[:20]))
	binary.BigEndian.PutUint16(inner[20:22], 9000)
	binary.BigEndian.PutUint16(inner[22:24], 50000)
	binary.BigEndian.PutUint16(inner[24:26], 12)
	copy(inner[28:], "wire")
	p := make([]byte, 14+20+8+16+len(inner))
	p[12] = 8
	p[14] = 0x45
	p[22] = 64
	p[23] = 17
	binary.BigEndian.PutUint16(p[16:18], uint16(len(p)-14))
	copy(p[26:30], c.Remote.AsSlice())
	copy(p[30:34], c.Local.AsSlice())
	binary.BigEndian.PutUint16(p[24:26], testpeer.Checksum(p[14:34]))
	binary.BigEndian.PutUint16(p[34:36], 2152)
	binary.BigEndian.PutUint16(p[36:38], 2152)
	binary.BigEndian.PutUint16(p[38:40], uint16(len(p)-34))
	p[42] = 0x34
	p[43] = 255
	binary.BigEndian.PutUint16(p[44:46], uint16(len(inner)+8))
	binary.BigEndian.PutUint32(p[46:50], c.DownlinkTEID)
	copy(p[50:58], []byte{0, 0, 0, 0x85, 1, 0, c.QFI, 0})
	copy(p[58:], inner)
	return p
}
func TestActualKernelPacketBoundsChecksumsAndTupleIsolation(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires privileged disposable namespace")
	}
	r := NewRegistry()
	require.NoError(t, r.load())
	defer r.collection.Close()
	_, _, c := isolatedRegistry()
	require.NoError(t, r.state.put("locals", ipv4(c.Local), uint32(1), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("downlinks", c.downKey(), ipv4(c.IPv4), ebpf.UpdateNoExist))
	require.NoError(t, r.state.put("sessions", ipv4(c.IPv4), c.binding(), ebpf.UpdateNoExist))
	tests := []struct {
		name   string
		change func([]byte) []byte
		result uint32
	}{
		{"valid", func(p []byte) []byte { return p }, 7},
		{"unrelatedUDP", func(p []byte) []byte { binary.BigEndian.PutUint16(p[36:38], 9999); return p }, ^uint32(0)},
		{"unownedN3", func(p []byte) []byte { p[33]++; return p }, ^uint32(0)},
		{"wrongTEID", func(p []byte) []byte { p[49]++; return p }, 2},
		{"wrongPeer", func(p []byte) []byte {
			p[29]++
			p[24] = 0
			p[25] = 0
			binary.BigEndian.PutUint16(p[24:26], testpeer.Checksum(p[14:34]))
			return p
		}, 2},
		{"badOuterChecksum", func(p []byte) []byte { p[25] ^= 1; return p }, 2},
		{"badUDPChecksum", func(p []byte) []byte { p[40] = 0x12; p[41] = 0x34; return p }, 2},
		{"badInnerChecksum", func(p []byte) []byte { p[69] ^= 1; return p }, 2},
		{"wrongQFI", func(p []byte) []byte { p[56]++; return p }, 2},
		{"wrongType", func(p []byte) []byte { p[43] = 1; return p }, 2},
		{"truncatedGTP", func(p []byte) []byte { return p[:49] }, 2},
		{"truncatedUDP", func(p []byte) []byte { return p[:38] }, ^uint32(0)},
		{"wrongInnerDestination", func(p []byte) []byte {
			p[77]++
			p[68] = 0
			p[69] = 0
			binary.BigEndian.PutUint16(p[68:70], testpeer.Checksum(p[58:78]))
			return p
		}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.change(wirePacket(c))
			context := [48]uint32{}
			context[9] = 1
			opts := &ebpf.RunOptions{Data: p, DataOut: make([]byte, len(p)+256), Context: context}
			result, err := r.collection.Programs["decap"].Run(opts)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
			if result == 7 {
				require.Equal(t, wirePacket(c)[58:], opts.DataOut[14:])
			}
		})
	}
	echo := wirePacket(c)[:54]
	echo[42] = 0x32
	echo[43] = 1
	binary.BigEndian.PutUint16(echo[44:46], 4)
	for i := 46; i < 54; i++ {
		echo[i] = 0
	}
	echo[50] = 0x12
	echo[51] = 0x34
	binary.BigEndian.PutUint16(echo[16:18], 40)
	echo[24] = 0
	echo[25] = 0
	binary.BigEndian.PutUint16(echo[24:26], testpeer.Checksum(echo[14:34]))
	binary.BigEndian.PutUint16(echo[38:40], 20)
	ectx := [48]uint32{}
	ectx[9] = 1
	eresult, eerr := r.collection.Programs["decap"].Run(&ebpf.RunOptions{Data: echo, Context: ectx})
	require.NoError(t, eerr)
	require.Equal(t, ^uint32(0), eresult, "validated Echo must reach management socket")
	// A complete nonzero checksum is checked before the tuple/inner parser.
	p := wirePacket(c)
	pseudo := append([]byte(nil), p[26:34]...)
	pseudo = append(pseudo, 0, 17, p[38], p[39])
	pseudo = append(pseudo, p[34:]...)
	binary.BigEndian.PutUint16(p[40:42], testpeer.Checksum(pseudo))
	ctx := [48]uint32{}
	ctx[9] = 1
	result, err := r.collection.Programs["decap"].Run(&ebpf.RunOptions{Data: p, Context: ctx})
	require.NoError(t, err)
	require.Equal(t, uint32(7), result)
}
