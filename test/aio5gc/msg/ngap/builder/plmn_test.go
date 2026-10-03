// SPDX-License-Identifier: Apache-2.0
package builder

import (
	"encoding/hex"
	"testing"

	"github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
)

func TestNGSetupResponsePLMNWireVectors(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc, wire string }{
		{"208", "93", "02f839"}, {"208", "010", "020801"},
		{"001", "001", "000110"}, {"001", "01", "00f110"},
		{"999", "070", "990907"}, {"999", "07", "99f970"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			plmn := models.PlmnId{Mcc: tc.mcc, Mnc: tc.mnc}
			served := models.PlmnIdNid{Mcc: tc.mcc, Mnc: tc.mnc}
			response := BuilNGSetupResponse("amf.5gc.3gppnetwork.org", "196673",
				[]models.Guami{{PlmnId: &served, AmfId: "196673"}},
				[]models.Nrf_NFMgmt_PlmnSnssai{{PlmnId: &plmn, SNssaiList: []models.ExtSnssai{{Sst: 1, Sd: "010203"}}}}, 100)
			encoded, err := response.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := message.Parse(encoded)
			if err != nil {
				t.Fatal(err)
			}
			setup := decoded.(*message.NGSetupResponse)
			guami := setup.ServedGUAMIList.List[0].GUAMI.PLMNIdentity.Value
			supported := setup.PLMNSupportList.List[0].PLMNIdentity.Value
			for name, wire := range map[string][]byte{"served GUAMI": guami, "supported PLMN": supported} {
				if got := hex.EncodeToString(wire); got != tc.wire {
					t.Fatalf("%s PLMN: got %s, want %s", name, got, tc.wire)
				}
			}
		})
	}
}
