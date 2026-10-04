// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"testing"
)

func ipv4(src, dst [4]byte) []byte {
	p := make([]byte, 24)
	p[0] = 0x45
	binary.BigEndian.PutUint16(p[2:4], uint16(len(p)))
	p[8] = 64
	p[9] = 17
	copy(p[12:16], src[:])
	copy(p[16:20], dst[:])
	copy(p[20:], []byte{1, 2, 3, 4})
	return p
}
func TestGTPSessionContainerAndOptionalChains(t *testing.T) {
	p := ipv4([4]byte{10, 0, 0, 1}, [4]byte{8, 8, 8, 8})
	for _, qfi := range []uint8{0, 1, 63} {
		wire, err := Encode(0x01020304, qfi, p)
		require.NoError(t, err)
		if qfi != 0 {
			require.Equal(t, []byte{0x34, 255, 0, 32, 1, 2, 3, 4, 0, 0, 0, 0x85, 1, 0x10, qfi, 0}, wire[:16])
		}
		id, inner, err := Decode(wire)
		require.NoError(t, err)
		require.Equal(t, uint32(0x01020304), id)
		require.Equal(t, p, inner)
	}
	// A valid chained UDP Port extension and PDU Session Container precede IP.
	chained := append([]byte{0x34, 255, 0, 36, 1, 2, 3, 4, 0, 0, 0, 0x40, 1, 0x08, 0x68, 0x85, 1, 0, 9, 0}, p...)
	_, inner, err := Decode(chained)
	require.NoError(t, err)
	require.Equal(t, p, inner)
	for _, mutate := range []func([]byte) []byte{
		func(b []byte) []byte { return b[:7] },
		func(b []byte) []byte { b[2] = 255; return b },
		func(b []byte) []byte { b[12] = 0; return b },
		func(b []byte) []byte { b[12] = 255; return b },
		func(b []byte) []byte { b[0] = 0x30; return b },
		func(b []byte) []byte { b[0] = 0x74; return b },
	} {
		_, _, err = Decode(mutate(append([]byte(nil), chained...)))
		require.Error(t, err)
	}
	_, err = Encode(0, 9, p)
	require.Error(t, err)
	_, err = Encode(1, 64, p)
	require.Error(t, err)
	_, err = Encode(1, 9, make([]byte, 65507))
	require.Error(t, err)
}
func TestEchoAndIPLengthValidation(t *testing.T) {
	require.Equal(t, []byte{0x32, 2, 0, 6, 0, 0, 0, 0, 0x12, 0x34, 0, 0, 14, 0}, echoResponse([]byte{0x32, 1, 0, 4, 0, 0, 0, 0, 0x12, 0x34, 0, 0}))
	require.Nil(t, echoResponse([]byte{0x30, 1, 0, 0, 0, 0, 0, 0}))
	p := ipv4([4]byte{10, 0, 0, 1}, [4]byte{8, 8, 8, 8})
	_, _, err := IPAddresses(p)
	require.NoError(t, err)
	p[0] = 0x44
	_, _, err = IPAddresses(p)
	require.Error(t, err)
	p[0] = 0x45
	p[3]--
	_, _, err = IPAddresses(p)
	require.Error(t, err)
}

func TestDownlinkOptionalHeadersAndAllocatedContainer(t *testing.T) {
	payload := ipv4([4]byte{192, 0, 2, 1}, [4]byte{10, 0, 0, 1})
	wire := func(flags byte, optional []byte) []byte {
		p := append([]byte{flags, 255, 0, 0, 0, 0, 0, 1}, optional...)
		p = append(p, payload...)
		binary.BigEndian.PutUint16(p[2:4], uint16(len(p)-8))
		return p
	}
	valid := [][]byte{
		wire(0x30, nil),
		wire(0x31, []byte{0, 0, 7, 0}),
		wire(0x33, []byte{0x12, 0x34, 7, 0}),
		wire(0x37, []byte{0x12, 0x34, 7, 0x40, 1, 8, 0x68, 0x85, 1, 0, 9, 0}),
		wire(0x34, []byte{0, 0, 0, 0x40, 1, 8, 0x68, 0}),
		wire(0x34, []byte{0, 0, 0, 0x85, 1, 1, 0x49, 0}),                                                       // spare/RQI with allocated QFI
		wire(0x34, []byte{0, 0, 0, 0x85, 2, 0, 0xc9, 0xe0, 0, 0, 0, 0}),                                        // PPP/PPI/RQI
		wire(0x34, []byte{0, 0, 0, 0x85, 3, 8, 9, 1, 2, 3, 4, 5, 6, 7, 8, 0}),                                  // QMP timestamp
		wire(0x34, []byte{0, 0, 0, 0x85, 5, 0x0f, 0xc9, 0xe0, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 1, 2, 3, 4, 0}), // all optional fields
	}
	for i, p := range valid {
		id, inner, err := DecodeDownlink(p, 9)
		require.NoError(t, err, "valid format %d", i)
		require.Equal(t, uint32(1), id)
		require.Equal(t, payload, inner)
	}
	// 32 bounded extensions are supported; 33 cannot consume unbounded work.
	chain := []byte{0, 0, 0, 0x40}
	for i := 0; i < 32; i++ {
		chain = append(chain, 1, 0, 0, 0x40)
	}
	chain[len(chain)-1] = 0
	_, _, err := DecodeDownlink(wire(0x34, chain), 9)
	require.NoError(t, err)
	chain[len(chain)-1] = 0x40
	chain = append(chain, 1, 0, 0, 0)
	_, _, err = DecodeDownlink(wire(0x34, chain), 9)
	require.Error(t, err)
	for _, optional := range [][]byte{
		{0, 0, 0, 0x85, 1, 0x10, 9, 0},                // uplink direction
		{0, 0, 0, 0x85, 1, 0, 8, 0},                   // wrong flow
		{0, 0, 0, 0x85, 1, 8, 9, 0},                   // missing timestamp
		{0, 0, 0, 0x85, 1, 0, 0x89, 0},                // missing PPI
		{0, 0, 0, 0x85, 1, 4, 9, 0},                   // missing QFI sequence
		{0, 0, 0, 0x85, 1, 2, 9, 0},                   // missing MBS sequence
		{0, 0, 0, 0x85, 2, 0x0f, 0xc9, 0, 0, 0, 0, 0}, // combined fields cannot fit
		{0, 0, 0, 0x85, 1, 0, 9, 0x85, 1, 0, 9, 0},    // duplicate PSC
		{0, 0, 0, 0x40, 1, 0, 0, 0x85, 1, 0x10, 9, 0}, // hidden wrong direction
	} {
		p := wire(0x34, optional)
		_, _, err := Decode(p)
		require.NoError(t, err, "public generic parser remains policy-free")
		_, _, err = DecodeDownlink(p, 9)
		require.Error(t, err)
	}
	for _, optional := range [][]byte{{0, 0, 0, 0x40, 0, 0, 0, 0}, {0, 0, 0, 0x40, 255, 0, 0, 0}} {
		_, _, err := DecodeDownlink(wire(0x34, optional), 9)
		require.Error(t, err)
	}
}
func FuzzDecode(f *testing.F) {
	valid, _ := Encode(1, 9, ipv4([4]byte{10, 0, 0, 1}, [4]byte{8, 8, 8, 8}))
	f.Add(valid)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		_, p, err := Decode(b)
		if err == nil {
			_, _, _ = IPAddresses(p)
		}
		_, _, _ = DecodeDownlink(b, 9)
	})
}
