/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024-2026 Valentin D'Emmanuele
 */
package mm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	log "my5G-RANTester/internal/log"
)

func ServiceRequest(ue *context.UEContext) []byte {
	psi := sessionStatus(ue)
	msg := &nas.SvcReq{
		Ngksi: &ue.UeSecurity.NgKsi, SvcType: &ie.SvcType{Value: 1},
		TMSI5GS:          &ie.MobileId5GS{TypeOfId: ie.IdType_5GS_TMSI, AllOneBits: 15, AMFSetID: ue.GetAmfSetId(), AMFPointer: ue.GetAmfPointer(), TMSI5G: ue.GetTMSI5G()},
		UplinkDataStatus: &ie.UplinkDataStatus{Psi: psi}, PDUSessStatus: &ie.PDUSessStatus{Psi: psi},
	}
	b := encodePlain(msg)
	if b == nil {
		return nil
	}
	encrypted, err := ue.NASSecurityContext().NASEncrypt(nas.DirectionUplink, b)
	if err != nil {
		log.Errorf("[UE][NAS] Encrypting service container: %v", err)
		return nil
	}
	msg.NASMsgCntr = &ie.NASMsgCntr{Contents: encrypted}
	msg.UplinkDataStatus, msg.PDUSessStatus = nil, nil
	return encodePlain(msg)
}
