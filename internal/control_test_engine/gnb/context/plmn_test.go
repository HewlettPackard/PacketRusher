/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 */
package context

import "testing"

func TestPLMNRejectsInvalidDigits(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc string }{
		{"", "01"}, {"01", "01"}, {"0001", "01"}, {"+01", "01"},
		{"001", ""}, {"001", "1"}, {"001", "1234"}, {"001", "+01"},
		{"001", "-1"}, {"00x", "01"}, {"001", "0x"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			gnb := &GNBContext{controlInfo: ControlInfo{mcc: tc.mcc, mnc: tc.mnc}}
			if got := gnb.GetMccAndMncInOctets(); got != nil {
				t.Fatalf("invalid PLMN encoded as %x", got)
			}
		})
	}
}
