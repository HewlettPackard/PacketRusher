// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"encoding/hex"
	"testing"
)

// Literal TS 38.413 vectors, independently checked for encoding and decoding.
// In particular, the three-digit NGAP MNC order must not use NAS PLMN packing.
func TestPLMNWireVectors(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc, wire string }{
		{"208", "93", "02f839"}, {"208", "123", "021832"},
		{"208", "010", "020801"}, {"001", "01", "00f110"},
		{"001", "001", "000110"}, {"001", "000", "000100"},
		{"310", "260", "132006"}, {"310", "26", "13f062"},
		{"999", "070", "990907"}, {"999", "07", "99f970"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			literal, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := EncodePLMN(tc.mcc, tc.mnc)
			if err != nil || hex.EncodeToString(encoded) != tc.wire {
				t.Fatalf("EncodePLMN: got %x, %v; want %s", encoded, err, tc.wire)
			}
			mcc, mnc, err := DecodePLMN(literal)
			if err != nil || mcc != tc.mcc || mnc != tc.mnc {
				t.Fatalf("DecodePLMN: got %s/%s, %v; want %s/%s", mcc, mnc, err, tc.mcc, tc.mnc)
			}
		})
	}
}

func TestPLMNRejectsMalformedInput(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc string }{
		{"", "01"}, {"01", "01"}, {"0001", "01"}, {"+01", "01"},
		{"001", ""}, {"001", "1"}, {"001", "1234"}, {"001", "+01"},
		{"001", "-1"}, {"00x", "01"}, {"001", "0x"},
	} {
		if b, err := EncodePLMN(tc.mcc, tc.mnc); err == nil || b != nil {
			t.Fatalf("EncodePLMN accepted %q/%q as %x", tc.mcc, tc.mnc, b)
		}
	}
	for _, wire := range []string{"", "00", "0001", "00011000", "a00110", "0a0110", "000a10", "00b110", "00011b", "0001b0", "f00110", "00f1f0"} {
		b, err := hex.DecodeString(wire)
		if err != nil {
			t.Fatal(err)
		}
		if mcc, mnc, err := DecodePLMN(b); err == nil || mcc != "" || mnc != "" {
			t.Fatalf("DecodePLMN accepted %s as %s/%s, %v", wire, mcc, mnc, err)
		}
	}
}
