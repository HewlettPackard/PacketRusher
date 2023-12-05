/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package msg

import (
	"my5G-RANTester/test/aio5gc/context"
	nasBuilder "my5G-RANTester/test/aio5gc/msg/nas/builder"
	ngapBuilder "my5G-RANTester/test/aio5gc/msg/ngap/builder"

	log "github.com/sirupsen/logrus"
)

func SendNGSetupResponse(gnb *context.GNBContext, fgc *context.Aio5gc) {

	amf := fgc.GetAMFContext()
	msg, err := ngapBuilder.NGSetupResponse(*amf)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send NG Setup Response")
	gnb.SendMsg(msg)
}

func SendAuthenticationRequest(fgc *context.Aio5gc, ue *context.UEContext) {

	gnb, err := fgc.GetAMFContext().FindGnbById(*ue.GetUserLocationInfo().GlobalGnbId)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("[5GC][NAS] Creating Authentication Request")
	nasRes, err := nasBuilder.AuthenticationRequest(ue)

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Authentication Request")
	gnb.SendMsg(msg)
}

func SendSecurityModeCommand(fgc *context.Aio5gc, ue *context.UEContext) {

	gnb, err := fgc.GetAMFContext().FindGnbById(*ue.GetUserLocationInfo().GlobalGnbId)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NAS] Creating Security Mode Command")
	nasRes, err := nasBuilder.SecurityModeCommand(ue)

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Security Mode Command")
	gnb.SendMsg(msg)
}

func SendRegistrationAccept(fgc *context.Aio5gc, ue *context.UEContext) {
	amf := fgc.GetAMFContext()
	gnb, err := amf.FindGnbById(*ue.GetUserLocationInfo().GlobalGnbId)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NAS] Creating Registration Accept")
	nasRes, err := nasBuilder.RegistrationAccept(ue)

	msg, err := ngapBuilder.InitialContextSetupRequest(nasRes, ue, *amf)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send Initial Context Setup Request - Registration Accept")
	gnb.SendMsg(msg)
}

func SendPDUSessionEstablishmentAccept(fgc *context.Aio5gc, ue *context.UEContext, smContext *context.SmContext) {
	amf := fgc.GetAMFContext()
	gnb, err := amf.FindGnbById(*ue.GetUserLocationInfo().GlobalGnbId)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NAS] Creating PDU Session Establishment Accept")
	nasRes, err := nasBuilder.PDUSessionEstablishmentAccept(ue, smContext)

	msg, err := ngapBuilder.PDUSessionResourceSetup(nasRes, *smContext, ue, fgc)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send PDU Session Establishment Accept")
	gnb.SendMsg(msg)
}

func SendConfigurationUpdateCommand(fgc *context.Aio5gc, ue *context.UEContext, nwName *context.NetworkName) {
	amf := fgc.GetAMFContext()
	gnb, err := amf.FindGnbById(*ue.GetUserLocationInfo().GlobalGnbId)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NAS] Configuration Update Command")
	nasRes, err := nasBuilder.ConfigurationUpdateCommand(ue, nwName)
	if err != nil {
		log.Fatal(err.Error())
	}

	msg, err := ngapBuilder.DownlinkNASTransport(nasRes, ue)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("[5GC][NGAP] Send Downlink NAS Transport - Configuration Update Command")
	gnb.SendMsg(msg)
}
