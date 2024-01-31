/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"
	"strings"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"

	"github.com/free5gc/nas"
	log "github.com/sirupsen/logrus"
)

func RegistrationRequest(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.Deregistered:
		err = DefaultRegistrationRequest(nasReq, amf, ue, gnb)

	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for RegistrationRequest", ue.GetState().Current())
	}
	return err
}

func DefaultRegistrationRequest(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) error {

	err := ue.GetUeFsm().SendEvent(ue.GetState(), context.RegistrationRequest, fsm.ArgsType{"ue": ue}, log.NewEntry(log.StandardLogger()))
	if err != nil {
		return err
	}
	regType := nasReq.RegistrationRequest.NgksiAndRegistrationType5GS.GetRegistrationType5GS()
	if regType != nasMessage.RegistrationType5GSInitialRegistration {
		return errors.New("[5GC][NAS] Received unsupported registration type")
	}

	// NgKsi
	ngKsi := models.NgKsi{}
	switch nasReq.NgksiAndRegistrationType5GS.GetTSC() {
	case nasMessage.TypeOfSecurityContextFlagNative:
		ngKsi.Tsc = models.ScType_NATIVE
	default:
		return errors.New("[5GC] Unsupported KSI sc type")
	}
	ngKsi.Ksi = int32(nasReq.NgksiAndRegistrationType5GS.GetNasKeySetIdentifiler())
	if ngKsi.Tsc == models.ScType_NATIVE && ngKsi.Ksi != 7 {
	} else {
		ngKsi.Tsc = models.ScType_NATIVE
		ngKsi.Ksi = 0
	}

	gmm := nasReq.GmmMessage
	ue.SetSecurityCapability(gmm.RegistrationRequest.UESecurityCapability)
	ue.SetNgKsi(ngKsi)

	mobileIdentity5GS := gmm.RegistrationRequest.MobileIdentity5GS
	if mobileIdentity5GS.Len <= 1 {
		// RegistrationRequest is missing MobileIdentity, we send an Identity Request
		msg.SendIdentityRequest(gnb, ue)
		return nil
	}

	_, mobileIdType, err := mobileIdentity5GS.GetMobileIdentity()
	if mobileIdType != "SUCI" {
		log.Warn("[5GC][NAS] UE id uses IDType " + mobileIdType + " but is not yet supported by aio5gc. Try to request SUCI identity.")
		msg.SendIdentityRequest(gnb, ue)
		return nil
	}
	return SetMobileIdentity(amf, ue, mobileIdentity5GS, gnb)
}

func IdentityResponse(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext, gnb *context.GNBContext) (err error) {

	identityResponse := nasReq.GmmMessage.IdentityResponse
	if identityResponse == nil {
		return errors.New("[5GC][NAS] Received Unexpected Message")
	}

	mobileIdentity := nasReq.GmmMessage.IdentityResponse.MobileIdentity

	mobileIdentity5GS := nasType.MobileIdentity5GS{
		Iei:    mobileIdentity.Iei,
		Len:    mobileIdentity.Len,
		Buffer: mobileIdentity.Buffer,
	}

	err = SetMobileIdentity(amf, ue, mobileIdentity5GS, gnb)

	return err
}

func SetMobileIdentity(amf *context.AMFContext, ue *context.UEContext, mobileIdentity nasType.MobileIdentity5GS, gnb *context.GNBContext) (err error) {
	mobileId, mobileIdType, err := mobileIdentity.GetMobileIdentity()
	if mobileIdType != "SUCI" {
		return errors.New("[5GC][NAS] UE id uses IDType " + mobileIdType + " but is not yet supported by aio5gc.")
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
