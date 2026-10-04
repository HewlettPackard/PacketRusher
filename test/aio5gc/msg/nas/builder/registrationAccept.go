/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func RegistrationAccept(ue *context.UEContext) ([]byte, error) {
	msg, err := buildRegistrationAccept(ue)
	if err != nil {
		return nil, err
	}
	return codec.Encode(ue, msg, nas.SecHdrTypeIntegrityProtectedAndCiphered)
}
func buildRegistrationAccept(ue *context.UEContext) (*nas.RegAccept, error) {
	guti := new(ie.MobileId5GS)
	if err := guti.FromGUTIStr(ue.GetGuti()); err != nil {
		return nil, err
	}
	tai := ue.GetUserLocationInfo().Tai
	snssai := ue.GetDefaultSNssai()
	timer := new(ie.GPRSTimer3)
	timer.Set(60)
	return &nas.RegAccept{
		RegResult5GS: &ie.RegResult5GS{Value: ie.RegResult_3gpp}, GUTI5G: guti,
		TAIList:             &ie.TrackingAreaIdList5GS{TAI: []ie.TrackingAreaId5GS{{PlmnId: ie.PlmnId{MCC: tai.PlmnId.Mcc, MNC: tai.PlmnId.Mnc}, TAC: tai.Tac}}},
		AllowedNSSAI:        &ie.NSSAI{SNSSAIs: []ie.SNSSAI{{SST: uint8(snssai.Sst), SD: snssai.Sd}}},
		ConfiguredNSSAI:     &ie.NSSAI{SNSSAIs: []ie.SNSSAI{{SST: uint8(snssai.Sst), SD: snssai.Sd}}},
		NwFeatureSupport5GS: &ie.NwFeatureSupport5GS{Length: 2, IMSVoPS3GPP: true}, T3512Value: timer,
	}, nil
}
