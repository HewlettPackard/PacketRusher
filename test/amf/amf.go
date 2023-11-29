/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package amf

import (
	"errors"
	"my5G-RANTester/config"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"
	"my5G-RANTester/test/amf/lib/msgHandler/nasMsgHandler"
	"my5G-RANTester/test/amf/lib/msgHandler/ngapMsgHandler"
	"my5G-RANTester/test/amf/lib/tools"
	"strconv"

	"github.com/free5gc/nas"

	"github.com/free5gc/openapi/models"

	log "github.com/sirupsen/logrus"
)

type Amf struct {
	context        context.AMFContext
	ngapReqHandler func(*ngapType.NGAPPDU, context.GNBContext) ([]byte, error)
	nasReqHandler  func(*ngapType.NASPDU, *context.UEContext) (uint8, error)
	conf           config.Config
}

func (a *Amf) GetGnb(amfAddr string) (context.GNBContext, error) {
	return a.context.GetGnb(amfAddr)
}

func (a *Amf) GetNasHandler() func(*ngapType.NASPDU, *context.UEContext) (uint8, error) {
	return a.nasReqHandler
}

func (a *Amf) GetContext() context.AMFContext {
	return a.context
}

func (a *Amf) Init(conf config.Config, handleRequest func(*ngapType.NGAPPDU, context.GNBContext) ([]byte, error), nasReqHandler func(*ngapType.NASPDU, *context.UEContext) (uint8, error)) error {
	amfId := "196673"                    // TODO generate ID
	amfName := "amf.5gc.3gppnetwork.org" // TODO generate Name
	plmn := models.PlmnId{
		Mcc: conf.GNodeB.PlmnList.Mcc,
		Mnc: conf.GNodeB.PlmnList.Mnc,
	}
	i, err := strconv.ParseInt(conf.GNodeB.SliceSupportList.Sst, 10, 32)
	if err != nil {
		err = errors.New("failed to convert config sst to int: " + err.Error())
		return err
	}
	sst := int32(i)
	supportedPlmns := []models.PlmnSnssai{
		{
			PlmnId: &plmn,
			SNssaiList: []models.Snssai{
				{
					Sst: sst,
					Sd:  conf.GNodeB.SliceSupportList.Sd,
				},
			},
		}}
	servedGuami := []models.Guami{
		{
			PlmnId: &plmn,
			AmfId:  amfId,
		},
	}

	a.conf = conf
	a.context = context.AMFContext{}
	a.context.NewAmfContext(
		amfName,
		amfId,
		supportedPlmns,
		servedGuami,
		100,
	)
	a.ngapReqHandler = handleRequest
	a.nasReqHandler = nasReqHandler
	go a.runServer(conf.AMF.Ip, conf.AMF.Port)
	return nil
}

func (a *Amf) NgapDefaultHandler(ngapMsg *ngapType.NGAPPDU, gnb context.GNBContext) (msg []byte, err error) {
	switch ngapMsg.Present {
	case ngapType.NGAPPDUPresentInitiatingMessage:

		switch ngapMsg.InitiatingMessage.ProcedureCode.Value {

		case ngapType.ProcedureCodeNGSetup:
			log.Info("[AMF][NGAP] Received NG Setup Request")
			msg, err = ngapMsgHandler.NGSetupRequest(ngapMsg.InitiatingMessage.Value.NGSetupRequest, a.context, gnb)

		case ngapType.ProcedureCodeInitialUEMessage:
			log.Info("[AMF][NGAP] Received Initial UE Message Request")
			msg, err = ngapMsgHandler.InitialUEMessage(ngapMsg.InitiatingMessage.Value.InitialUEMessage, &a.context, gnb, a.nasReqHandler)

		case ngapType.ProcedureCodeUplinkNASTransport:
			log.Info("[AMF][NGAP] Received uplink NAS Transport")
			msg, err = ngapMsgHandler.UplinkNASTransport(ngapMsg.InitiatingMessage.Value.UplinkNASTransport, a.context, gnb, a.nasReqHandler)

		default:
			err = errors.New("[AMF][NGAP] Received unknown NGAP NGAPPDUPresentInitiatingMessage ProcedureCode")
		}

	case ngapType.NGAPPDUPresentSuccessfulOutcome:

		switch ngapMsg.SuccessfulOutcome.ProcedureCode.Value {

		case ngapType.ProcedureCodeInitialContextSetup:
			log.Info("[AMF][NGAP] Received initial context setup response")
			ngapMsgHandler.InitialContextSetupResponse(ngapMsg.SuccessfulOutcome.Value.InitialContextSetupResponse, a.context)

		default:
			err = errors.New("[AMF][NGAP] Received unknown NGAP NGAPPDUPresentSuccessfulOutcome ProcedureCode")
		}

	case ngapType.NGAPPDUPresentUnsuccessfulOutcome:

		switch ngapMsg.UnsuccessfulOutcome.ProcedureCode.Value {

		default:
			err = errors.New("[AMF][NGAP] Received unknown NGAP NGAPPDUPresentUnsuccessfulOutcome ProcedureCode")
		}
	default:
		err = errors.New("[AMF][NGAP] Received unknown NGAP message")
	}
	return msg, err
}

func (a *Amf) NasDefaultHandler(nasPDU *ngapType.NASPDU, ueContext *context.UEContext) (resNasType uint8, err error) {
	resNasType = 0
	payload := nasPDU.Value
	m := new(nas.Message)
	m.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & 0x0f
	var msg *nas.Message

	if m.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {
		var integrityProtected bool
		msg, integrityProtected, err = tools.Decode(ueContext, payload, false)
		if !integrityProtected {
			return resNasType, errors.New("[AMF][NAS] message integrity could not be verified")
		}

	} else {
		msg, err = tools.DecodePlainNasNoIntegrityCheck(payload)

	}

	amfContext := a.GetContext()
	switch msg.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		log.Info("[AMF][NAS] Received Registration Request")
		err = nasMsgHandler.RegistrationRequest(msg, &amfContext, ueContext)
		resNasType = nas.MsgTypeAuthenticationRequest

	case nas.MsgTypeAuthenticationResponse:
		log.Info("[AMF][NAS] Received Authentication Response")
		err = nasMsgHandler.AuthenticationResponse(msg, ueContext)
		resNasType = nas.MsgTypeSecurityModeCommand

	case nas.MsgTypeSecurityModeComplete:
		log.Info("[AMF][NAS] Received Security Mode Complete")
		err = nasMsgHandler.SecurityModeComplete(msg, &amfContext, ueContext)
		resNasType = nas.MsgTypeRegistrationAccept

	case nas.MsgTypeRegistrationComplete:
		log.Info("[AMF][NAS] Received Registration Complete")
		//TODO: resNasType = nas.MsgTypeConfigurationUpdateCommand

	case nas.MsgTypeULNASTransport:
		err = nasMsgHandler.NASTransport(msg, &amfContext, ueContext)
		log.Info("[AMF][NAS] Received UL NAS Transport")
		resNasType = nas.MsgTypePDUSessionEstablishmentAccept

	default:
		err = errors.New("[AMF][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	if err == nil {
		return resNasType, nil
	} else {
		return resNasType, err
	}
}
