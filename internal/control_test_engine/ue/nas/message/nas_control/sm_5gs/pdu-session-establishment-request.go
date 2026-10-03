/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	log "github.com/sirupsen/logrus"
)

func encodePlain(msg nas.Message) []byte {
	b, err := msg.MarshalBinary()
	if err != nil {
		log.Errorf("[UE][NAS] Encoding %s: %v", msg.MsgType(), err)
		return nil
	}
	return b
}
func GetPduSessionEstablishmentRequest(id uint8, sessionTypes ...uint8) []byte {
	sessionType := ie.PDUSessType_IPv4
	if len(sessionTypes) > 0 {
		sessionType = sessionTypes[0]
	}
	if sessionType < ie.PDUSessType_IPv4 || sessionType > ie.PDUSessType_IPv4v6 {
		return nil
	}
	return encodePlain(&nas.PDUSessEstReq{
		PDUSessId: id, PTI: 1,
		IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{Uplink: 255, Downlink: 255},
		PDUSessType:                    &ie.PDUSessType{Value: sessionType},
		ExtendedProtCfgOpts:            &ie.ExtendedProtCfgOpts{FromMs: &ie.ExtCfgOptFromMs{DNSV4Req: true, DNSV6Req: true}},
	})
}
func GetPduSessionReleaseRequest(id uint8) []byte {
	return encodePlain(&nas.PDUSessRelReq{PDUSessId: id, PTI: 1})
}
