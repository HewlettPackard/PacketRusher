package handler

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	serviceGateway "my5G-RANTester/internal/control_test_engine/gnb/data/service"
	serviceGtp "my5G-RANTester/internal/control_test_engine/gnb/gtp/service"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/lib/aper"
	"my5G-RANTester/lib/ngap/ngapType"
	"time"
)

func HandlerDownlinkNasTransport(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte

	valueMessage := message.InitiatingMessage.Value.DownlinkNASTransport

	for _, ies := range valueMessage.ProtocolIEs.List {
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:
			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE id is nil")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:
			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE id is nil")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDNASPDU:
			if ies.Value.NASPDU == nil {
				log.Fatal("[GNB][NGAP] NAS PDU is nil")
			}
			messageNas = ies.Value.NASPDU.Value
		}
	}

	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] UE is not found in GNB POOL")
		// TODO SEND ERROR INDICATION
	}

	// update AMF UE id.
	if ue.GetAmfUeId() == 0 {
		ue.SetAmfUeId(amfUeId)
	}

	// ue.SetStateOngoing()

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)
}

func HandlerInitialContextSetupRequest(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte

	valueMessage := message.InitiatingMessage.Value.InitialContextSetupRequest

	for _, ies := range valueMessage.ProtocolIEs.List {

		// TODO MORE FIELDS TO CHECK HERE
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:
			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE id is nil")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:
			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE id is nil")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDNASPDU:
			if ies.Value.NASPDU == nil {
				log.Fatal("[GNB][NGAP] NAS PDU is nil")
			}
			messageNas = ies.Value.NASPDU.Value
		}
	}

	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] UE is not found in GNB POOL")
		// TODO SEND ERROR INDICATION
	}

	// check if AMF UE id.
	if ue.GetAmfUeId() != amfUeId {
		log.Fatal("[GNB][NGAP] Problem in receive AMF UE ID from CORE")
		// TODO SEND ERROR INDICATION
	}

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)

	// send Initial Context Setup Response.
	trigger.SendInitialContextSetupResponse(ue)

}

func HandlerPduSessionResourceSetupRequest(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var pduSessionId int64
	var ulTeid uint32
	var upfAddress []byte
	var messageNas []byte

	valueMessage := message.InitiatingMessage.Value.PDUSessionResourceSetupRequest

	for _, ies := range valueMessage.ProtocolIEs.List {

		// TODO MORE FIELDS TO CHECK HERE
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:

			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE id is nil")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:

			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE id is nil")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq:

			if ies.Value.PDUSessionResourceSetupListSUReq == nil {
				log.Fatal("[GNB][NGAP] PDUSessionResourceSetupListSUReq is nil")
			}
			pDUSessionResourceSetupList := ies.Value.PDUSessionResourceSetupListSUReq
			for _, item := range pDUSessionResourceSetupList.List {
				if item.PDUSessionNASPDU != nil {
					messageNas = item.PDUSessionNASPDU.Value
				} else {
					log.Fatal("[GNB][NGAP] NAS PDU is nil")
				}
				pduSessionId = item.PDUSessionID.Value
				if item.PDUSessionResourceSetupRequestTransfer != nil {
					// pduSessionResourceSetupRequestTransfer = item.PDUSessionResourceSetupRequestTransfer
					pdu := &ngapType.PDUSessionResourceSetupRequestTransfer{}
					err := aper.UnmarshalWithParams(item.PDUSessionResourceSetupRequestTransfer, pdu, "valueExt")
					if err == nil {
						for _, ies := range pdu.ProtocolIEs.List {
							if ies.Id.Value == ngapType.ProtocolIEIDULNGUUPTNLInformation {
								ulTeid = binary.BigEndian.Uint32(ies.Value.ULNGUUPTNLInformation.GTPTunnel.GTPTEID.Value)
								upfAddress = ies.Value.ULNGUUPTNLInformation.GTPTunnel.TransportLayerAddress.Value.Bytes
							}
						}
					} else {
						log.Info("[GNB][NGAP] Error in decode Pdu Session Resource Setup Request Transfer")
					}
				} else {
					log.Info("[GNB][NGAP] Pdu Session Resource Setup Request Transfer is nil")
				}

			}
		}
	}

	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] UE is not found in GNB POOL")
		// TODO SEND ERROR INDICATION
	}

	// check if AMF UE id.
	if ue.GetAmfUeId() != amfUeId {
		log.Fatal("[GNB][NGAP] Problem in receive AMF UE ID from CORE")
		// TODO SEND ERROR INDICATION
	}

	// set PDU Session ID for GNB UE Context.
	ue.SetPduSessionId(pduSessionId)

	// set uplink teid.
	// downlink is solved when make a GNBue.
	ue.SetTeidUplink(ulTeid)

	// get UPF ip.
	if gnb.GetUpfIp() == "" {
		upfIp := fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3])
		gnb.SetUpfIp(upfIp)
	}

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)

	// send PDU Session Resource Setup Response.
	trigger.SendPduSessionResourceSetupResponse(ue, gnb)

	// configure GTP tunnel and listen.
	if gnb.GetN3Plane() == nil {
		// TODO check if GTP tunnel and gateway is ok.
		serviceGtp.InitGTPTunnel(gnb)
		serviceGateway.InitGatewayGnb(gnb)
	}

	time.Sleep(20 * time.Millisecond)

	// ue is ready for data plane.
	// send GNB UE IP message to UE.
	UeGnBIp := ue.GetIp()
	sender.SendToUe(ue, UeGnBIp)

}

func HandlerNgSetupResponse(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	// check information about AMF and add in AMF context.
	amf.SetStateActive()

}

func HandlerNgSetupFailure(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	// check information about AMF and add in AMF context.
	// redundant
	amf.SetStateInactive()

}
