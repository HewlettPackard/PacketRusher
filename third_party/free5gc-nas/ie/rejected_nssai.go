package ie

import (
	"encoding/hex"

	"github.com/pkg/errors"
)

// RejectedNSSAI is detailed in 9.11.3.46 Rejected NSSAI, 24.501

const (
	SNSSAIRej_NotAvailInCurrPLMN       uint8 = 0x00
	SNSSAIRej_NotAvailInCurrRegArea    uint8 = 0x01
	SNSSAIRej_NotAvailAuthAuthorFailed uint8 = 0x02
)

type rejSNSSAI struct {
	cause uint8
	sst   uint8
	sd    string // HEX String -> [3] byte
}

type RejectedNSSAI struct {
	RejSnssai []rejSNSSAI
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RejectedNSSAI) UnmarshalBinary(b []byte) error {
	ttlLen := uint8(len(b))
	var offset uint8 = 0

	for ttlLen > 0 {
		length := Get4Bits85(b[offset])
		switch length {
		case 1:
			r := rejSNSSAI{cause: Get4Bits41(b[offset]), sst: b[offset+1]}
			i.RejSnssai = append(i.RejSnssai, r)
		case 4:
			r := rejSNSSAI{cause: Get4Bits41(b[offset]), sst: b[offset+1], sd: hex.EncodeToString(b[offset+2 : offset+5])}
			i.RejSnssai = append(i.RejSnssai, r)
		default:
			return errors.Errorf("RejectedNSSAI: bad length (%d)?", Get4Bits85(b[offset]))
		}
		if len(i.RejSnssai) >= 8 {
			// The number of rejected S-NSSAI(s) shall not exceed eight.
			break
		}
		ttlLen -= length + 1
		offset += length + 1
	}
	return nil
}

func (i *RejectedNSSAI) Append(cause, sst uint8, sd string) {
	r := rejSNSSAI{cause: cause, sst: sst, sd: sd}
	i.RejSnssai = append(i.RejSnssai, r)
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RejectedNSSAI) MarshalBinary() ([]byte, error) {
	// Get total length
	var ttlLen uint8 = 0
	for _, r := range i.RejSnssai {
		if r.sd != "" {
			ttlLen += 5
		} else {
			ttlLen += 2
		}
	}

	b := make([]byte, int(ttlLen))

	// Fill in the content
	offset := 0
	for i, r := range i.RejSnssai {
		if i >= 8 {
			// The number of rejected S-NSSAI(s) shall not exceed eight.
			break
		}
		if r.sd != "" {
			b[offset] = Set4Bits85(b[offset], 4)
			b[offset] = Set4Bits41(b[offset], r.cause)
			offset++
			b[offset] = r.sst
			offset++
			if _, err := hex.Decode(b[offset:offset+3], []byte(r.sd)); err != nil {
				return nil, errors.Errorf("RejectedNSSAI: bad r.SD(%s)", r.sd)
			}
			offset += 3
		} else {
			b[offset] = Set4Bits85(b[offset], 1)
			b[offset] = Set4Bits41(b[offset], r.cause)
			offset++
			b[offset] = r.sst
			offset += 1
		}
	}

	return b, nil
}
