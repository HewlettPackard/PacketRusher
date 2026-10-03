/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"encoding/hex"
	"net/netip"
	"testing"

	nasIE "github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	ngapIE "github.com/free5gc/ngap/ie"
	ngap "github.com/free5gc/ngap/message"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbctx "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/nas_transport"
	uectx "my5G-RANTester/internal/control_test_engine/ue/context"
	corectx "my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
)

// NGAP packs MCC then MNC digits in order (TS 38.413 9.3.3.5), while NAS
// places MNC digit 3 in octet 2. Assert independent literals for both protocols;
// an encoder/decoder round trip alone hides a shared PLMN packing error.
func TestNGAPPLMNAuthentication(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc, ngapWire, nasWire string }{
		{"208", "93", "02f839", "02f839"},
		{"208", "123", "021832", "023821"},
		{"208", "010", "020801", "020810"},
		{"001", "01", "00f110", "00f110"},
		{"001", "001", "000110", "001100"},
		{"001", "000", "000100", "000100"},
		{"310", "260", "132006", "130062"},
		{"310", "26", "13f062", "13f062"},
		{"999", "070", "990907", "990970"},
		{"999", "07", "99f970", "99f970"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			gnb := new(gnbctx.GNBContext)
			gnb.NewRanGnbContext("000008", tc.mcc, tc.mnc, "000001", "01", "000001", netip.MustParseAddrPort("127.0.0.1:9487"), netip.MustParseAddrPort("127.0.0.1:2152"))
			wire, err := nas_transport.GetInitialUEMessage(1, []byte{0x7e, 0, 0x41}, nil, gnb)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := ngap.Parse(wire)
			if err != nil {
				t.Fatal(err)
			}
			location := decoded.(*ngap.InitialUEMessage).UserLocationInformation.Choice.(*ngapIE.UserLocationInformationNR)
			if got := hex.EncodeToString(location.TAI.PLMNIdentity.Value); got != tc.ngapWire {
				t.Fatalf("NGAP PLMN wire = %s, want %s", got, tc.ngapWire)
			}
			model := convert.NRLocationToModels(location)
			if model.Tai.PlmnId.Mcc != tc.mcc || model.Tai.PlmnId.Mnc != tc.mnc {
				t.Fatalf("decoded location = %v, want %s/%s", model.Tai.PlmnId, tc.mcc, tc.mnc)
			}
			security := new(corectx.SecurityContext)
			security.SetAuthSubscription("00112233445566778899AABBCCDDEEFF", "00112233445566778899AABBCCDDEEFF", "", "8000", "000000000000")
			core := new(corectx.UEContext)
			core.SetSecurityContext(security)
			core.SetNgKsi(nasIE.NASKeySetId{Tsc: nasIE.SecCtxTypeNative, Ksi: 0})
			core.SetUserLocationInfo(model)
			authWire, err := AuthenticationRequest(core)
			if err != nil {
				t.Fatal(err)
			}
			authParsed, err := nas.Parse(authWire, nil)
			if err != nil {
				t.Fatal(err)
			}
			challenge := authParsed.(*nas.AuthReq)
			cfg := config.Config{}
			cfg.Ue.Ciphering.Nea2, cfg.Ue.Integrity.Nia2 = true, true
			ue := new(uectx.UEContext)
			ue.NewRanUeContext("000000120", cfg.GetUESecurityCapability(), "00112233445566778899AABBCCDDEEFF", "00112233445566778899AABBCCDDEEFF", "", "8000", "000000000000", tc.mcc, tc.mnc, sidf.HomeNetworkPublicKey{ProtectionScheme: "0", PublicKeyID: "0"}, "0000", "internet", 1, "000001", config.TunnelDisabled, nil, nil, 1)
			ue.SetAmfMccAndMnc(tc.mcc, tc.mnc)
			if got := hex.EncodeToString(ue.GetMccAndMncInOctets()); got != tc.nasWire {
				t.Fatalf("NAS UE PLMN = %s, want %s", got, tc.nasWire)
			}
			res, status := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, challenge.AuthParamRAND5GAuthChlg.Rand, ue.UeSecurity.Snn, challenge.AuthParamAUTN5GAuthChlg.Autn)
			if status != "successful" || hex.EncodeToString(res) != security.GetXresStar() {
				t.Fatalf("authentication failed for %s/%s: status %s, UE RES* %x, core XRES* %s", tc.mcc, tc.mnc, status, res, security.GetXresStar())
			}
		})
	}
}
