/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func encodePlain(msg nas.Message) []byte {
	b, err := msg.MarshalBinary()
	if err != nil {
		log.Errorf("[UE][NAS] Encoding %s: %v", msg.MsgType(), err)
		return nil
	}
	return b
}

func GetRegistrationRequest(registrationType uint8, requestedNSSAI *ie.NSSAI, uplinkDataStatus *ie.UplinkDataStatus, capability bool, ue *context.UEContext) []byte {
	identity := ue.GetSuci()
	if ue.Get5gGuti() != nil {
		identity = *ue.Get5gGuti()
	}
	msg := &nas.RegReq{
		RegType5GS: &ie.RegType5GS{Value: registrationType, FOR_Pending: true},
		Ngksi:      &ue.UeSecurity.NgKsi, MobileId5GS: &identity,
		UESecCapability: ue.GetUeSecurityCapability(), ReqNSSAI: requestedNSSAI,
		UplinkDataStatus: uplinkDataStatus,
	}
	if capability {
		msg.Capability5GMM = &ie.Capability5GMM{Length: 1, N3Data: true, LPP: true, HOAttach: true, S1Mode: true}
	}
	psi := sessionStatus(ue)
	active := false
	for _, set := range psi.PSI {
		active = active || set
	}
	if active {
		msg.UplinkDataStatus = &ie.UplinkDataStatus{Psi: psi}
		msg.PDUSessStatus = &ie.PDUSessStatus{Psi: psi}
	}
	b := encodePlain(msg)
	if b == nil || !active {
		return b
	}
	encrypted, err := ue.NASSecurityContext().NASEncrypt(nas.DirectionUplink, b)
	if err != nil {
		log.Errorf("[UE][NAS] Encrypting registration container: %v", err)
		return nil
	}
	msg.NASMsgCntr = &ie.NASMsgCntr{Contents: encrypted}
	msg.UplinkDataStatus, msg.PDUSessStatus = nil, nil
	return encodePlain(msg)
}

func sessionStatus(ue *context.UEContext) ie.Psi {
	var psi ie.Psi
	for i, session := range ue.PduSession {
		if session != nil && i+1 < len(psi.PSI) {
			psi.PSI[i+1] = true
		}
	}
	return psi
}
