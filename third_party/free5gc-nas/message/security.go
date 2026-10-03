package message

import (
	"encoding/hex"

	"github.com/pkg/errors"

	"github.com/free5gc/nas/ie"
	"github.com/free5gc/nas/internal/algo"
)

// TS 33.501 Annex A.8 Algorithm distinguisher For Knas_int Knas_enc
const (
	NNASEncAlg uint8 = 0x01
	NNASIntAlg uint8 = 0x02
	NRRCEncAlg uint8 = 0x03
	NRRCIntAlg uint8 = 0x04
	NUpEncAlg  uint8 = 0x05
	NUpIntAlg  uint8 = 0x06
)

// TS 33.501 5.11.1.1 Algorithm identifier values For integrity algorithm
const (
	AlgIntegrity128NIA0 ie.AlgIntegrity = 0x00 // NULL
	AlgIntegrity128NIA1 ie.AlgIntegrity = 0x01 // 128-Snow3G
	AlgIntegrity128NIA2 ie.AlgIntegrity = 0x02 // 128-AES
	AlgIntegrity128NIA3 ie.AlgIntegrity = 0x03 // 128-ZUC
)

// TS 33.501 5.11.1.1 Algorithm identifier values For ciphering algorithm
const (
	AlgCiphering128NEA0 ie.AlgCiphering = 0x00 // NULL
	AlgCiphering128NEA1 ie.AlgCiphering = 0x01 // 128-Snow3G
	AlgCiphering128NEA2 ie.AlgCiphering = 0x02 // 128-AES
	AlgCiphering128NEA3 ie.AlgCiphering = 0x03 // 128-ZUC
)

type Direction uint8

// 1bit
const (
	DirectionUplink   Direction = 0x00
	DirectionDownlink Direction = 0x01
)

type BearerType uint8

// 5bits
const (
	OnlyOneBearer BearerType = 0x00
	Bearer3GPP    BearerType = 0x01
	BearerNon3GPP BearerType = 0x02
)

type AccessType uint8

// TS 33501 Annex A.0 Access type distinguisher For Kgnb Kn3iwf
const (
	AccessType3GPP    AccessType = 0x01
	AccessTypeNon3GPP AccessType = 0x02
)

const (
	SecHdrLen  uint8 = 7
	MacLen     uint8 = 4
	KnasEncLen uint8 = 16
	KnasIntLen uint8 = 16
)

type Side int

const (
	UESide Side = iota
	CoreNetworkSide
)

type SecProtectedHdr struct {
	SecHdrType     SecHdrType
	MAC            []byte
	SequenceNumber uint8
}

func (h *SecProtectedHdr) UnmarshalBinary(b []byte) error {
	if len(b) < int(SecHdrLen) {
		return errors.Errorf("secProtectedHdr: len(b)==%d, < %d", len(b), SecHdrLen)
	}
	if b[0] != byte(Epd5GSMobilityMgmtMsg) {
		return errors.Errorf("secProtectedHdr's EPD (%d) != %d, %s", b[0], Epd5GSMobilityMgmtMsg,
			hex.EncodeToString(b))
	}
	h.SecHdrType = SecHdrType(b[1] & 0x0f)
	h.MAC = []byte{b[2], b[3], b[4], b[5]}
	h.SequenceNumber = b[6]
	return nil
}

func (h *SecProtectedHdr) MarshalBinary() ([]byte, error) {
	if len(h.MAC) != int(MacLen) {
		return nil, errors.Errorf("secProtectedHdr: len(h.MAC)=%d, != %d", len(h.MAC), MacLen)
	}
	b := []byte{
		byte(Epd5GSMobilityMgmtMsg), byte(h.SecHdrType),
		h.MAC[0], h.MAC[1], h.MAC[2], h.MAC[3],
		h.SequenceNumber,
	}
	return b, nil
}

type SecCtx struct {
	Side          Side
	Bearer        BearerType
	UplinkCount   *Count
	DownlinkCount *Count
	CipheringAlg  ie.AlgCiphering
	IntegrityAlg  ie.AlgIntegrity
	KnasEnc       [KnasEncLen]uint8
	KnasInt       [KnasIntLen]uint8
}

// TODO: wrap the nas security context, maybe for UE side
func NewSecCtx(side Side, bearer BearerType, cipheringAlg ie.AlgCiphering,
	integrityAlg ie.AlgIntegrity, knasEnc, knasInt []byte,
) *SecCtx {
	secCtx := new(SecCtx)
	secCtx.Side = side
	secCtx.Bearer = bearer
	secCtx.UplinkCount = &Count{Count: 0}
	secCtx.DownlinkCount = &Count{Count: 0}
	secCtx.CipheringAlg = cipheringAlg
	secCtx.IntegrityAlg = integrityAlg
	copy(secCtx.KnasEnc[:], knasEnc)
	copy(secCtx.KnasInt[:], knasInt)

	return secCtx
}

func (s *SecCtx) Clone() *SecCtx {
	secCtx := SecCtx{
		Side:         s.Side,
		Bearer:       s.Bearer,
		CipheringAlg: s.CipheringAlg,
		IntegrityAlg: s.IntegrityAlg,
	}
	if s.UplinkCount != nil {
		secCtx.UplinkCount = &Count{Count: s.UplinkCount.Count}
	}
	if s.DownlinkCount != nil {
		secCtx.DownlinkCount = &Count{Count: s.DownlinkCount.Count}
	}
	copy(secCtx.KnasEnc[:], s.KnasEnc[:])
	copy(secCtx.KnasInt[:], s.KnasInt[:])
	return &secCtx
}

func (s *SecCtx) SetEncAlgo(algo ie.AlgCiphering) {
	s.CipheringAlg = algo
}

func (s *SecCtx) GetEncAlgo() ie.AlgCiphering {
	return s.CipheringAlg
}

func (s *SecCtx) SetIntAlgo(algo ie.AlgIntegrity) {
	s.IntegrityAlg = algo
}

func (s *SecCtx) GetIntAlgo() ie.AlgIntegrity {
	return s.IntegrityAlg
}

func (s *SecCtx) CountReset(direction Direction) {
	if direction == DirectionDownlink {
		s.DownlinkCount.Set(0, 0)
	} else {
		s.UplinkCount.Set(0, 0)
	}
}

func (s *SecCtx) CountAddOne(direction Direction) {
	if direction == DirectionDownlink {
		s.DownlinkCount.AddOne()
	} else {
		s.UplinkCount.AddOne()
	}
}

func (s *SecCtx) GetCountSQN(direction Direction) uint8 {
	if direction == DirectionDownlink {
		return s.DownlinkCount.SQN()
	} else {
		return s.UplinkCount.SQN()
	}
}

func (s *SecCtx) NASEncrypt(direction Direction, payload []byte) ([]byte, error) {
	if s.Bearer > BearerNon3GPP {
		return nil, errors.Errorf("unknown Bearer[%d]", s.Bearer)
	}
	if direction > 1 {
		return nil, errors.Errorf("direction is beyond 1 bits")
	}
	if payload == nil {
		return nil, errors.Errorf("nas Payload is nil")
	}

	var count *Count
	if direction == DirectionDownlink {
		count = s.DownlinkCount
	} else {
		count = s.UplinkCount
	}

	switch s.CipheringAlg {
	case AlgCiphering128NEA0:
		return payload, nil
	case AlgCiphering128NEA1:
		output, err := algo.NEA1(s.KnasEnc, count.Get(), uint32(s.Bearer), uint32(direction),
			payload, uint32(len(payload))*8)
		if err != nil {
			return nil, err
		}
		return output, nil
	case AlgCiphering128NEA2:
		output, err := algo.NEA2(s.KnasEnc, count.Get(), uint8(s.Bearer), uint8(direction), payload)
		if err != nil {
			return nil, err
		}
		return output, nil
	case AlgCiphering128NEA3:
		output, err := algo.NEA3(s.KnasEnc, count.Get(), uint8(s.Bearer),
			uint8(direction), payload, uint32(len(payload))*8)
		if err != nil {
			return nil, err
		}
		return output, nil
	default:
		return nil, errors.Errorf("unknown Algorithm Identity[%d]", s.CipheringAlg)
	}
}

func (s *SecCtx) NASMacCalculate(direction Direction, msg []byte) ([]byte, error) {
	if s.Bearer > BearerNon3GPP {
		return nil, errors.Errorf("unknown Bearer[%d]", s.Bearer)
	}
	if direction > 1 {
		return nil, errors.Errorf("direction is beyond 1 bits")
	}
	if msg == nil {
		return nil, errors.Errorf("nas Payload is nil")
	}

	var count *Count
	if direction == DirectionDownlink {
		count = s.DownlinkCount
	} else {
		count = s.UplinkCount
	}

	switch s.IntegrityAlg {
	case AlgIntegrity128NIA0:
		// TS 33.501 D.1: The NIA0 algorithm shall be implemented in such way that
		// it shall generate a 32 bit MAC-I/NAS-MAC and XMAC-I/XNAS-M of all zeroes
		return make([]byte, 4), nil
	case AlgIntegrity128NIA1:
		return algo.NIA1(s.KnasInt, count.Get(), byte(s.Bearer), uint32(direction), msg, uint64(len(msg))*8)
	case AlgIntegrity128NIA2:
		return algo.NIA2(s.KnasInt, count.Get(), byte(s.Bearer), uint8(direction), msg)
	case AlgIntegrity128NIA3:
		return algo.NIA3(s.KnasInt, count.Get(), byte(s.Bearer), uint8(direction), msg, uint32(len(msg))*8)
	default:
		return nil, errors.Errorf("unknown Algorithm Identity[%d]", s.IntegrityAlg)
	}
}
