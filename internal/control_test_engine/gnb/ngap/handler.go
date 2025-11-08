/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2023-2024 Valentin D'Emmanuele
 */
package ngap

import (
	"encoding/binary"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"net/netip"
	"reflect"

	_ "net"

	"github.com/free5gc/aper"

	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
	_ "github.com/vishvananda/netlink"
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
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot send DownlinkNASTransport message to UE with RANUEID %d as it does not know this UE", ranUeId)
		return
	}

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
	var ueSecurityCapabilities *ngapType.UESecurityCapabilities
	var pDUSessionResourceSetupListCxtReq *ngapType.PDUSessionResourceSetupListCxtReq
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
			ueSecurityCapabilities = ies.Value.UESecurityCapabilities

		case ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtReq:
			if ies.Value.PDUSessionResourceSetupListCxtReq == nil {
				log.Fatal("[GNB][NGAP] PDUSessionResourceSetupListCxtReq is missing")
			}
			pDUSessionResourceSetupListCxtReq = ies.Value.PDUSessionResourceSetupListCxtReq
		}

	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Warn("[GNB][NGAP] Cannot setup context for unknown or terminated UE with RANUEID ", ranUeId, ", ignoring request")
		return
	}
	// create UE context.
	ue.CreateUeContext(mobilityRestrict, maskedImeisv, sst, sd, ueSecurityCapabilities)

	// show UE context.
	log.Info("[GNB][UE] UE Context was created with successful")
	log.Info("[GNB][UE] UE RAN ID ", ue.GetRanUeId())
	log.Info("[GNB][UE] UE AMF ID ", ue.GetAmfUeId())
	mcc, mnc := ue.GetUeMobility()
	log.Info("[GNB][UE] UE Mobility Restrict --Plmn-- Mcc: ", mcc, " Mnc: ", mnc)
	log.Info("[GNB][UE] UE Masked Imeisv: ", ue.GetUeMaskedImeiSv())
	log.Info("[GNB][UE] Allowed Nssai-- Sst: ", sst, " Sd: ", sd)

	if messageNas != nil {
		sender.SendToUe(ue, messageNas)
	}

	if pDUSessionResourceSetupListCxtReq != nil {
		log.Info("[GNB][NGAP] AMF is requesting some PDU Session to be setup during Initial Context Setup")
		for _, pDUSessionResourceSetupItemCtxReq := range pDUSessionResourceSetupListCxtReq.List {
			pduSessionId := pDUSessionResourceSetupItemCtxReq.PDUSessionID.Value
			sst := fmt.Sprintf("%x", pDUSessionResourceSetupItemCtxReq.SNSSAI.SST.Value)
			sd := "not informed"
			if pDUSessionResourceSetupItemCtxReq.SNSSAI.SD != nil {
				sd = fmt.Sprintf("%x", pDUSessionResourceSetupItemCtxReq.SNSSAI.SD.Value)
			}

			pDUSessionResourceSetupRequestTransferBytes := pDUSessionResourceSetupItemCtxReq.PDUSessionResourceSetupRequestTransfer
			pDUSessionResourceSetupRequestTransfer := &ngapType.PDUSessionResourceSetupRequestTransfer{}
			err := aper.UnmarshalWithParams(pDUSessionResourceSetupRequestTransferBytes, pDUSessionResourceSetupRequestTransfer, "valueExt")
			if err != nil {
				log.Error("[GNB] Unable to unmarshall PDUSessionResourceSetupRequestTransfer: ", err)
				continue
			}

			var gtpTunnel *ngapType.GTPTunnel
			var upfIp string
			var teidUplink aper.OctetString
			for _, ie := range pDUSessionResourceSetupRequestTransfer.ProtocolIEs.List {
				switch ie.Id.Value {

				case ngapType.ProtocolIEIDULNGUUPTNLInformation:
					uLNGUUPTNLInformation := ie.Value.ULNGUUPTNLInformation

					gtpTunnel = uLNGUUPTNLInformation.GTPTunnel
					upfIp, _ = ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
					teidUplink = gtpTunnel.GTPTEID.Value
				}
			}

			_, err = ue.CreatePduSession(pduSessionId, upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue))
			if err != nil {
				log.Error("[GNB] ", err)
			}

			if pDUSessionResourceSetupItemCtxReq.NASPDU != nil {
				sender.SendToUe(ue, pDUSessionResourceSetupItemCtxReq.NASPDU.Value)
			}
		}

		msg := context.UEMessage{GNBPduSessions: ue.GetPduSessions(), GnbIp: gnb.GetN3GnbIp()}
		sender.SendMessageToUe(ue, msg)
	}

	// send Initial Context Setup Response.
	log.Info("[GNB][NGAP][AMF] Send Initial Context Setup Response.")
	trigger.SendInitialContextSetupResponse(ue, gnb)
}

func HandlerPduSessionResourceSetupRequest(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	var ranUeId int64
	var amfUeId int64
	var pDUSessionResourceSetupList *ngapType.PDUSessionResourceSetupListSUReq

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
			pDUSessionResourceSetupList = ies.Value.PDUSessionResourceSetupListSUReq
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Warn("[GNB][NGAP] Cannot setup PDU Session for unknown or terminated UE with RANUEID ", ranUeId, ", ignoring request")
		return
	}

	var configuredPduSessions []*context.GnbPDUSession
	for _, item := range pDUSessionResourceSetupList.List {
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

		upfIp := fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3])

		// create PDU Session for GNB UE.
		pduSession, err := ue.CreatePduSession(pduSessionId, upfIp, sst, sd, pduSType, qosId, priArp, fiveQi, ulTeid, gnb.GetUeTeid(ue))
		if err != nil {
			log.Error("[GNB][NGAP] Error in Pdu Session Resource Setup Request.")
			log.Error("[GNB][NGAP] ", err)

			// If PDU session already exists, retrieve it instead of failing
			existingSession, getErr := ue.GetPduSession(pduSessionId)
			if getErr == nil && existingSession != nil {
				log.Warn("[GNB][NGAP] PDU Session ", pduSessionId, " already exists, using existing session")
				pduSession = existingSession
			} else {
				// Cannot create or retrieve session, skip this item
				log.Error("[GNB][NGAP] Cannot create or retrieve PDU Session ", pduSessionId, ", skipping")
				continue
			}
		}
		configuredPduSessions = append(configuredPduSessions, pduSession)

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

		// send NAS message to UE.
		sender.SendToUe(ue, messageNas)

		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession
		msg := context.UEMessage{GnbIp: gnb.GetN3GnbIp(), GNBPduSessions: pduSessions}

		sender.SendMessageToUe(ue, msg)
	}

	// send PDU Session Resource Setup Response.
	trigger.SendPduSessionResourceSetupResponse(configuredPduSessions, ue, gnb)
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
	if ue == nil {
		log.Warn("[GNB][NGAP] Cannot release PDU Session for unknown or terminated UE with RANUEID ", ranUeId, ", ignoring request")
		return
	}

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
			log.Error("[GNB][NGAP] Received failure from AMF: ", causeToString(ies.Value.Cause))

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

	var cause *ngapType.Cause
	var ue_id *ngapType.RANUENGAPID

	for _, ies := range valueMessage.ProtocolIEs.List {

		switch ies.Id.Value {

		case ngapType.ProtocolIEIDUENGAPIDs:
			ue_id = &ies.Value.UENGAPIDs.UENGAPIDPair.RANUENGAPID

		case ngapType.ProtocolIEIDCause:
			cause = ies.Value.Cause
		}
	}

	if ue_id == nil {
		log.Warn("[GNB][NGAP] UE Context Release Command missing UE ID")
		return
	}

	ue, err := gnb.GetGnbUe(ue_id.Value)
	if err != nil || ue == nil {
		log.Warn("[GNB][NGAP] AMF is trying to free the context of an unknown UE with RANUEID ", ue_id.Value, ", ignoring")
		return
	}

	log.Info("[GNB][NGAP] Releasing UE Context for UE ", ue_id.Value, ", cause: ", causeToString(cause))

	// Send UEContextReleaseComplete before deleting context
	trigger.SendUeContextReleaseComplete(ue)

	// Clean up UE context
	gnb.DeleteGnBUe(ue)
}

func HandlerAmfConfigurationUpdate(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	log.Debugf("Before Update:")
	for oldAmf := range gnb.IterGnbAmf() {
		tnla := oldAmf.GetTNLA()
		log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
			oldAmf.GetAmfName(), oldAmf.GetAmfIpPort().Addr(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
	}

	var amfName string
	var amfCapacity int64
	var amfRegionId, amfSetId, amfPointer aper.BitString

	valueMessage := message.InitiatingMessage.Value.AMFConfigurationUpdate
	for _, ie := range valueMessage.ProtocolIEs.List {
		switch ie.Id.Value {

		case ngapType.ProtocolIEIDAMFName:
			amfName = ie.Value.AMFName.Value

		case ngapType.ProtocolIEIDServedGUAMIList:
			for _, servedGuamiItem := range ie.Value.ServedGUAMIList.List {
				amfRegionId = servedGuamiItem.GUAMI.AMFRegionID.Value
				amfSetId = servedGuamiItem.GUAMI.AMFSetID.Value
				amfPointer = servedGuamiItem.GUAMI.AMFPointer.Value
			}
		case ngapType.ProtocolIEIDRelativeAMFCapacity:
			amfCapacity = ie.Value.RelativeAMFCapacity.Value

		case ngapType.ProtocolIEIDAMFTNLAssociationToAddList:
			toAddList := ie.Value.AMFTNLAssociationToAddList
			for _, toAddItem := range toAddList.List {
				ipv4String, _ := ngapConvert.IPAddressToString(*toAddItem.AMFTNLAssociationAddress.EndpointIPAddress)
				if ipv4String == "" {
					// ignore AMF that does not have IPv4 address
					continue
				}
				ipv4Port := netip.AddrPortFrom(netip.MustParseAddr(ipv4String), 38412) // with default sctp port

				if oldAmf := gnb.FindGnbAmfByIpPort(ipv4Port); oldAmf != nil {
					log.Info("[GNB] SCTP/NGAP service exists")
					continue
				}

				newAmf := gnb.NewGnBAmf(ipv4Port)
				newAmf.SetAmfName(amfName)
				newAmf.SetAmfCapacity(amfCapacity)
				newAmf.SetRegionId(amfRegionId)
				newAmf.SetSetId(amfSetId)
				newAmf.SetPointer(amfPointer)
				newAmf.SetTNLAUsage(toAddItem.TNLAssociationUsage.Value)
				newAmf.SetTNLAWeight(toAddItem.TNLAddressWeightFactor.Value)

				// start communication with AMF(SCTP).
				if err := InitConn(newAmf, gnb); err != nil {
					log.Fatal("Error in", err)
				} else {
					log.Info("[GNB] SCTP/NGAP service is running")
					// wg.Add(1)
				}

				trigger.SendNgSetupRequest(gnb, newAmf)

			}

		case ngapType.ProtocolIEIDAMFTNLAssociationToRemoveList:
			toRemoveList := ie.Value.AMFTNLAssociationToRemoveList
			for _, toRemoveItem := range toRemoveList.List {
				ipv4String, _ := ngapConvert.IPAddressToString(*toRemoveItem.AMFTNLAssociationAddress.EndpointIPAddress)
				if ipv4String == "" {
					// ignore AMF that does not have IPv4 address
					continue
				}
				ipv4Port := netip.AddrPortFrom(netip.MustParseAddr(ipv4String), 38412) // with default sctp port

				oldAmf := gnb.FindGnbAmfByIpPort(ipv4Port)
				if oldAmf == nil {
					continue
				}

				log.Info("[GNB][AMF] Remove AMF:", amf.GetAmfName(), " IP:", amf.GetAmfIpPort().Addr())
				tnla := amf.GetTNLA()
				tnla.Release() // Close SCTP Conntection
				gnb.DeleteGnBAmf(oldAmf.GetAmfId())
			}

		case ngapType.ProtocolIEIDAMFTNLAssociationToUpdateList:
			toUpdateList := ie.Value.AMFTNLAssociationToUpdateList
			for _, toUpdateItem := range toUpdateList.List {
				ipv4String, _ := ngapConvert.IPAddressToString(*toUpdateItem.AMFTNLAssociationAddress.EndpointIPAddress)
				if ipv4String == "" {
					// ignore AMF that does not have IPv4 address
					continue
				}
				ipv4Port := netip.AddrPortFrom(netip.MustParseAddr(ipv4String), 38412) // with default sctp port

				oldAmf := gnb.FindGnbAmfByIpPort(ipv4Port)
				if oldAmf == nil {
					continue
				}

				oldAmf.SetAmfName(amfName)
				oldAmf.SetAmfCapacity(amfCapacity)
				oldAmf.SetRegionId(amfRegionId)
				oldAmf.SetSetId(amfSetId)
				oldAmf.SetPointer(amfPointer)

				oldAmf.SetTNLAUsage(toUpdateItem.TNLAssociationUsage.Value)
				oldAmf.SetTNLAWeight(toUpdateItem.TNLAddressWeightFactor.Value)
			}

			// default:
		}
	}

	log.Debugf("After Update:")
	for oldAmf := range gnb.IterGnbAmf() {
		tnla := oldAmf.GetTNLA()
		log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
			oldAmf.GetAmfName(), oldAmf.GetAmfIpPort().Addr(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
	}

	trigger.SendAmfConfigurationUpdateAcknowledge(amf)
}

func HandlerAmfStatusIndication(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	valueMessage := message.InitiatingMessage.Value.AMFStatusIndication
	for _, ie := range valueMessage.ProtocolIEs.List {
		switch ie.Id.Value {
		case ngapType.ProtocolIEIDUnavailableGUAMIList:
			for _, unavailableGuamiItem := range ie.Value.UnavailableGUAMIList.List {
				octetStr := unavailableGuamiItem.GUAMI.PLMNIdentity.Value
				hexStr := fmt.Sprintf("%02x%02x%02x", octetStr[0], octetStr[1], octetStr[2])
				var unavailableMcc, unavailableMnc string
				unavailableMcc = string(hexStr[1]) + string(hexStr[0]) + string(hexStr[3])
				unavailableMnc = string(hexStr[5]) + string(hexStr[4])
				if hexStr[2] != 'f' {
					unavailableMnc = string(hexStr[2]) + string(hexStr[5]) + string(hexStr[4])
				}

				// select backup AMF
				var backupAmf *context.GNBAmf
				for oldAmf := range gnb.IterGnbAmf() {
					if unavailableGuamiItem.BackupAMFName != nil &&
						oldAmf.GetAmfName() == unavailableGuamiItem.BackupAMFName.Value {
						backupAmf = oldAmf
						break
					}
				}
				if backupAmf == nil {
					return
				}

				for oldAmf := range gnb.IterGnbAmf() {
					for j := 0; j < oldAmf.GetLenPlmns(); j++ {
						oldAmfSupportMcc, oldAmfSupportMnc := oldAmf.GetPlmnSupport(j)

						if oldAmfSupportMcc == unavailableMcc && oldAmfSupportMnc == unavailableMnc &&
							reflect.DeepEqual(oldAmf.GetRegionId(), unavailableGuamiItem.GUAMI.AMFRegionID.Value) &&
							reflect.DeepEqual(oldAmf.GetSetId(), unavailableGuamiItem.GUAMI.AMFSetID.Value) &&
							reflect.DeepEqual(oldAmf.GetPointer(), unavailableGuamiItem.GUAMI.AMFPointer.Value) {

							log.Info("[GNB][AMF] Remove AMF: [",
								"Id: ", oldAmf.GetAmfId(),
								"Name: ", oldAmf.GetAmfName(),
								"Ipv4: ", oldAmf.GetAmfIpPort().Addr(),
								"]",
							)

							tnla := oldAmf.GetTNLA()

							// NGAP UE-TNLA Rebinding
							uePool := gnb.GetUePool()
							uePool.Range(func(k, v any) bool {
								ue, ok := v.(*context.GNBUe)
								if !ok {
									return true
								}

								if ue.GetAmfId() == oldAmf.GetAmfId() {
									// set amfId and SCTP association for UE.
									ue.SetAmfId(backupAmf.GetAmfId())
									ue.SetSCTP(backupAmf.GetSCTPConn())
								}

								return true
							})

							prUePool := gnb.GetPrUePool()
							prUePool.Range(func(k, v any) bool {
								ue, ok := v.(*context.GNBUe)
								if !ok {
									return true
								}

								if ue.GetAmfId() == oldAmf.GetAmfId() {
									// set amfId and SCTP association for UE.
									ue.SetAmfId(backupAmf.GetAmfId())
									ue.SetSCTP(backupAmf.GetSCTPConn())
								}

								return true
							})

							tnla.Release()
							gnb.DeleteGnBAmf(oldAmf.GetAmfId())

							break
						}
					}
				}
			}
		}
	}
}

func HandlerPathSwitchRequestAcknowledge(gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	var pduSessionResourceSwitchedList *ngapType.PDUSessionResourceSwitchedList
	valueMessage := message.SuccessfulOutcome.Value.PathSwitchRequestAcknowledge

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

		case ngapType.ProtocolIEIDPDUSessionResourceSwitchedList:
			pduSessionResourceSwitchedList = ies.Value.PDUSessionResourceSwitchedList
			if pduSessionResourceSwitchedList == nil {
				log.Fatal("[GNB][NGAP] PduSessionResourceSwitchedList is missing")
				// TODO SEND ERROR INDICATION
			}
		}

	}
	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot Xn Handover unknown UE With RANUEID %d", ranUeId)
		return
	}

	if pduSessionResourceSwitchedList == nil || len(pduSessionResourceSwitchedList.List) == 0 {
		log.Warn("[GNB] No PDU Sessions to be switched")
		return
	}

	for _, pduSessionResourceSwitchedItem := range pduSessionResourceSwitchedList.List {
		pduSessionId := pduSessionResourceSwitchedItem.PDUSessionID.Value
		pduSession, err := ue.GetPduSession(pduSessionId)
		if err != nil {
			log.Error("[GNB] Trying to path switch an unknown PDU Session ID ", pduSessionId, ": ", err)
			continue
		}

		pathSwitchRequestAcknowledgeTransferBytes := pduSessionResourceSwitchedItem.PathSwitchRequestAcknowledgeTransfer
		pathSwitchRequestAcknowledgeTransfer := &ngapType.PathSwitchRequestAcknowledgeTransfer{}
		err = aper.UnmarshalWithParams(pathSwitchRequestAcknowledgeTransferBytes, pathSwitchRequestAcknowledgeTransfer, "valueExt")
		if err != nil {
			log.Error("[GNB] Unable to unmarshall PathSwitchRequestAcknowledgeTransfer: ", err)
			continue
		}

		if pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation != nil {
			gtpTunnel := pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation.GTPTunnel
			upfIpv4, _ := ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
			teidUplink := gtpTunnel.GTPTEID.Value

			// Set new Teid Uplink received in PathSwitchRequestAcknowledge
			pduSession.SetTeidUplink(binary.BigEndian.Uint32(teidUplink))
			pduSession.SetUpfIp(upfIpv4)
		}
		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession

		msg := context.UEMessage{GNBPduSessions: pduSessions, GnbIp: gnb.GetN3GnbIp()}

		sender.SendMessageToUe(ue, msg)
	}

	log.Info("[GNB] Handover completed successfully for UE ", ue.GetRanUeId())
}

func HandlerHandoverRequest(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	var ueSecurityCapabilities *ngapType.UESecurityCapabilities
	var sst []string
	var sd []string
	var maskedImeisv string
	var sourceToTargetContainer *ngapType.SourceToTargetTransparentContainer
	var pDUSessionResourceSetupListHOReq *ngapType.PDUSessionResourceSetupListHOReq
	var amfUeId int64

	valueMessage := message.InitiatingMessage.Value.HandoverRequest

	for _, ies := range valueMessage.ProtocolIEs.List {
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDAMFUENGAPID:
			if ies.Value.AMFUENGAPID == nil {
				log.Fatal("[GNB][NGAP] AMF UE ID is missing")
			}
			amfUeId = ies.Value.AMFUENGAPID.Value

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

		case ngapType.ProtocolIEIDMaskedIMEISV:
			// that field is not mandatory.
			// TODO using for mapping UE context
			if ies.Value.MaskedIMEISV == nil {
				log.Info("[GNB][NGAP] Masked IMEISV is missing")
				maskedImeisv = "not informed"
			} else {
				maskedImeisv = fmt.Sprintf("%x", ies.Value.MaskedIMEISV.Value.Bytes)
			}

		case ngapType.ProtocolIEIDSourceToTargetTransparentContainer:
			sourceToTargetContainer = ies.Value.SourceToTargetTransparentContainer
			if sourceToTargetContainer == nil {
				log.Fatal("[GNB][NGAP] sourceToTargetContainer is missing")
				// TODO SEND ERROR INDICATION
			}

		case ngapType.ProtocolIEIDPDUSessionResourceSetupListHOReq:
			pDUSessionResourceSetupListHOReq = ies.Value.PDUSessionResourceSetupListHOReq
			if pDUSessionResourceSetupListHOReq == nil {
				log.Fatal("[GNB][NGAP] pDUSessionResourceSetupListHOReq is missing")
				// TODO SEND ERROR INDICATION
			}

		case ngapType.ProtocolIEIDUESecurityCapabilities:
			if ies.Value.UESecurityCapabilities == nil {
				log.Fatal("[GNB][NGAP] UE Security Capabilities is missing")
			}
			ueSecurityCapabilities = ies.Value.UESecurityCapabilities
		}
	}

	if sourceToTargetContainer == nil {
		log.Error("[GNB] HandoverRequest message from AMF is missing mandatory SourceToTargetTransparentContainer")
		return
	}

	sourceToTargetContainerBytes := sourceToTargetContainer.Value
	sourceToTargetContainerNgap := &ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
	err := aper.UnmarshalWithParams(sourceToTargetContainerBytes, sourceToTargetContainerNgap, "valueExt")
	if err != nil {
		log.Error("[GNB] Unable to unmarshall SourceToTargetTransparentContainer: ", err)
		return
	}
	if sourceToTargetContainerNgap.IndexToRFSP == nil {
		log.Error("[GNB] SourceToTargetTransparentContainer from source gNodeB is missing IndexToRFSP")
		return
	}
	prUeId := sourceToTargetContainerNgap.IndexToRFSP.Value

	ue, err := gnb.NewGnBUe(nil, nil, prUeId, nil)
	if ue == nil || err != nil {
		log.Fatalf("[GNB] HandoverFailure: %s", err)
	}
	ue.SetAmfUeId(amfUeId)

	ue.CreateUeContext("not informed", maskedImeisv, sst, sd, ueSecurityCapabilities)

	for _, pDUSessionResourceSetupItemHOReq := range pDUSessionResourceSetupListHOReq.List {
		pduSessionId := pDUSessionResourceSetupItemHOReq.PDUSessionID.Value
		sst := fmt.Sprintf("%x", pDUSessionResourceSetupItemHOReq.SNSSAI.SST.Value)
		sd := "not informed"
		if pDUSessionResourceSetupItemHOReq.SNSSAI.SD != nil {
			sd = fmt.Sprintf("%x", pDUSessionResourceSetupItemHOReq.SNSSAI.SD.Value)
		}

		handOverRequestTransferBytes := pDUSessionResourceSetupItemHOReq.HandoverRequestTransfer
		handOverRequestTransfer := &ngapType.PDUSessionResourceSetupRequestTransfer{}
		err := aper.UnmarshalWithParams(handOverRequestTransferBytes, handOverRequestTransfer, "valueExt")
		if err != nil {
			log.Error("[GNB] Unable to unmarshall HandOverRequestTransfer: ", err)
			continue
		}

		var gtpTunnel *ngapType.GTPTunnel
		var upfIp string
		var teidUplink aper.OctetString
		for _, ie := range handOverRequestTransfer.ProtocolIEs.List {
			switch ie.Id.Value {

			case ngapType.ProtocolIEIDULNGUUPTNLInformation:
				uLNGUUPTNLInformation := ie.Value.ULNGUUPTNLInformation

				gtpTunnel = uLNGUUPTNLInformation.GTPTunnel
				upfIp, _ = ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
				teidUplink = gtpTunnel.GTPTEID.Value
			}
		}

		_, err = ue.CreatePduSession(pduSessionId, upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue))
		if err != nil {
			log.Error("[GNB] ", err)
		}
	}

	trigger.SendHandoverRequestAcknowledge(gnb, ue)
}

func HandlerHandoverCommand(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapType.NGAPPDU) {
	valueMessage := message.SuccessfulOutcome.Value.HandoverCommand

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
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot NGAP  Handover unknown UE With RANUEID %d", ranUeId)
		return
	}
	newGnb := ue.GetHandoverGnodeB()
	if newGnb == nil {
		log.Error("[GNB] AMF is sending a Handover Command for an UE we did not send a Handover Required message")
		// TODO SEND ERROR INDICATION
		return
	}

	newGnbRx := make(chan context.UEMessage, 1)
	newGnbTx := make(chan context.UEMessage, 1)
	newGnb.GetInboundChannel() <- context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, PrUeId: ue.GetPrUeId(), IsHandover: true}

	msg := context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, GNBInboundChannel: newGnb.GetInboundChannel()}

	sender.SendMessageToUe(ue, msg)
}

func HandlerPaging(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

	valueMessage := message.InitiatingMessage.Value.Paging

	var uEPagingIdentity *ngapType.UEPagingIdentity
	var tAIListForPaging *ngapType.TAIListForPaging

	for _, ies := range valueMessage.ProtocolIEs.List {
		switch ies.Id.Value {

		case ngapType.ProtocolIEIDUEPagingIdentity:

			if ies.Value.UEPagingIdentity == nil {
				log.Fatal("[GNB][NGAP] UE Paging Identity is missing")
			}
			uEPagingIdentity = ies.Value.UEPagingIdentity

		case ngapType.ProtocolIEIDTAIListForPaging:

			if ies.Value.TAIListForPaging == nil {
				log.Fatal("[GNB][NGAP] TAI List For Paging is missing")
			}
			tAIListForPaging = ies.Value.TAIListForPaging
		}
	}
	_ = tAIListForPaging

	gnb.AddPagedUE(uEPagingIdentity.FiveGSTMSI)

	log.Info("[GNB][AMF] Paging UE")
}

func HandlerErrorIndication(gnb *context.GNBContext, message *ngapType.NGAPPDU) {

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

	log.Warn("[GNB][AMF] Received an Error Indication for UE with AMF UE ID: ", amfUeId, " RAN UE ID: ", ranUeId)
}

func getUeFromContext(gnb *context.GNBContext, ranUeId int64, amfUeId int64) *context.GNBUe {
	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Warn("[GNB][NGAP] RAN UE NGAP ID ", ranUeId, " not found or already cleaned up")
		return nil
		// TODO SEND ERROR INDICATION
	}

	// Validate UE is not in Down state
	if ue.GetState() == context.Down {
		log.Warn("[GNB][NGAP] UE ", ranUeId, " is in Down state, ignoring NGAP message")
		return nil
	}

	ue.SetAmfUeId(amfUeId)

	return ue
}

func causeToString(cause *ngapType.Cause) string {
	if cause != nil {
		switch cause.Present {
		case ngapType.CausePresentRadioNetwork:
			return "radioNetwork: " + causeRadioNetworkToString(cause.RadioNetwork)
		case ngapType.CausePresentTransport:
			return "transport: " + causeTransportToString(cause.Transport)
		case ngapType.CausePresentNas:
			return "nas: " + causeNasToString(cause.Nas)
		case ngapType.CausePresentProtocol:
			return "protocol: " + causeProtocolToString(cause.Protocol)
		case ngapType.CausePresentMisc:
			return "misc: " + causeMiscToString(cause.Misc)
		}
	}
	return "Cause not found"
}

func causeRadioNetworkToString(network *ngapType.CauseRadioNetwork) string {
	switch network.Value {
	case ngapType.CauseRadioNetworkPresentUnspecified:
		return "Unspecified cause for radio network"
	case ngapType.CauseRadioNetworkPresentTxnrelocoverallExpiry:
		return "Transfer the overall timeout of radio resources during handover"
	case ngapType.CauseRadioNetworkPresentSuccessfulHandover:
		return "Successful handover"
	case ngapType.CauseRadioNetworkPresentReleaseDueToNgranGeneratedReason:
		return "Release due to NG-RAN generated reason"
	case ngapType.CauseRadioNetworkPresentReleaseDueTo5gcGeneratedReason:
		return "Release due to 5GC generated reason"
	case ngapType.CauseRadioNetworkPresentHandoverCancelled:
		return "Handover cancelled"
	case ngapType.CauseRadioNetworkPresentPartialHandover:
		return "Partial handover"
	case ngapType.CauseRadioNetworkPresentHoFailureInTarget5GCNgranNodeOrTargetSystem:
		return "Handover failure in target 5GC NG-RAN node or target system"
	case ngapType.CauseRadioNetworkPresentHoTargetNotAllowed:
		return "Handover target not allowed"
	case ngapType.CauseRadioNetworkPresentTngrelocoverallExpiry:
		return "Transfer the overall timeout of radio resources during target NG-RAN relocation"
	case ngapType.CauseRadioNetworkPresentTngrelocprepExpiry:
		return "Transfer the preparation timeout of radio resources during target NG-RAN relocation"
	case ngapType.CauseRadioNetworkPresentCellNotAvailable:
		return "Cell not available"
	case ngapType.CauseRadioNetworkPresentUnknownTargetID:
		return "Unknown target ID"
	case ngapType.CauseRadioNetworkPresentNoRadioResourcesAvailableInTargetCell:
		return "No radio resources available in the target cell"
	case ngapType.CauseRadioNetworkPresentUnknownLocalUENGAPID:
		return "Unknown local UE NGAP ID"
	case ngapType.CauseRadioNetworkPresentInconsistentRemoteUENGAPID:
		return "Inconsistent remote UE NGAP ID"
	case ngapType.CauseRadioNetworkPresentHandoverDesirableForRadioReason:
		return "Handover desirable for radio reason"
	case ngapType.CauseRadioNetworkPresentTimeCriticalHandover:
		return "Time-critical handover"
	case ngapType.CauseRadioNetworkPresentResourceOptimisationHandover:
		return "Resource optimization handover"
	case ngapType.CauseRadioNetworkPresentReduceLoadInServingCell:
		return "Reduce load in serving cell"
	case ngapType.CauseRadioNetworkPresentUserInactivity:
		return "User inactivity"
	case ngapType.CauseRadioNetworkPresentRadioConnectionWithUeLost:
		return "Radio connection with UE lost"
	case ngapType.CauseRadioNetworkPresentRadioResourcesNotAvailable:
		return "Radio resources not available"
	case ngapType.CauseRadioNetworkPresentInvalidQosCombination:
		return "Invalid QoS combination"
	case ngapType.CauseRadioNetworkPresentFailureInRadioInterfaceProcedure:
		return "Failure in radio interface procedure"
	case ngapType.CauseRadioNetworkPresentInteractionWithOtherProcedure:
		return "Interaction with other procedure"
	case ngapType.CauseRadioNetworkPresentUnknownPDUSessionID:
		return "Unknown PDU session ID"
	case ngapType.CauseRadioNetworkPresentUnkownQosFlowID:
		return "Unknown QoS flow ID"
	case ngapType.CauseRadioNetworkPresentMultiplePDUSessionIDInstances:
		return "Multiple PDU session ID instances"
	case ngapType.CauseRadioNetworkPresentMultipleQosFlowIDInstances:
		return "Multiple QoS flow ID instances"
	case ngapType.CauseRadioNetworkPresentEncryptionAndOrIntegrityProtectionAlgorithmsNotSupported:
		return "Encryption and/or integrity protection algorithms not supported"
	case ngapType.CauseRadioNetworkPresentNgIntraSystemHandoverTriggered:
		return "NG intra-system handover triggered"
	case ngapType.CauseRadioNetworkPresentNgInterSystemHandoverTriggered:
		return "NG inter-system handover triggered"
	case ngapType.CauseRadioNetworkPresentXnHandoverTriggered:
		return "Xn handover triggered"
	case ngapType.CauseRadioNetworkPresentNotSupported5QIValue:
		return "Not supported 5QI value"
	case ngapType.CauseRadioNetworkPresentUeContextTransfer:
		return "UE context transfer"
	case ngapType.CauseRadioNetworkPresentImsVoiceEpsFallbackOrRatFallbackTriggered:
		return "IMS voice EPS fallback or RAT fallback triggered"
	case ngapType.CauseRadioNetworkPresentUpIntegrityProtectionNotPossible:
		return "UP integrity protection not possible"
	case ngapType.CauseRadioNetworkPresentUpConfidentialityProtectionNotPossible:
		return "UP confidentiality protection not possible"
	case ngapType.CauseRadioNetworkPresentSliceNotSupported:
		return "Slice not supported"
	case ngapType.CauseRadioNetworkPresentUeInRrcInactiveStateNotReachable:
		return "UE in RRC inactive state not reachable"
	case ngapType.CauseRadioNetworkPresentRedirection:
		return "Redirection"
	case ngapType.CauseRadioNetworkPresentResourcesNotAvailableForTheSlice:
		return "Resources not available for the slice"
	case ngapType.CauseRadioNetworkPresentUeMaxIntegrityProtectedDataRateReason:
		return "UE maximum integrity protected data rate reason"
	case ngapType.CauseRadioNetworkPresentReleaseDueToCnDetectedMobility:
		return "Release due to CN detected mobility"
	default:
		return "Unknown cause for radio network"
	}
}

func causeTransportToString(transport *ngapType.CauseTransport) string {
	switch transport.Value {
	case ngapType.CauseTransportPresentTransportResourceUnavailable:
		return "Transport resource unavailable"
	case ngapType.CauseTransportPresentUnspecified:
		return "Unspecified cause for transport"
	default:
		return "Unknown cause for transport"
	}
}

func causeNasToString(nas *ngapType.CauseNas) string {
	switch nas.Value {
	case ngapType.CauseNasPresentNormalRelease:
		return "Normal release"
	case ngapType.CauseNasPresentAuthenticationFailure:
		return "Authentication failure"
	case ngapType.CauseNasPresentDeregister:
		return "Deregister"
	case ngapType.CauseNasPresentUnspecified:
		return "Unspecified cause for NAS"
	default:
		return "Unknown cause for NAS"
	}
}

func causeProtocolToString(protocol *ngapType.CauseProtocol) string {
	switch protocol.Value {
	case ngapType.CauseProtocolPresentTransferSyntaxError:
		return "Transfer syntax error"
	case ngapType.CauseProtocolPresentAbstractSyntaxErrorReject:
		return "Abstract syntax error - Reject"
	case ngapType.CauseProtocolPresentAbstractSyntaxErrorIgnoreAndNotify:
		return "Abstract syntax error - Ignore and notify"
	case ngapType.CauseProtocolPresentMessageNotCompatibleWithReceiverState:
		return "Message not compatible with receiver state"
	case ngapType.CauseProtocolPresentSemanticError:
		return "Semantic error"
	case ngapType.CauseProtocolPresentAbstractSyntaxErrorFalselyConstructedMessage:
		return "Abstract syntax error - Falsely constructed message"
	case ngapType.CauseProtocolPresentUnspecified:
		return "Unspecified cause for protocol"
	default:
		return "Unknown cause for protocol"
	}
}

func causeMiscToString(misc *ngapType.CauseMisc) string {
	switch misc.Value {
	case ngapType.CauseMiscPresentControlProcessingOverload:
		return "Control processing overload"
	case ngapType.CauseMiscPresentNotEnoughUserPlaneProcessingResources:
		return "Not enough user plane processing resources"
	case ngapType.CauseMiscPresentHardwareFailure:
		return "Hardware failure"
	case ngapType.CauseMiscPresentOmIntervention:
		return "OM (Operations and Maintenance) intervention"
	case ngapType.CauseMiscPresentUnknownPLMN:
		return "Unknown PLMN (Public Land Mobile Network)"
	case ngapType.CauseMiscPresentUnspecified:
		return "Unspecified cause for miscellaneous"
	default:
		return "Unknown cause for miscellaneous"
	}
}
