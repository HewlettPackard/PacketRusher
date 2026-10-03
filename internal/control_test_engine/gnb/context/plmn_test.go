/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"encoding/hex"
	"testing"
)

func TestNGAPPLMNEncodingAndContextDecoding(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc, wire string }{
		{"208", "93", "02f839"}, {"208", "010", "020801"},
		{"001", "001", "000110"}, {"001", "01", "00f110"},
		{"999", "070", "990907"}, {"999", "07", "99f970"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			gnb := &GNBContext{controlInfo: ControlInfo{mcc: tc.mcc, mnc: tc.mnc}}
			if got := hex.EncodeToString(gnb.GetMccAndMncInOctets()); got != tc.wire {
				t.Fatalf("gNB PLMN: got %s, want %s", got, tc.wire)
			}
			amf := &GNBAmf{}
			amf.AddedPlmn(tc.wire)
			if mcc, mnc := amf.GetPlmnSupport(0); mcc != tc.mcc || mnc != tc.mnc {
				t.Fatalf("AMF PLMN support: got %s/%s, want %s/%s", mcc, mnc, tc.mcc, tc.mnc)
			}
			ue := &GNBUe{}
			ue.CreateUeContext(tc.wire, "", nil, nil, nil)
			if ue.context.mobilityInfo.mcc != tc.mcc || ue.context.mobilityInfo.mnc != tc.mnc {
				t.Fatalf("UE mobility PLMN: got %+v, want %s/%s", ue.context.mobilityInfo, tc.mcc, tc.mnc)
			}
		})
	}
}

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
