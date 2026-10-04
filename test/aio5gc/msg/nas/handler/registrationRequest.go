/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"
	"strings"

	ie "github.com/free5gc/nas/ie"

	"github.com/free5gc/util/fsm"

	nas "github.com/free5gc/nas/message"
	"github.com/sirupsen/logrus"
)

func RegistrationRequest(nasReq *nas.RegReq, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.Deregistered:
		err = DefaultRegistrationRequest(nasReq, amf, ue, gnb)

	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for RegistrationRequest", ue.GetState().Current())
	}
	return err
}

func DefaultRegistrationRequest(nasReq *nas.RegReq, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {

	err := ue.GetUeFsm().SendEvent(ue.GetState(), context.RegistrationRequest, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		return err
	}
	regType := nasReq.RegType5GS.Value
	if regType != ie.RegType_InitialReg {
		return errors.New("[5GC][NAS] Received unsupported registration type")
	}

	ngKsi := *nasReq.Ngksi
	if ngKsi.Tsc != ie.SecCtxTypeNative {
		return errors.New("[5GC] Unsupported KSI context type")
	}
	if ngKsi.Ksi == ie.NASKeyNA {
		ngKsi.Ksi = 0
	}
	ue.SetSecurityCapability(nasReq.UESecCapability)
	ue.SetNgKsi(ngKsi)
	mobileIdentity5GS := nasReq.MobileId5GS
	if mobileIdentity5GS == nil || mobileIdentity5GS.TypeOfId != ie.IdType_5GS_SUCI {
		msg.SendIdentityRequest(gnb, ue)
		return nil
	}

	return SetMobileIdentity(amf, ue, mobileIdentity5GS, gnb)
}

func IdentityResponse(nasReq *nas.IdRsp, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) (err error) {

	if nasReq.MobileId == nil {
		return errors.New("[5GC][NAS] Missing mobile identity")
	}
	mobileIdentity5GS := nasReq.MobileId

	err = SetMobileIdentity(amf, ue, mobileIdentity5GS, gnb)

	return err
}

func SetMobileIdentity(amf *context.AMFContext, ue *context.UEContext, mobileIdentity *ie.MobileId5GS, gnb *context.GNBContext) (err error) {
	mobileId := mobileIdentity.SUCIStr()
	if mobileIdentity.TypeOfId != ie.IdType_5GS_SUCI {
		return errors.New("[5GC][NAS] Mobile identity must be SUCI")
	}
	suci := strings.Split(mobileId, "-")
	prov, err := amf.FindProvisionedData(suci[len(suci)-1])
	if err != nil {
		return err
	}
	sCtx := prov.GetSecurityContext()
	sCtx.SetSuci(mobileId)
	sCtx.SetSupi("imsi-" + suci[2] + suci[3] + suci[len(suci)-1])
	ue.SetSecurityContext(&sCtx)

	msg.SendAuthenticationRequest(gnb, ue)

	return nil
}
