/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
)

func InitialContextSetupRequest(nas []byte, ue *context.UEContext, amf *context.AMFContext) ([]byte, error) {
	msg, err := buildInitialContextSetupRequest(nas, ue, amf)
	if err != nil {
		return nil, err
	}
	return msg.MarshalBinary()
}
func buildInitialContextSetupRequest(nas []byte, ue *context.UEContext, amf *context.AMFContext) (*message.InitialContextSetupRequest, error) {
	caps := ue.GetSecurityCapability()
	enc, integrity := byte(0), byte(0)
	if caps.EA1_128_5G {
		enc |= 128
	}
	if caps.EA2_128_5G {
		enc |= 64
	}
	if caps.EA3_128_5G {
		enc |= 32
	}
	if caps.IA1_128_5G {
		integrity |= 128
	}
	if caps.IA2_128_5G {
		integrity |= 64
	}
	if caps.IA3_128_5G {
		integrity |= 32
	}
	msg := &message.InitialContextSetupRequest{
		AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: ue.GetRanNgapId()},
		GUAMI: convert.GUAMIToNGAP(amf.GetServedGuami()[0]), AllowedNSSAI: new(ie.AllowedNSSAI),
		UESecurityCapabilities: &ie.UESecurityCapabilities{
			NRencryptionAlgorithms:             &ie.NRencryptionAlgorithms{Value: aper.BitString{Bytes: []byte{enc, 0}, BitLength: 16}},
			NRintegrityProtectionAlgorithms:    &ie.NRintegrityProtectionAlgorithms{Value: aper.BitString{Bytes: []byte{integrity, 0}, BitLength: 16}},
			EUTRAencryptionAlgorithms:          &ie.EUTRAencryptionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
			EUTRAintegrityProtectionAlgorithms: &ie.EUTRAintegrityProtectionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
		},
		SecurityKey: &ie.SecurityKey{Value: aper.BitString{Bytes: ue.GetSecurityContext().GetKGNB(), BitLength: 256}},
	}
	for _, slice := range amf.GetSupportedPlmnSnssai()[0].SNssaiList {
		msg.AllowedNSSAI.List = append(msg.AllowedNSSAI.List, ie.AllowedNSSAIItem{SNSSAI: convert.SNSSAIToNGAP(models.Snssai{Sst: slice.Sst, Sd: slice.Sd})})
	}
	if nas != nil {
		msg.NASPDU = &ie.NASPDU{Value: nas}
	}
	return msg, nil
}
