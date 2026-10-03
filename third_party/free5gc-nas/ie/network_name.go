package ie

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"unicode/utf8"
)

// NwName is detailed in 10.5.3.5a Network Name, 24.008
// Ref: TS 23.038 6.1.2.2 CBS Packing , 6.2.1 GSM 7 bit Default Alphabet
// If a character number α is noted in the following way:
// b7 b6 b5 b4 b3 b2 b1
// αa αb αc αd αe αf αg
//
// 		bit number
//  7    6 5 4 3 2 1 0
//  octet number
//  1    2g 1a 1b 1c 1d 1e 1f 1g
//  2    3f 3g 2a 2b 2c 2d 2e 2f
//  3    4e 4f 4g 3a 3b 3c 3d 3e
//  4    5d 5e 5f 5g 4a 4b 4c 4d
//  5    6c 6d 6e 6f 6g 5a 5b 5c
//  6    7b 7c 7d 7e 7f 7g 6a 6b
//  7    8a 8b 8c 8d 8e 8f 8g 7a
//  8    10g 9a 9b 9c 9d 9e 9f 9g
//  .
//  .
//  81   93d 93e 93f 93g 92a 92b 92c 92d
//  82   0 0 0 0 0 93a 93b 93c
// The bit number zero is always transmitted first.

type NwName struct {
	Ext          uint8  // 8 -> 8 , 3 -> 3
	CodeScheme   uint8  // 7 -> 5 , 3 -> 3
	AddCi        uint8  // 4 -> 4 , 3 -> 3  CI: Country Initials
	NumSpareBits uint8  // 3 -> 1 , 3 -> 3
	TextStr      string // 8 -> 1 , 4 -> n
}

const (
	CodeScheme_Default uint8 = 0
	CodeScheme_UCS2    uint8 = 1

	AddCI_NoAddLetters uint8 = 0
	AddCI_ToAddLetters uint8 = 0
)

// Ref: https://embeddedfreak.wordpress.com/2008/10/09/decoding-gsm-sms-septets/
func decodeStrByCodeScheme(
	bs []byte,
	numSpareBitsInLastOctet uint8,
	codeScheme uint8,
) (string, error) {
	switch codeScheme {
	case CodeScheme_Default:
		numRealBits := len(bs)*8 - int(numSpareBitsInLastOctet)
		if numRealBits%7 != 0 {
			return "", errors.Errorf(
				"decodeStrByCodeScheme(): Value leads to a Text String whose length is not a multiple of 7 bits")
		}
		numBytes := numRealBits / 7
		result := make([]byte, numBytes)
		bsIdx := 0
		numHole := 1 // cycle: 1~8
		for resultIdx := 0; resultIdx < numBytes; resultIdx++ {
			// #bit to recover from prev encodedByte
			numBitToAppend := numHole - 1
			// appended bits should be 0 when no prev encodedByte
			appendedBits := byte(0)
			if bsIdx > 0 {
				appendedBits = bs[bsIdx-1] >> (8 - numBitToAppend)
			}
			baseVal := byte(0) // set baseVal to 0 when numBitToAppend==7
			if numBitToAppend != 7 {
				baseVal = bs[bsIdx] << byte(numBitToAppend) & 0x7f
				bsIdx++
			}
			result[resultIdx] = baseVal | appendedBits
			if numHole == 8 {
				numHole = 1
			} else {
				numHole++
			}
		}
		return string(result[:]), nil
	case CodeScheme_UCS2:
		if len(bs)%2 != 0 || numSpareBitsInLastOctet != 0 {
			return "", errors.Errorf("UCS-2 network name needs complete octets and 16-bit characters")
		}
		text := make([]rune, len(bs)/2)
		for j := range text {
			value := binary.BigEndian.Uint16(bs[2*j:])
			if value >= 0xd800 && value <= 0xdfff {
				return "", errors.Errorf("UCS-2 network name contains a surrogate")
			}
			text[j] = rune(value)
		}
		return string(text), nil
	default:
		return "", errors.Errorf("decodeStrByCodeScheme(): unsupported code scheme[%d]", codeScheme)
	}
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NwName) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("The NwName IE length(%d) is incorrect", len(b))
	}
	i.Ext = GetBit8(b[0])
	i.CodeScheme = Get3Bits75(b[0])
	i.AddCi = GetBit4(b[0])
	i.NumSpareBits = Get3Bits31(b[0])
	txtStr, err := decodeStrByCodeScheme(
		b[1:], i.NumSpareBits, i.CodeScheme,
	)
	if err != nil {
		return errors.Wrap(err, "NwName UnmarshalBinary()")
	}
	i.TextStr = txtStr
	return nil
}

// Ref: https://embeddedfreak.wordpress.com/2008/10/09/encoding-gsm-sms-septets/
func encodeStrByCodeScheme(
	s string,
	codeScheme uint8,
) ([]byte, uint8, error) {
	switch codeScheme {
	case CodeScheme_Default:
		numBytes := len(s) * 7 / 8
		numRemainBits := len(s) * 7 % 8
		numSpareBitsInLastOctet := 0
		if numRemainBits > 0 {
			numSpareBitsInLastOctet = 8 - numRemainBits
			numBytes += 1
		}
		bs := make([]byte, numBytes)
		bsIdx := 0
		numHole := 1 // cycle: 1~8
		for i := 0; i < len(s); i++ {
			baseVal := s[i] >> (byte(numHole - 1))
			holeVal := byte(0) // if no next char, hole val is 0
			if i+1 < len(s) {
				holeVal = s[i+1] << (8 - numHole)
			}
			if numHole == 8 {
				// Nothing to be filled in this round.
				// Don't add the resultByte to the result.
				numHole = 1
			} else {
				bs[bsIdx] = baseVal | holeVal
				bsIdx++
				numHole++
			}
		}
		return bs, uint8(numSpareBitsInLastOctet), nil
	case CodeScheme_UCS2:
		if !utf8.ValidString(s) {
			return nil, 0, errors.Errorf("invalid UTF-8 network name")
		}
		text := []rune(s)
		bs := make([]byte, 2*len(text))
		for j, value := range text {
			if value > 0xffff || value >= 0xd800 && value <= 0xdfff {
				return nil, 0, errors.Errorf("network name character cannot be encoded as UCS-2")
			}
			binary.BigEndian.PutUint16(bs[2*j:], uint16(value))
		}
		return bs, 0, nil
	default:
		return nil, 0, errors.Errorf("encodeStrByCodeScheme(): unsupported code scheme[%d]", codeScheme)
	}
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NwName) MarshalBinary() ([]byte, error) {
	i.Ext = 1
	bs, numSpareBits, err := encodeStrByCodeScheme(i.TextStr, i.CodeScheme)
	if err != nil {
		return nil, errors.Wrap(err, "NwName MarshalBinary()")
	}
	i.NumSpareBits = numSpareBits

	l := 1 + len(bs)
	b := make([]byte, l)
	b[0] = SetBit8(b[0], i.Ext)
	b[0] = Set3Bits75(b[0], i.CodeScheme)
	b[0] = SetBit4(b[0], i.AddCi)
	b[0] = Set3Bits31(b[0], i.NumSpareBits)
	if len(bs) > 0 {
		copy(b[1:], bs)
	}

	return b, nil
}
