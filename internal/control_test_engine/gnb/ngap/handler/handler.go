package handler

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	_ "github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/lib/aper"
	"my5G-RANTester/lib/ngap/ngapType"
	_ "net"
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
				log.Fatal("[GNB][NGAP] AMF UE NGAP ID is missing")
				// TODO SEND ERROR INDICATION
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:
			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE NGAP ID is missing")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDNASPDU:
			if ies.Value.NASPDU == nil {
				log.Fatal("[GNB][NGAP] NAS PDU is missing")
				// TODO SEND ERROR INDICATION
			}
			messageNas = ies.Value.NASPDU.Value
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)
}

func HandlerInitialContextSetupRequest(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte
	var sst []string
	var sd []string
	var mobilityRestrict = "not informed"
	var maskedImeisv string
	// var securityKey []byte

	valueMessage := message.InitiatingMessage.Value.InitialContextSetupRequest

	for _, ies := range valueMessage.ProtocolIEs.List {

		// TODO MORE FIELDS TO CHECK HERE
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:
			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE NGAP ID is missing")
				// TODO SEND ERROR INDICATION
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:
			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE NGAP ID is missing")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDNASPDU:
			if ies.Value.NASPDU == nil {
				log.Info("[GNB][NGAP] NAS PDU is missing")
				// TODO SEND ERROR INDICATION
			}
			messageNas = ies.Value.NASPDU.Value

		case ngapType.ProtocolIEIDSecurityKey:
			// TODO using for create new security context between GNB and UE.
			if ies.Value.SecurityKey == nil {
				log.Fatal("[GNB][NGAP] Security-Key is missing")
			}
			// securityKey = ies.Value.SecurityKey.Value.Bytes

		case ngapType.ProtocolIEIDGUAMI:
			if ies.Value.GUAMI == nil {
				log.Fatal("[GNB][NGAP] GUAMI is missing")
			}

		case ngapType.ProtocolIEIDAllowedNSSAI:
			if ies.Value.AllowedNSSAI == nil {
				log.Fatal("[GNB][NGAP] Allowed NSSAI is missing")
			}

			valor := len(ies.Value.AllowedNSSAI.List)
			sst = make([]string, valor)
			sd = make([]string, valor)

			// list S-NSSAI(Single – Network Slice Selection Assistance Information).
			for i, items := range ies.Value.AllowedNSSAI.List {

				if items.SNSSAI.SST.Value != nil {
					sst[i] = fmt.Sprintf("%x", items.SNSSAI.SST.Value)
				} else {
					sst[i] = "not informed"
				}

				if items.SNSSAI.SD != nil {
					sd[i] = fmt.Sprintf("%x", items.SNSSAI.SD.Value)
				} else {
					sd[i] = "not informed"
				}
			}

		case ngapType.ProtocolIEIDMobilityRestrictionList:
			// that field is not mandatory.
			if ies.Value.MobilityRestrictionList == nil {
				log.Info("[GNB][NGAP] Mobility Restriction is missing")
				mobilityRestrict = "not informed"
			} else {
				mobilityRestrict = fmt.Sprintf("%x", ies.Value.MobilityRestrictionList.ServingPLMN.Value)
			}

		case ngapType.ProtocolIEIDMaskedIMEISV:
			// that field is not mandatory.
			// TODO using for mapping UE context
			if ies.Value.MaskedIMEISV == nil {
				log.Info("[GNB][NGAP] Masked IMEISV is missing")
				maskedImeisv = "not informed"
			} else {
				maskedImeisv = fmt.Sprintf("%x", ies.Value.MaskedIMEISV.Value.Bytes)
			}

		case ngapType.ProtocolIEIDUESecurityCapabilities:
			// TODO using for create new security context between UE and GNB.
			// TODO algorithms for create new security context between UE and GNB.
			if ies.Value.UESecurityCapabilities == nil {
				log.Fatal("[GNB][NGAP] UE Security Capabilities is missing")
			}
		}

	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)

	// create UE context.
	ue.CreateUeContext(mobilityRestrict, maskedImeisv, sst, sd)

	// show UE context.
	log.Info("[GNB][UE] UE Context was created with successful")
	log.Info("[GNB][UE] UE RAN ID ", ue.GetRanUeId())
	log.Info("[GNB][UE] UE AMF ID ", ue.GetAmfUeId())
	mcc, mnc := ue.GetUeMobility()
	log.Info("[GNB][UE] UE Mobility Restrict --Plmn-- Mcc: ", mcc, " Mnc: ", mnc)
	log.Info("[GNB][UE] UE Masked Imeisv: ", ue.GetUeMaskedImeiSv())
	log.Info("[GNB][UE] Allowed Nssai-- Sst: ", sst, " Sd: ", sd)

	// send NAS message to UE.
	log.Info("[GNB][NAS][UE] Send Registration Accept.")
	sender.SendToUe(ue, messageNas)

	// send Initial Context Setup Response.
	log.Info("[GNB][NGAP][AMF] Send Initial Context Setup Response.")
	trigger.SendInitialContextSetupResponse(ue)
}

func HandlerPduSessionResourceSetupRequest(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var pduSessionId int64
	var ulTeid uint32
	var upfAddress []byte
	var messageNas []byte
	var sst string
	var sd string
	var pduSType uint64
	var qosId int64
	var fiveQi int64
	var priArp int64

	valueMessage := message.InitiatingMessage.Value.PDUSessionResourceSetupRequest

	for _, ies := range valueMessage.ProtocolIEs.List {

		// TODO MORE FIELDS TO CHECK HERE
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:

			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE ID is missing")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:

			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE ID is missing")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq:

			if ies.Value.PDUSessionResourceSetupListSUReq == nil {
				log.Fatal("[GNB][NGAP] PDU SESSION RESOURCE SETUP LIST SU REQ is missing")
			}
			pDUSessionResourceSetupList := ies.Value.PDUSessionResourceSetupListSUReq

			for _, item := range pDUSessionResourceSetupList.List {

				// check PDU Session NAS PDU.
				if item.PDUSessionNASPDU != nil {
					messageNas = item.PDUSessionNASPDU.Value
				} else {
					log.Fatal("[GNB][NGAP] NAS PDU is missing")
				}

				// check pdu session id and nssai information for create a PDU Session.

				// create a PDU session(PDU SESSION ID + NSSAI).
				pduSessionId = item.PDUSessionID.Value

				if item.SNSSAI.SD != nil {
					sd = fmt.Sprintf("%x", item.SNSSAI.SD.Value)
				} else {
					sd = "not informed"
				}

				if item.SNSSAI.SST.Value != nil {
					sst = fmt.Sprintf("%x", item.SNSSAI.SST.Value)
				} else {
					sst = "not informed"
				}

				if item.PDUSessionResourceSetupRequestTransfer != nil {

					pdu := &ngapType.PDUSessionResourceSetupRequestTransfer{}

					err := aper.UnmarshalWithParams(item.PDUSessionResourceSetupRequestTransfer, pdu, "valueExt")
					if err == nil {
						for _, ies := range pdu.ProtocolIEs.List {

							switch ies.Id.Value {

							case ngapType.ProtocolIEIDULNGUUPTNLInformation:
								ulTeid = binary.BigEndian.Uint32(ies.Value.ULNGUUPTNLInformation.GTPTunnel.GTPTEID.Value)
								upfAddress = ies.Value.ULNGUUPTNLInformation.GTPTunnel.TransportLayerAddress.Value.Bytes

							case ngapType.ProtocolIEIDQosFlowSetupRequestList:
								for _, itemsQos := range ies.Value.QosFlowSetupRequestList.List {
									qosId = itemsQos.QosFlowIdentifier.Value
									fiveQi = itemsQos.QosFlowLevelQosParameters.QosCharacteristics.NonDynamic5QI.FiveQI.Value
									priArp = itemsQos.QosFlowLevelQosParameters.AllocationAndRetentionPriority.PriorityLevelARP.Value
								}

							case ngapType.ProtocolIEIDPDUSessionAggregateMaximumBitRate:

							case ngapType.ProtocolIEIDPDUSessionType:
								pduSType = uint64(ies.Value.PDUSessionType.Value)

							case ngapType.ProtocolIEIDSecurityIndication:

							}
						}
					} else {
						log.Info("[GNB][NGAP] Error in decode Pdu Session Resource Setup Request Transfer")
					}
				} else {
					log.Fatal("[GNB][NGAP] Error in Pdu Session Resource Setup Request, Pdu Session Resource Setup Request Transfer is missing")
				}

			}
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)

	// create PDU Session for GNB UE.
	pduSession, err := ue.CreatePduSession(pduSessionId, sst, sd, pduSType, qosId, priArp, fiveQi, ulTeid, gnb.GetUeTeid(ue))
	if err != nil {
		log.Error("[GNB][NGAP] Error in Pdu Session Resource Setup Request.")
		log.Error("[GNB][NGAP] ", err)

	}
	log.Info("[GNB][NGAP][UE] PDU Session was created with successful.")
	log.Info("[GNB][NGAP][UE] PDU Session Id: ", pduSession.GetPduSessionId())
	sst, sd = ue.GetSelectedNssai(pduSession.GetPduSessionId())
	log.Info("[GNB][NGAP][UE] NSSAI Selected --- sst: ", sst, " sd: ", sd)
	log.Info("[GNB][NGAP][UE] PDU Session Type: ", pduSession.GetPduType())
	log.Info("[GNB][NGAP][UE] QOS Flow Identifier: ", pduSession.GetQosId())
	log.Info("[GNB][NGAP][UE] Uplink Teid: ", pduSession.GetTeidUplink())
	log.Info("[GNB][NGAP][UE] Downlink Teid: ", pduSession.GetTeidDownlink())
	log.Info("[GNB][NGAP][UE] Non-Dynamic-5QI: ", pduSession.GetFiveQI())
	log.Info("[GNB][NGAP][UE] Priority Level ARP: ", pduSession.GetPriorityARP())
	log.Info("[GNB][NGAP][UE] UPF Address: ", fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3]), " :2152")


	// get UPF ip.
	if gnb.GetUpfIp() == "" {
		upfIp := fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3])
		gnb.SetUpfIp(upfIp)
	}

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)

	// send PDU Session Resource Setup Response.
	trigger.SendPduSessionResourceSetupResponse(pduSession, ue, gnb)

	time.Sleep(20 * time.Millisecond)

    msg := context.UEMessage{}
    msg.NewUeMessage(pduSessionId, gnb.GetN3GnbIp(), gnb.GetUpfIp(), fmt.Sprint(pduSession.GetTeidUplink()), fmt.Sprint(pduSession.GetTeidDownlink()))

	sender.SendMessageToUe(ue, msg)
}

func HandlerPduSessionReleaseCommand(gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	valueMessage := message.InitiatingMessage.Value.PDUSessionResourceReleaseCommand

	var amfUeId int64
	var ranUeId int64
	var messageNas aper.OctetString
	var pduSessionIds []ngapType.PDUSessionID

	for _, ies := range valueMessage.ProtocolIEs.List {

		// TODO MORE FIELDS TO CHECK HERE
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:

			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE ID is missing")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:

			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE ID is missing")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value

		case ngapType.ProtocolIEIDNASPDU:
			if ies.Value.NASPDU == nil {
				log.Info("[GNB][NGAP] NAS PDU is missing")
				// TODO SEND ERROR INDICATION
			}
			messageNas = ies.Value.NASPDU.Value

		case ngapType.ProtocolIEIDPDUSessionResourceToReleaseListRelCmd:

			if ies.Value.PDUSessionResourceToReleaseListRelCmd == nil {
				log.Fatal("[GNB][NGAP] PDU SESSION RESOURCE SETUP LIST SU REQ is missing")
			}
			pDUSessionRessourceToReleaseListRelCmd := ies.Value.PDUSessionResourceToReleaseListRelCmd

			for _, pDUSessionRessourceToReleaseItemRelCmd := range pDUSessionRessourceToReleaseListRelCmd.List {
				pduSessionIds = append(pduSessionIds, pDUSessionRessourceToReleaseItemRelCmd.PDUSessionID)
			}
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)

	for _, pduSessionId := range pduSessionIds {
		pduSession, err := ue.GetPduSession(pduSessionId.Value)
		if pduSession == nil || err != nil {
			log.Error("[GNB][NGAP] Unable to delete PDU Session ", pduSessionId.Value, " from UE as the PDU Session was not found. Ignoring.")
			continue
		}
		ue.DeletePduSession(pduSessionId.Value)
		log.Info("[GNB][NGAP] Successfully deleted PDU Session ", pduSessionId.Value, " from UE Context")
	}

	trigger.SendPduSessionReleaseResponse(pduSessionIds, ue)

	sender.SendToUe(ue, messageNas)
}

func HandlerNgSetupResponse(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	err := false
	var plmn string

	// check information about AMF and add in AMF context.
	valueMessage := message.SuccessfulOutcome.Value.NGSetupResponse

	for _, ies := range valueMessage.ProtocolIEs.List {

		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFName:
			if ies.Value.AMFName == nil {
				// TODO error indication. This field is mandatory critically reject
				log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,AMF Name is missing")
				log.Info("[GNB][NGAP] AMF is inactive")
				err = true
			} else {
				amfName := ies.Value.AMFName.Value
				amf.SetAmfName(amfName)
			}

		case ngapType.ProtocolIEIDServedGUAMIList:
			if ies.Value.ServedGUAMIList.List == nil {
				// TODO error indication. This field is mandatory critically reject
				log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Serverd Guami list is missing")
				log.Info("[GNB][NGAP] AMF is inactive")
				err = true
			}
			for _, items := range ies.Value.ServedGUAMIList.List {
				if items.GUAMI.AMFRegionID.Value.Bytes == nil {
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFRegionId is missing")
					log.Info("[GNB][NGAP] AMF is inactive")
					err = true
				}
				if items.GUAMI.AMFPointer.Value.Bytes == nil {
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFPointer is missing")
					log.Info("[GNB][NGAP] AMF is inactive")
					err = true
				}
				if items.GUAMI.AMFSetID.Value.Bytes == nil {
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFSetId is missing")
					log.Info("[GNB][NGAP] AMF is inactive")
					err = true
				}
			}

		case ngapType.ProtocolIEIDRelativeAMFCapacity:
			if ies.Value.RelativeAMFCapacity != nil {
				amfCapacity := ies.Value.RelativeAMFCapacity.Value
				amf.SetAmfCapacity(amfCapacity)
			}

		case ngapType.ProtocolIEIDPLMNSupportList:

			if ies.Value.PLMNSupportList == nil {
				log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, PLMN Support list is missing")
				err = true
			}

			for _, items := range ies.Value.PLMNSupportList.List {

				plmn = fmt.Sprintf("%x", items.PLMNIdentity.Value)
				amf.AddedPlmn(plmn)

				if items.SliceSupportList.List == nil {
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, PLMN Support list is inappropriate")
					log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, Slice Support list is missing")
					err = true
				}

				for _, slice := range items.SliceSupportList.List {

					var sd string
					var sst string

					if slice.SNSSAI.SST.Value != nil {
						sst = fmt.Sprintf("%x", slice.SNSSAI.SST.Value)
					} else {
						sst = "was not informed"
					}

					if slice.SNSSAI.SD != nil {
						sd = fmt.Sprintf("%x", slice.SNSSAI.SD.Value)
					} else {
						sd = "was not informed"
					}

					// update amf slice supported
					amf.AddedSlice(sst, sd)
				}
			}
		}

	}

	if err {
		log.Fatal("[GNB][AMF] AMF is inactive")
		amf.SetStateInactive()
	} else {
		amf.SetStateActive()
		log.Info("[GNB][AMF] AMF Name: ", amf.GetAmfName())
		log.Info("[GNB][AMF] State of AMF: Active")
		log.Info("[GNB][AMF] Capacity of AMF: ", amf.GetAmfCapacity())
		for i := 0; i < amf.GetLenPlmns(); i++ {
			mcc, mnc := amf.GetPlmnSupport(i)
			log.Info("[GNB][AMF] PLMNs Identities Supported by AMF -- mcc: ", mcc, " mnc:", mnc)
		}
		for i := 0; i < amf.GetLenSlice(); i++ {
			sst, sd := amf.GetSliceSupport(i)
			log.Info("[GNB][AMF] List of AMF slices Supported by AMF -- sst:", sst, " sd:", sd)
		}
	}

}

func HandlerNgSetupFailure(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	// check information about AMF and add in AMF context.
	valueMessage := message.UnsuccessfulOutcome.Value.NGSetupFailure

	for _, ies := range valueMessage.ProtocolIEs.List {

		switch ies.Id.Value {

		case ngapType.ProtocolIEIDCause:

			switch ies.Value.Cause.Present {

			case ngapType.CausePresentMisc:

				causeMisc := ies.Value.Cause.Misc.Value

				switch causeMisc {

				case ngapType.CauseMiscPresentUnknownPLMN:
					// Cannot find Served TAI in AMF.
					log.Info("[GNB][AMF][NGAP] Unknown PLMN in Supported TAI")

				case ngapType.CauseMiscPresentUnspecified:
					// No supported TA exist in NG-Setup request.
					log.Info("[GNB][AMF][NGAP] Error cause unspecified")

				}

			case ngapType.CausePresentRadioNetwork:
				// TODO treatment error

			case ngapType.CausePresentTransport:
				// TODO treatment error

			case ngapType.CausePresentProtocol:
				// TODO treatment error

			case ngapType.CausePresentNas:
				// TODO treatment error

			}

		case ngapType.ProtocolIEIDTimeToWait:

			switch ies.Value.TimeToWait.Value {

			case ngapType.TimeToWaitPresentV1s:
			case ngapType.TimeToWaitPresentV2s:
			case ngapType.TimeToWaitPresentV5s:
			case ngapType.TimeToWaitPresentV10s:
			case ngapType.TimeToWaitPresentV20s:
			case ngapType.TimeToWaitPresentV60s:

			}

		case ngapType.ProtocolIEIDCriticalityDiagnostics:

			// TODO treatment error

			// ies.Value.CriticalityDiagnostics
			// errors.IECriticality.Value
			// ngapType.CriticalityPresentReject:
			// ngapType.CriticalityPresentIgnore:
			// ngapType.CriticalityPresentNotify:
			// ngapType.TypeOfErrorPresentNotUnderstood:
			// ngapType.TypeOfErrorPresentMissing:

		}
	}

	// redundant but useful for information about code.
	amf.SetStateInactive()

	log.Info("[GNB][NGAP] AMF is inactive")
}

func HandlerUeContextReleaseCommand(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	valueMessage := message.InitiatingMessage.Value.UEContextReleaseCommand

	var cause int
	var ue_id *ngapType.RANUENGAPID

	for _, ies := range valueMessage.ProtocolIEs.List {

		switch ies.Id.Value {

		case ngapType.ProtocolIEIDUENGAPIDs:
			ue_id = &ies.Value.UENGAPIDs.UENGAPIDPair.RANUENGAPID

		case ngapType.ProtocolIEIDCause:

			switch ies.Value.Cause.Present {
			case ngapType.CausePresentNas:
				cause = int(ies.Value.Cause.Nas.Value)
			case ngapType.CausePresentProtocol:
				cause = int(ies.Value.Cause.Protocol.Value)
			default:
				// TODO treatment error
			}
		}
	}

	ue, err := gnb.GetGnbUe(ue_id.Value)
	if err != nil {
		log.Error("[GNB][AMF] AMF is trying to free the context of an unknown UE")
		return
	}
	// Wait for all messages to be sent to UE before deleting context
	// TODO: Better synchro
	time.Sleep(1 * time.Second)
	gnb.DeleteGnBUe(ue)

	// Send UEContextReleaseComplete
	trigger.SendUeContextReleaseComplete(ue)

	log.Info("[GNB][NGAP] Releasing UE Context, cause: ", cause)
}

func HandlerAmfConfiguratonUpdate(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU)  {

	// TODO: Implement update AMF Context from AMFConfigurationUpdate
	_ = message.InitiatingMessage.Value.AMFConfigurationUpdate
	log.Warn("[GNB][AMF] Ignoring AMFConfigurationUpdate but send AMFConfigurationUpdateAcknowledge")
	log.Warn("[GNB][AMF] TODO: Implement update AMF Context from AMFConfigurationUpdate")

	trigger.SendAmfConfigurationUpdateAcknowledge(amf)
}

func HandlerErrorIndication(gnb *context.GNBContext, message *ngapType.NGAPPDU)  {

	valueMessage := message.InitiatingMessage.Value.ErrorIndication

	var amfUeId, ranUeId int64

	for _, ies := range valueMessage.ProtocolIEs.List {
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:

			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE ID is missing")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

		case ngapType.ProtocolIEIDRANUENGAPID:

			if ies.Value.RANUENGAPID == nil {
				log.Fatal("[GNB][NGAP] RAN UE ID is missing")
				// TODO SEND ERROR INDICATION
			}
			ranUeId = ies.Value.RANUENGAPID.Value
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)

	log.Warn("[GNB][AMF] Received an Error Indication for UE with AMF UE ID: ", ue.GetAmfUeId(), ", RAN UE ID: ", ue.GetRanUeId())
}

func getUeFromContext(gnb *context.GNBContext, ranUeId int64, amfUeId int64) *context.GNBUe {
	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] RAN UE NGAP ID is incorrect, found: ", ranUeId)
		// TODO SEND ERROR INDICATION
	}

	// update AMF UE id.
	if ue.GetAmfUeId() == 0 {
		ue.SetAmfUeId(amfUeId)
	} else {
		if ue.GetAmfUeId() != amfUeId {
			log.Fatal("[GNB][NGAP] AMF UE NGAP ID is incorrect, found: ", amfUeId)
		}
	}
	return ue
}