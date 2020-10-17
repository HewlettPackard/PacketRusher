package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"my5G-RANTester/lib/nas/security/snow3g"

	"github.com/aead/cmac"
)

func NASEncrypt(AlgoID uint8, KnasEnc [16]byte, Count uint32, Bearer uint8, Direction uint8, payload []byte) error {
	if Bearer > 0x1f {
		return fmt.Errorf("Bearer is beyond 5 bits")
	}
	if Direction > 1 {
		return fmt.Errorf("Direction is beyond 1 bits")
	}
	if payload == nil {
		return fmt.Errorf("Nas Payload is nil")
	}

	switch AlgoID {
	case AlgCiphering128NEA0:
		// logger.SecurityLog.Debugf("ALG_CIPHERING is ALG_CIPHERING_128_NEA0")
		return nil
	case AlgCiphering128NEA1:
		output, err := NEA1(KnasEnc, Count, uint32(Bearer), uint32(Direction), payload, uint32(len(payload))*8)
		if err != nil {
			return err
		}
		// Override payload with NEA1 output
		copy(payload, output)
		return nil
	case AlgCiphering128NEA2:
		output, err := NEA2(KnasEnc, Count, Bearer, Direction, payload)
		if err != nil {
			return err
		}
		// Override payload with NEA2 output
		copy(payload, output)
		return nil
	case AlgCiphering128NEA3:
		return fmt.Errorf("NEA3 not implement yet.")
	default:
		return fmt.Errorf("Unknown Algorithm Identity[%d]", AlgoID)
	}
}

func NASMacCalculate(AlgoID uint8, KnasInt [16]uint8, Count uint32, Bearer uint8, Direction uint8, msg []byte) ([]byte, error) {
	if Bearer > 0x1f {
		return nil, fmt.Errorf("Bearer is beyond 5 bits")
	}
	if Direction > 1 {
		return nil, fmt.Errorf("Direction is beyond 1 bits")
	}
	if msg == nil {
		return nil, fmt.Errorf("Nas Payload is nil")
	}

	switch AlgoID {
	case AlgIntegrity128NIA0:
		// logger.SecurityLog.Warningln("Integrity NIA0 is emergency.")
		return nil, nil
	case AlgCiphering128NEA1:
		return NIA1(KnasInt, Count, Bearer, uint32(Direction), msg, uint64(len(msg))*8)
	case AlgIntegrity128NIA2:
		return NIA2(KnasInt, Count, Bearer, Direction, msg)
	case AlgIntegrity128NIA3:
		// logger.SecurityLog.Errorf("NIA3 not implement yet.")
		return nil, nil
	default:
		return nil, fmt.Errorf("Unknown Algorithm Identity[%d]", AlgoID)
	}

}

func NEA1(ck [16]byte, countC, bearer, direction uint32, ibs []byte, length uint32) (obs []byte, err error) {
	var k [4]uint32
	for i := uint32(0); i < 4; i++ {
		k[i] = binary.BigEndian.Uint32(ck[4*(3-i) : 4*(3-i+1)])
	}
	iv := [4]uint32{(bearer << 27) | (direction << 26), countC, (bearer << 27) | (direction << 26), countC}
	snow3g.InitSnow3g(k, iv)

	l := (length + 31) / 32
	r := length % 32
	ks := make([]uint32, l)
	snow3g.GenerateKeystream(int(l), ks)
	// Clear keystream bits which exceed length
	ks[l-1] &= ^((1 << (32 - r)) - 1)

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
func NEA2(key [16]byte, count uint32, bearer uint8, direction uint8, ibs []byte) (obs []byte, err error) {
	// Couter[0..32] | BEARER[0..4] | DIRECTION[0] | 0^26 | 0^64
	couterBlk := make([]byte, 16)
	//First 32 bits are count
	binary.BigEndian.PutUint32(couterBlk, count)
	//Put Bearer and direction together
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

func NEA3() {

}

// mulx() is for NIA1()
func mulx(V, c uint64) uint64 {
	if V&0x8000000000000000 != 0 {
		return (V << 1) ^ c
	} else {
		return V << 1
	}
}

// mulxPow() is for NIA1()
func mulxPow(V, i, c uint64) uint64 {
	if i == 0 {
		return V
	} else {
		return mulx(mulxPow(V, i-1, c), c)
	}
}

// mul() is for NIA1()
func mul(V, P, c uint64) uint64 {
	rst := uint64(0)
	for i := uint64(0); i < 64; i++ {
		if (P>>i)&1 == 1 {
			rst ^= mulxPow(V, i, c)
		}
	}
	return rst
}

func NIA1(ik [16]byte, countI uint32, bearer byte, direction uint32, msg []byte, length uint64) (mac []byte, err error) {
	fresh := uint32(bearer) << 27
	var k [4]uint32
	for i := uint32(0); i < 4; i++ {
		k[i] = binary.BigEndian.Uint32(ik[4*(3-i) : 4*(3-i+1)])
	}
	iv := [4]uint32{fresh ^ (direction << 15), countI ^ (direction << 31), fresh, countI}
	D := ((length + 63) / 64) + 1
	var z = make([]uint32, 5)
	snow3g.InitSnow3g(k, iv)
	snow3g.GenerateKeystream(5, z)

	P := (uint64(z[0]) << 32) | uint64(z[1])
	Q := (uint64(z[2]) << 32) | uint64(z[3])

	var Eval uint64 = 0
	for i := uint64(0); i < D-2; i++ {
		M := binary.BigEndian.Uint64(msg[8*i:])
		Eval = mul(Eval^M, P, 0x000000000000001b)
	}

	tmp := make([]byte, 8)
	copy(tmp, msg[8*(D-2):])
	M := binary.BigEndian.Uint64(tmp)
	Eval = mul(Eval^M, P, 0x000000000000001b)

	Eval = Eval ^ uint64(length)
	Eval = mul(Eval, Q, 0x000000000000001b)
	MacI := uint32(uint32(Eval>>32) ^ z[4])
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, MacI)
	return b, nil
}

func NIA2(key [16]byte, count uint32, bearer uint8, direction uint8, msg []byte) (mac []byte, err error) {
	// Couter[0..32] | BEARER[0..4] | DIRECTION[0] | 0^26
	m := make([]byte, len(msg)+8)
	//First 32 bits are count
	binary.BigEndian.PutUint32(m, count)
	//Put Bearer and direction together
	m[4] = (bearer << 3) | (direction << 2)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	copy(m[8:], msg)

	mac, err = cmac.Sum(m, block, 16)
	if err != nil {
		return nil, err
	}
	// only get the most significant 32 bits to be mac value
	mac = mac[:4]
	return mac, nil
}

func NIA3() {

}
