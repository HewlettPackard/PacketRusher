/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"encoding/hex"
	"fmt"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/tools"
)

func AuthenticationRequest(ue *context.UEContext) ([]byte, error) {
	plmn := ue.GetUserLocationInfo().Tai.PlmnId
	serving := fmt.Sprintf("5G:mnc%03s.mcc%s.3gppnetwork.org", plmn.Mnc, plmn.Mcc)
	result, problem := tools.AuthProcedure(ue.GetSecurityContext().GetAuthSubscription(), serving)
	if problem != nil {
		return nil, fmt.Errorf("AKA generation failed: %+v", problem)
	}
	ue.GetSecurityContext().SetXresStar(result.AuthenticationVector.XresStar)
	_, kseaf, err := tools.DeriveHXRES(result, serving)
	if err != nil {
		return nil, err
	}
	ue.GetSecurityContext().SetKseaf(kseaf)
	rand, err := hex.DecodeString(result.AuthenticationVector.Rand)
	if err != nil {
		return nil, err
	}
	autn, err := hex.DecodeString(result.AuthenticationVector.Autn)
	if err != nil {
		return nil, err
	}
	ksi := ue.GetNgKsi()
	msg := &nas.AuthReq{
		Ngksi: &ksi, ABBA: &ie.ABBA{Abba: []byte{0, 0}},
		AuthParamRAND5GAuthChlg: &ie.AuthParamRAND{Rand: rand}, AuthParamAUTN5GAuthChlg: &ie.AuthParamAUTN{Autn: autn},
	}
	return msg.MarshalBinary()
}
