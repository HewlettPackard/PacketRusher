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
func FuzzDecode(f *testing.F) {
	valid, _ := Encode(1, 9, ipv4([4]byte{10, 0, 0, 1}, [4]byte{8, 8, 8, 8}))
	f.Add(valid)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		_, p, err := Decode(b)
		if err == nil {
			_, _, _ = IPAddresses(p)
		}
	})
}
