package algo

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"

	"github.com/free5gc/nas/internal/algo/snow3g"
	"github.com/free5gc/nas/internal/algo/zuc"
)

func NEA1(ck [16]byte, countC, bearer, direction uint32, ibs []byte, length uint32) (obs []byte, err error) {
	var k [4]uint32
	for i := uint32(0); i < 4; i++ {
		k[i] = binary.BigEndian.Uint32(ck[4*(3-i) : 4*(3-i+1)])
	}
	iv := [4]uint32{(bearer << 27) | (direction << 26), countC, (bearer << 27) | (direction << 26), countC}

	l := (length + 31) / 32
	r := length % 32
	ks := snow3g.GetKeyStream(k, iv, int(l))
	// Clear keystream bits which exceed length
	if r != 0 {
		ks[l-1] &= ^((1 << (32 - r)) - 1)
	}

	obs = make([]byte, len(ibs))
	var i uint32
	for i = 0; i < length/32; i++ {
		for j := uint32(0); j < 4; j++ {
			obs[4*i+j] = ibs[4*i+j] ^ byte((ks[i]>>(8*(3-j)))&0xff)
		}
	}
	if r != 0 {
		ll := (r + 7) / 8
		for j := uint32(0); j < ll; j++ {
			obs[4*i+j] = ibs[4*i+j] ^ byte((ks[i]>>(8*(3-j)))&0xff)
		}
	}
	return obs, nil
}

// ibs: input bit stream, obs: output bit stream
func NEA2(key [16]byte, count uint32, bearer uint8, direction uint8,
	ibs []byte,
) (obs []byte, err error) {
	// Couter[0..32] | BEARER[0..4] | DIRECTION[0] | 0^26 | 0^64
	couterBlk := make([]byte, 16)
	// First 32 bits are count
	binary.BigEndian.PutUint32(couterBlk, count)
	// Put Bearer and direction together
	couterBlk[4] = (bearer << 3) | (direction << 2)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	obs = make([]byte, len(ibs))

	stream := cipher.NewCTR(block, couterBlk)
	stream.XORKeyStream(obs, ibs)
	return obs, nil
}

// NEA3 ibs: input bit stream, obs: output bit stream
// ref: https://www.gsma.com/security/wp-content/uploads/2019/05/EEA3_EIA3_specification_v1_8.pdf
func NEA3(ck [16]byte, count uint32, bearer uint8, direction uint8,
	ibs []byte, length uint32,
) (obs []byte, err error) {
	iv := make([]byte, 16)
	binary.BigEndian.PutUint32(iv, count)
	iv[4] = (bearer << 3) | (direction << 2)

	for i := 0; i < 8; i++ {
		iv[i+8] = iv[i]
	}

	l := (length + 31) / 32
	stream := zuc.Zuc(ck[:], iv, l)

	obs = make([]byte, len(ibs))

	for i := 0; i < int(l); i++ {
		for j := 0; j < 4 && (i*4+j) < int((length+7)/8); j++ {
			obs[i*4+j] = ibs[i*4+j] ^ byte((stream[i]>>(8*(3-j)))&0xff)
		}
	}

	if length%8 != 0 {
		obs[length/8] &= (uint8(0xff) << (8 - length%8))
	}

	for j := int(length/8 + 1); j < len(obs); j++ {
		obs[j] = 0
	}
	return obs, nil
}
