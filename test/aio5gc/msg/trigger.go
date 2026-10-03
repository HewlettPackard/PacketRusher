/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package msg

import (
	"my5G-RANTester/test/aio5gc/context"
	nasBuilder "my5G-RANTester/test/aio5gc/msg/nas/builder"
	ngapBuilder "my5G-RANTester/test/aio5gc/msg/ngap/builder"

	"github.com/free5gc/ngap/aper"

	log "github.com/sirupsen/logrus"
)

func SendNGSetupResponse(gnb *context.GNBContext, amf *context.AMFContext) error {
	msg, err := ngapBuilder.NGSetupResponse(amf)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send NG Setup Response")
	return gnb.SendMsg(msg)
}

func SendAuthenticationRequest(gnb *context.GNBContext, ue *context.UEContext) error {
	log.Info("[5GC][NAS] Creating Authentication Request")
	nasRes, err := nasBuilder.AuthenticationRequest(ue)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Authentication Request")
	return gnb.SendMsg(msg)
}

func SendIdentityRequest(gnb *context.GNBContext, ue *context.UEContext) error {
	log.Info("[5GC][NAS] Creating Identity Request")
	nasRes, err := nasBuilder.IdentityRequest(ue)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Identity Request")
	return gnb.SendMsg(msg)
}

func SendSecurityModeCommand(gnb *context.GNBContext, ue *context.UEContext) error {

	log.Info("[5GC][NAS] Creating Security Mode Command")
	nasRes, err := nasBuilder.SecurityModeCommand(ue)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Security Mode Command")
	return gnb.SendMsg(msg)
}

func SendRegistrationAccept(gnb *context.GNBContext, ue *context.UEContext, amf *context.AMFContext) error {

	log.Info("[5GC][NAS] Creating Registration Accept")
	nasRes, err := nasBuilder.RegistrationAccept(ue)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.InitialContextSetupRequest(nasRes, ue, amf)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send Initial Context Setup Request - Registration Accept")
	return gnb.SendMsg(msg)
}

func SendPDUSessionEstablishmentAccept(gnb *context.GNBContext, ue *context.UEContext, smContext *context.SmContext, session *context.SessionContext) error {

	log.Info("[5GC][NAS] Creating PDU Session Establishment Accept")
	nasRes, err := nasBuilder.PDUSessionEstablishmentAccept(ue, smContext)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.PDUSessionResourceSetup(nasRes, *smContext, ue, session)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send PDU Session Establishment Accept")
	return gnb.SendMsg(msg)
}

func SendConfigurationUpdateCommand(gnb *context.GNBContext, ue *context.UEContext, nwName *context.NetworkName) error {

	log.Info("[5GC][NAS] Configuration Update Command")
	nasRes, err := nasBuilder.ConfigurationUpdateCommand(ue, nwName)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Configuration Update Command")
	return gnb.SendMsg(msg)
}

func SendPDUSessionReleaseCommand(gnb *context.GNBContext, ue *context.UEContext, smContext context.SmContext, cause uint8) error {
	log.Info("[5GC][NAS] PDU Session Release Command")
	nasRes, err := nasBuilder.PDUSessionReleaseCommand(ue, smContext, cause)
	if err != nil {
		return err
	}

	msg, err := ngapBuilder.PDUSessionResourceRelease(nasRes, ue, smContext.GetPduSessionId())
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] Send PDU Session Ressource Release - PDU Session Release Command")
	return gnb.SendMsg(msg)
}

func SendUEContextReleaseCommand(gnb *context.GNBContext, ue *context.UEContext, causePresent int, cause aper.Enumerated) error {
	msg, err := ngapBuilder.UEContextReleaseCommand(ue, causePresent, cause)
	if err != nil {
		return err
	}

	log.Info("[5GC][NGAP] UE Context Release Command")
	return gnb.SendMsg(msg)
}
