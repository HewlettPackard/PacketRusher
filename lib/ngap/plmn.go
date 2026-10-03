// SPDX-License-Identifier: Apache-2.0
package ngap

import "fmt"

// EncodePLMN implements TS 38.413 section 9.3.3.5: MCC then MNC digits,
// low nibble first, with a filler before a two-digit MNC. NAS PLMN encoding
// differs for three-digit MNCs and must not be reused for NGAP identities.
func EncodePLMN(mcc, mnc string) ([]byte, error) {
	if len(mcc) != 3 || (len(mnc) != 2 && len(mnc) != 3) {
		return nil, fmt.Errorf("MCC must have three digits and MNC two or three digits")
	}
	for _, digit := range mcc + mnc {
		if digit < '0' || digit > '9' {
			return nil, fmt.Errorf("MCC and MNC must contain decimal digits")
		}
	}
	encoded := []byte{(mcc[1]-'0')<<4 | (mcc[0] - '0'), mcc[2] - '0', 0}
	if len(mnc) == 2 {
		encoded[1] |= 0xf0
		encoded[2] = (mnc[1]-'0')<<4 | (mnc[0] - '0')
	} else {
		encoded[1] |= (mnc[0] - '0') << 4
		encoded[2] = (mnc[2]-'0')<<4 | (mnc[1] - '0')
	}
	return encoded, nil
}

// DecodePLMN preserves leading zeros and the two-/three-digit MNC distinction.
func DecodePLMN(encoded []byte) (mcc, mnc string, err error) {
	if len(encoded) != 3 {
		return "", "", fmt.Errorf("NGAP PLMN must contain exactly three octets")
	}
	digits := []byte{encoded[0] & 0xf, encoded[0] >> 4, encoded[1] & 0xf,
		encoded[1] >> 4, encoded[2] & 0xf, encoded[2] >> 4}
	twoDigitMNC := digits[3] == 0xf
	for i, digit := range digits {
		if i == 3 && twoDigitMNC {
			continue
		}
		if digit > 9 {
			return "", "", fmt.Errorf("NGAP PLMN contains a non-decimal digit")
		}
		digits[i] = '0' + digit
	}
	mcc = string(digits[:3])
	if twoDigitMNC {
		mnc = string(digits[4:])
	} else {
		mnc = string(digits[3:])
	}
	return mcc, mnc, nil
}
