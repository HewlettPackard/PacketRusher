package snow3g

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnow3g(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		k      [4]uint32
		iv     [4]uint32
		z      []uint32
		length int
	}{
		{
			name:   "TestCase1",
			k:      [4]uint32{0x2bd6459f, 0x82c5b300, 0x952c4910, 0x4881ff48},
			iv:     [4]uint32{0xea024714, 0xad5c4d84, 0xdf1f9b25, 0x1c0bf45f},
			z:      []uint32{0xabee9704, 0x7ac31373},
			length: 2,
		},
		{
			name:   "TestCase2",
			k:      [4]uint32{0x8ce33e2c, 0xc3c0b5fc, 0x1f3de8a6, 0xdc66b1f3},
			iv:     [4]uint32{0xd3c5d592, 0x327fb11c, 0xde551988, 0xceb2f9b7},
			z:      []uint32{0xeff8a342, 0xf751480f},
			length: 2,
		},
		{
			name:   "TestCase3",
			k:      [4]uint32{0x4035c668, 0x0af8c6d1, 0xa8ff8667, 0xb1714013},
			iv:     [4]uint32{0x62a54098, 0x1ba6f9b7, 0x4592b0e7, 0x8690f71b},
			z:      []uint32{0xa8c874a9, 0x7ae7c4f8},
			length: 2,
		},
		{
			name:   "TestCase4",
			k:      [4]uint32{0x0ded7263, 0x109cf92e, 0x3352255a, 0x140e0f76},
			iv:     [4]uint32{0x6b68079a, 0x41a7c4c9, 0x1befd79f, 0x7fdcc233},
			z:      []uint32{0xd712c05c, 0xa937c2a6, 0xeb7eaae3},
			length: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ks := GetKeyStream(tc.k, tc.iv, tc.length)
			require.Equal(t, tc.z, ks)
		})
	}
}
