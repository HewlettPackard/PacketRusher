// SPDX-License-Identifier: Apache-2.0
package config

import (
	"bytes"
	"testing"
)

func TestGNBIdentityWireBoundaries(t *testing.T) {
	for _, tc := range []struct {
		id       string
		bits     uint8
		cell     uint16
		gnb, nci []byte
	}{
		{"000001", 0, 0, []byte{0, 0, 1}, []byte{0, 0, 1, 0, 0}},
		{"3FFFFF", 22, 16383, []byte{255, 255, 252}, []byte{255, 255, 255, 255, 240}},
		{"7FFFFF", 23, 8191, []byte{255, 255, 254}, []byte{255, 255, 255, 255, 240}},
		{"FFFFFF", 24, 4095, []byte{255, 255, 255}, []byte{255, 255, 255, 255, 240}},
		{"1FFFFFF", 25, 2047, []byte{255, 255, 255, 128}, []byte{255, 255, 255, 255, 240}},
		{"FFFFFFF", 28, 255, []byte{255, 255, 255, 240}, []byte{255, 255, 255, 255, 240}},
		{"FFFFFFFF", 32, 15, []byte{255, 255, 255, 255}, []byte{255, 255, 255, 255, 240}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			id, err := ParseGNBIdentity(tc.id, tc.bits, tc.cell)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(id.Bytes(), tc.gnb) || !bytes.Equal(id.CellBytes(), tc.nci) {
				t.Fatalf("wire: gNB=%x NCI=%x, want %x / %x", id.Bytes(), id.CellBytes(), tc.gnb, tc.nci)
			}
		})
	}
}

func TestGNBIdentityRejectsUnrepresentableConfiguration(t *testing.T) {
	for _, tc := range []struct {
		id   string
		bits uint8
		cell uint16
	}{
		{"1", 21, 0}, {"1", 33, 0}, {"400000", 22, 0}, {"100000000", 32, 0},
		{"1", 32, 16}, {"1", 22, 16384}, {"", 24, 0}, {"0x1", 24, 0},
		{"+1", 24, 0}, {"-1", 24, 0}, {"1 ", 24, 0}, {"wrong", 24, 0},
	} {
		if _, err := ParseGNBIdentity(tc.id, tc.bits, tc.cell); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
	}
	p := PlmnList{GnbId: "FFFFFFFE", GnbIDLength: 32}
	if id, err := p.GNBIDAt(1); err != nil || id != "FFFFFFFF" {
		t.Fatalf("last ID: %q %v", id, err)
	}
	for _, offset := range []int{-1, 2, 100} {
		if _, err := p.GNBIDAt(offset); err == nil {
			t.Fatalf("accepted overflowing offset %d", offset)
		}
	}
	if id, err := (PlmnList{GnbId: "abcdef"}).GNBIDAt(0); err != nil || id != "ABCDEF" {
		t.Fatalf("canonical ID: %q %v", id, err)
	}
}
