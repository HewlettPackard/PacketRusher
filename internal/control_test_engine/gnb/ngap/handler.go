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

	"github.com/free5gc/ngap/aper"

	ngapType "github.com/free5gc/ngap/ie"
	ngapmsg "github.com/free5gc/ngap/message"
	log "github.com/sirupsen/logrus"
	_ "github.com/vishvananda/netlink"
	ngapConvert "my5G-RANTester/lib/ngap"
)

func HandlerDownlinkNasTransport(gnb *context.GNBContext, message *ngapmsg.DownlinkNASTransport) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte

	valueMessage := message

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

	}
	if valueMessage.NASPDU != nil {

		messageNas = valueMessage.NASPDU.Value

	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot send DownlinkNASTransport message to UE with RANUEID %d as it does not know this UE", ranUeId)
		return
	}

	// send NAS message to UE.
	sender.SendToUe(ue, messageNas)
}

func HandlerInitialContextSetupRequest(gnb *context.GNBContext, message *ngapmsg.InitialContextSetupRequest) {

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

	valueMessage := message

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

	}
	if valueMessage.NASPDU != nil {

		messageNas = valueMessage.NASPDU.Value

	}
	if valueMessage.SecurityKey != nil {

		// TODO using for create new security context between GNB and UE.

		// securityKey = valueMessage.SecurityKey.Value.Bytes

	}
	if valueMessage.GUAMI != nil {

	}
	if valueMessage.AllowedNSSAI != nil {

		valor := len(valueMessage.AllowedNSSAI.List)
		sst = make([]string, valor)
		sd = make([]string, valor)

		// list S-NSSAI(Single – Network Slice Selection Assistance Information).
		for i, items := range valueMessage.AllowedNSSAI.List {

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

	}
	if valueMessage.MobilityRestrictionList != nil {

		// that field is not mandatory.

		mobilityRestrict = fmt.Sprintf("%x", valueMessage.MobilityRestrictionList.ServingPLMN.Value)

	}
	if valueMessage.MaskedIMEISV != nil {

		// that field is not mandatory.
		// TODO using for mapping UE context

		maskedImeisv = fmt.Sprintf("%x", valueMessage.MaskedIMEISV.Value.Bytes)

	}
	if valueMessage.UESecurityCapabilities != nil {

		// TODO using for create new security context between UE and GNB.
		// TODO algorithms for create new security context between UE and GNB.

		ueSecurityCapabilities = valueMessage.UESecurityCapabilities

	}
	if valueMessage.PDUSessionResourceSetupListCxtReq != nil {

		pDUSessionResourceSetupListCxtReq = valueMessage.PDUSessionResourceSetupListCxtReq

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
			err := ngapConvert.Unmarshal(pDUSessionResourceSetupRequestTransferBytes, pDUSessionResourceSetupRequestTransfer)
			if err != nil {
				log.Error("[GNB] Unable to unmarshall PDUSessionResourceSetupRequestTransfer: ", err)
				continue
			}

			var gtpTunnel *ngapType.GTPTunnel
			var upfIp string
			var teidUplink aper.OctetString
			for _, ie := range pDUSessionResourceSetupRequestTransfer.ProtocolIEs.List {
				switch ie.Id().Value {

				case ngapType.ProtocolIEIDULNGUUPTNLInformation:
					uLNGUUPTNLInformation := ie.ULNGUUPTNLInformation

					gtpTunnel, err = ngapConvert.Tunnel(uLNGUUPTNLInformation)
					if err != nil {
						log.Error("[GNB][NGAP] Invalid uplink tunnel: ", err)
						continue
					}
					upfIp, _ = ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
					teidUplink = gtpTunnel.GTPTEID.Value
				}
			}

			// Without a UL tunnel teidUplink is empty, and reading it would panic.
			if gtpTunnel == nil {
				log.Error("[GNB][NGAP] No UL NG-U UP TNL Information for PDU Session ", pduSessionId, ", skipping")
				continue
			}

			_, err = ue.CreatePduSession(pduSessionId, upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue))
			if err != nil {
				log.Error("[GNB] ", err)
			}

			if pDUSessionResourceSetupItemCtxReq.NASPDU != nil {
				sender.SendToUe(ue, pDUSessionResourceSetupItemCtxReq.NASPDU.Value)
			}
		}

		msg := context.UEMessage{GNBPduSessions: ue.GetPduSessions(), GnbIp: gnb.GetN3GnbIp(), GtpDevice: gnb.GetGtpDevice()}
		sender.SendMessageToUe(ue, msg)
	}

	// send Initial Context Setup Response.
	log.Info("[GNB][NGAP][AMF] Send Initial Context Setup Response.")
	trigger.SendInitialContextSetupResponse(ue, gnb)
}

func HandlerPduSessionResourceSetupRequest(gnb *context.GNBContext, message *ngapmsg.PDUSessionResourceSetupRequest) {

	var ranUeId int64
	var amfUeId int64
	var pDUSessionResourceSetupList *ngapType.PDUSessionResourceSetupListSUReq

	valueMessage := message

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

	}
	if valueMessage.PDUSessionResourceSetupListSUReq != nil {

		pDUSessionResourceSetupList = valueMessage.PDUSessionResourceSetupListSUReq

	}

	// The loop above only catches the IE present but empty; an absent one leaves
	// the list nil.
	if pDUSessionResourceSetupList == nil {
		log.Error("[GNB][NGAP] PDU Session Resource Setup Request is missing mandatory PDU SESSION RESOURCE SETUP LIST SU REQ")
		return
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
		// NAS PDU is optional (TS 38.413 V19.0.0.0 section 9.2.1.1)
		if item.PDUSessionNASPDU != nil {
			messageNas = item.PDUSessionNASPDU.Value
		} else {
			log.Info("[GNB][NGAP] NAS PDU is missing")
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

			err := ngapConvert.Unmarshal(item.PDUSessionResourceSetupRequestTransfer, pdu)
			if err == nil {
				for _, ies := range pdu.ProtocolIEs.List {

					switch ies.Id().Value {

					case ngapType.ProtocolIEIDULNGUUPTNLInformation:
						tunnel, err := ngapConvert.Tunnel(ies.ULNGUUPTNLInformation)
						if err != nil {
							log.Error("[GNB][NGAP] Invalid uplink tunnel: ", err)
							continue
						}
						ulTeid = binary.BigEndian.Uint32(tunnel.GTPTEID.Value)
						upfAddress = tunnel.TransportLayerAddress.Value.Bytes

					case ngapType.ProtocolIEIDQosFlowSetupRequestList:
						for _, itemsQos := range ies.QosFlowSetupRequestList.List {
							qosId = itemsQos.QosFlowIdentifier.Value
							// QoS Characteristics is a CHOICE: a dynamic 5QI leaves NonDynamic5QI
							// nil and carries its 5QI only optionally.
							qosCharacteristics := itemsQos.QosFlowLevelQosParameters.QosCharacteristics
							switch descriptor := qosCharacteristics.Choice.(type) {
							case *ngapType.NonDynamic5QIDescriptor:
								fiveQi = descriptor.FiveQI.Value
							case *ngapType.Dynamic5QIDescriptor:
								if descriptor.FiveQI != nil {
									fiveQi = descriptor.FiveQI.Value
								}
							}
							priArp = itemsQos.QosFlowLevelQosParameters.AllocationAndRetentionPriority.PriorityLevelARP.Value
						}

					case ngapType.ProtocolIEIDPDUSessionAggregateMaximumBitRate:

					case ngapType.ProtocolIEIDPDUSessionType:
						pduSType = uint64(ies.PDUSessionType.Value)

					case ngapType.ProtocolIEIDSecurityIndication:

					}
				}
			} else {
				log.Info("[GNB][NGAP] Error in decode Pdu Session Resource Setup Request Transfer")
			}
		} else {
			log.Error("[GNB][NGAP] Error in Pdu Session Resource Setup Request, Pdu Session Resource Setup Request Transfer is missing")
			continue
		}

		// A transfer that failed to decode, or carried no UL tunnel, leaves upfAddress
		// empty; indexing it would panic and end the process just as the Fatal above did.
		if len(upfAddress) < 4 {
			log.Error("[GNB][NGAP] No usable UL NG-U UP TNL Information for PDU Session ", pduSessionId, ", skipping")
			continue
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

		// send NAS message to UE if present.
		if messageNas != nil {
			sender.SendToUe(ue, messageNas)
		}

		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession
		msg := context.UEMessage{GnbIp: gnb.GetN3GnbIp(), GtpDevice: gnb.GetGtpDevice(), GNBPduSessions: pduSessions}

		sender.SendMessageToUe(ue, msg)
	}

	// Every requested session was skipped, so there is nothing to acknowledge. Building a
	// Setup Response from an empty list yields a SEQUENCE with no items, which the encoder
	// rejects on its lower bound -- and that was fatal, taking the whole simulator and
	// every other UE on this gNB with it. TS 38.413 wants these reported in a PDU Session
	// Resource Failed To Setup list, but that IE is an empty stub here: it references a
	// transfer builder that was never written. Until that exists, staying up and saying so
	// beats sending a malformed message or dying.
	if len(configuredPduSessions) == 0 {
		log.Error("[GNB][NGAP] No PDU session could be set up for RAN UE ", ue.GetRanUeId(),
			"; sending no Setup Response (failed-to-setup reporting is unimplemented)")
		return
	}

	// send PDU Session Resource Setup Response.
	trigger.SendPduSessionResourceSetupResponse(configuredPduSessions, ue, gnb)
}

func HandlerPduSessionReleaseCommand(gnb *context.GNBContext, message *ngapmsg.PDUSessionResourceReleaseCommand) {
	valueMessage := message

	var amfUeId int64
	var ranUeId int64
	var messageNas aper.OctetString
	var pduSessionIds []*ngapType.PDUSessionID

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

	}
	if valueMessage.NASPDU != nil {

		messageNas = valueMessage.NASPDU.Value

	}
	if valueMessage.PDUSessionResourceToReleaseListRelCmd != nil {

		pDUSessionRessourceToReleaseListRelCmd := valueMessage.PDUSessionResourceToReleaseListRelCmd

		for _, pDUSessionRessourceToReleaseItemRelCmd := range pDUSessionRessourceToReleaseListRelCmd.List {
			pduSessionIds = append(pduSessionIds, pDUSessionRessourceToReleaseItemRelCmd.PDUSessionID)
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

func HandlerNgSetupResponse(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.NGSetupResponse) {

	err := false
	var plmn string

	// check information about AMF and add in AMF context. A re-established association
	// runs NG Setup again, so start from empty lists rather than appending.
	amf.ResetSupported()
	valueMessage := message

	if valueMessage.AMFName != nil {

		amfName := valueMessage.AMFName.Value
		amf.SetAmfName(string(amfName))

	}
	if valueMessage.ServedGUAMIList != nil {

		if valueMessage.ServedGUAMIList.List == nil {
			// TODO error indication. This field is mandatory critically reject
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Serverd Guami list is missing")
			log.Info("[GNB][NGAP] AMF is inactive")
			err = true
		}
		for _, items := range valueMessage.ServedGUAMIList.List {
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

	}
	if valueMessage.RelativeAMFCapacity != nil {

		amfCapacity := valueMessage.RelativeAMFCapacity.Value
		amf.SetAmfCapacity(amfCapacity)

	}
	if valueMessage.PLMNSupportList != nil {

		for _, items := range valueMessage.PLMNSupportList.List {

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

func HandlerNgSetupFailure(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.NGSetupFailure) {

	// check information about AMF and add in AMF context.
	valueMessage := message

	if valueMessage.Cause != nil {

		log.Error("[GNB][NGAP] Received failure from AMF: ", causeToString(valueMessage.Cause))

	}
	if valueMessage.TimeToWait != nil {

		switch valueMessage.TimeToWait.Value {

		case ngapType.TimeToWaitPresentV1s:
		case ngapType.TimeToWaitPresentV2s:
		case ngapType.TimeToWaitPresentV5s:
		case ngapType.TimeToWaitPresentV10s:
		case ngapType.TimeToWaitPresentV20s:
		case ngapType.TimeToWaitPresentV60s:

		}

	}
	if valueMessage.CriticalityDiagnostics != nil {

		// TODO treatment error

		// valueMessage.CriticalityDiagnostics
		// errors.IECriticality.Value
		// ngapType.CriticalityPresentReject:
		// ngapType.CriticalityPresentIgnore:
		// ngapType.CriticalityPresentNotify:
		// ngapType.TypeOfErrorPresentNotUnderstood:
		// ngapType.TypeOfErrorPresentMissing:

	}

	// redundant but useful for information about code.
	amf.SetStateInactive()

	log.Info("[GNB][NGAP] AMF is inactive")
}

func HandlerUeContextReleaseCommand(gnb *context.GNBContext, message *ngapmsg.UEContextReleaseCommand) {

	valueMessage := message

	var cause *ngapType.Cause
	var ue_ids *ngapType.UENGAPIDs

	if valueMessage.UENGAPIDs != nil {

		ue_ids = valueMessage.UENGAPIDs

	}
	if valueMessage.Cause != nil {

		cause = valueMessage.Cause

	}

	// UE NGAP IDs is a CHOICE (TS 38.413 9.2.2.5): the AMF may identify the UE
	// by the pair or by its AMF UE NGAP ID alone.
	var ue *context.GNBUe
	var err error
	if ue_ids == nil {
		log.Warn("[GNB][NGAP] UE Context Release Command missing UE ID")
		return
	}
	switch identifier := ue_ids.Choice.(type) {
	case *ngapType.UENGAPIDPair:
		if identifier.RANUENGAPID == nil {
			return
		}
		ue, err = gnb.GetGnbUe(identifier.RANUENGAPID.Value)
	case *ngapType.AMFUENGAPID:
		ue, err = gnb.GetGnbUeByAmfUeId(identifier.Value)
	default:
		log.Warn("[GNB][NGAP] UE Context Release Command missing UE ID")
		return
	}
	if err != nil || ue == nil {
		log.Warn("[GNB][NGAP] AMF is trying to free the context of an unknown UE, ignoring")
		return
	}

	log.Info("[GNB][NGAP] Releasing UE Context for UE ", ue.GetRanUeId(), ", cause: ", causeToString(cause))

	// Send UEContextReleaseComplete before deleting context
	trigger.SendUeContextReleaseComplete(ue)

	// Clean up UE context
	gnb.DeleteGnBUe(ue)
}

func HandlerAmfConfigurationUpdate(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.AMFConfigurationUpdate) {
	log.Debugf("Before Update:")
	for oldAmf := range gnb.IterGnbAmf() {
		tnla := oldAmf.GetTNLA()
		log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
			oldAmf.GetAmfName(), oldAmf.GetAmfIpPort().Addr(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
	}

	var amfName string
	var amfCapacity int64
	var amfRegionId, amfSetId, amfPointer aper.BitString

	valueMessage := message
	if valueMessage.AMFName != nil {

		amfName = string(valueMessage.AMFName.Value)

	}
	if valueMessage.ServedGUAMIList != nil {

		for _, servedGuamiItem := range valueMessage.ServedGUAMIList.List {
			amfRegionId = servedGuamiItem.GUAMI.AMFRegionID.Value
			amfSetId = servedGuamiItem.GUAMI.AMFSetID.Value
			amfPointer = servedGuamiItem.GUAMI.AMFPointer.Value
		}

	}
	if valueMessage.RelativeAMFCapacity != nil {

		amfCapacity = valueMessage.RelativeAMFCapacity.Value

	}
	if valueMessage.AMFTNLAssociationToAddList != nil {

		toAddList := valueMessage.AMFTNLAssociationToAddList
		for _, toAddItem := range toAddList.List {
			ipv4String, _ := ngapConvert.IPAddressToString(ngapConvert.CPAddress(toAddItem.AMFTNLAssociationAddress))
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

	}
	if valueMessage.AMFTNLAssociationToRemoveList != nil {

		toRemoveList := valueMessage.AMFTNLAssociationToRemoveList
		for _, toRemoveItem := range toRemoveList.List {
			ipv4String, _ := ngapConvert.IPAddressToString(ngapConvert.CPAddress(toRemoveItem.AMFTNLAssociationAddress))
			if ipv4String == "" {
				// ignore AMF that does not have IPv4 address
				continue
			}
			ipv4Port := netip.AddrPortFrom(netip.MustParseAddr(ipv4String), 38412) // with default sctp port

			oldAmf := gnb.FindGnbAmfByIpPort(ipv4Port)
			if oldAmf == nil {
				continue
			}

			log.Info("[GNB][AMF] Remove AMF:", oldAmf.GetAmfName(), " IP:", oldAmf.GetAmfIpPort().Addr())
			gnb.RemoveGnbAmf(oldAmf) // Close SCTP Conntection
		}

	}
	if valueMessage.AMFTNLAssociationToUpdateList != nil {

		toUpdateList := valueMessage.AMFTNLAssociationToUpdateList
		for _, toUpdateItem := range toUpdateList.List {
			ipv4String, _ := ngapConvert.IPAddressToString(ngapConvert.CPAddress(toUpdateItem.AMFTNLAssociationAddress))
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

	log.Debugf("After Update:")
	for oldAmf := range gnb.IterGnbAmf() {
		tnla := oldAmf.GetTNLA()
		log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
			oldAmf.GetAmfName(), oldAmf.GetAmfIpPort().Addr(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
	}

	trigger.SendAmfConfigurationUpdateAcknowledge(amf)
}

func HandlerAmfStatusIndication(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.AMFStatusIndication) {
	valueMessage := message
	if valueMessage.UnavailableGUAMIList != nil {

		for _, unavailableGuamiItem := range valueMessage.UnavailableGUAMIList.List {
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
					oldAmf.GetAmfName() == string(unavailableGuamiItem.BackupAMFName.Value) {
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

						gnb.RemoveGnbAmf(oldAmf)

						break
					}
				}
			}
		}

	}

}

func HandlerPathSwitchRequestAcknowledge(gnb *context.GNBContext, message *ngapmsg.PathSwitchRequestAcknowledge) {
	var pduSessionResourceSwitchedList *ngapType.PDUSessionResourceSwitchedList
	valueMessage := message

	var amfUeId, ranUeId int64

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

	}
	if valueMessage.PDUSessionResourceSwitchedList != nil {

		pduSessionResourceSwitchedList = valueMessage.PDUSessionResourceSwitchedList
		if pduSessionResourceSwitchedList == nil {
			log.Error("[GNB][NGAP] PduSessionResourceSwitchedList is missing")
			// TODO SEND ERROR INDICATION
			return
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

	completed := 0
	for _, pduSessionResourceSwitchedItem := range pduSessionResourceSwitchedList.List {
		pduSessionId := pduSessionResourceSwitchedItem.PDUSessionID.Value
		pduSession, err := ue.GetPduSession(pduSessionId)
		// An empty slot is (nil, nil), not an error.
		if err != nil || pduSession == nil {
			log.Error("[GNB] Trying to path switch an unknown PDU Session ID ", pduSessionId, ": ", err)
			continue
		}

		pathSwitchRequestAcknowledgeTransferBytes := pduSessionResourceSwitchedItem.PathSwitchRequestAcknowledgeTransfer
		pathSwitchRequestAcknowledgeTransfer := &ngapType.PathSwitchRequestAcknowledgeTransfer{}
		err = ngapConvert.Unmarshal(pathSwitchRequestAcknowledgeTransferBytes, pathSwitchRequestAcknowledgeTransfer)
		if err != nil {
			log.Error("[GNB] Unable to unmarshall PathSwitchRequestAcknowledgeTransfer: ", err)
			continue
		}

		if pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation != nil {
			gtpTunnel, err := ngapConvert.Tunnel(pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation)
			if err != nil {
				log.Error("[GNB][NGAP] Invalid path switch tunnel: ", err)
				continue
			}
			upfIpv4, _ := ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
			teidUplink := gtpTunnel.GTPTEID.Value

			// Set new Teid Uplink received in PathSwitchRequestAcknowledge
			pduSession.SetTeidUplink(binary.BigEndian.Uint32(teidUplink))
			pduSession.SetUpfIp(upfIpv4)
		}
		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession

		msg := context.UEMessage{GNBPduSessions: pduSessions, GnbIp: gnb.GetN3GnbIp(), GtpDevice: gnb.GetGtpDevice()}

		sender.SendMessageToUe(ue, msg)
		completed++
	}
	if completed == len(pduSessionResourceSwitchedList.List) {
		ue.SetStateReady()
		log.Info("[GNB] Handover completed successfully for UE ", ue.GetRanUeId())
	}
}

func HandlerHandoverRequest(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.HandoverRequest) {
	var ueSecurityCapabilities *ngapType.UESecurityCapabilities
	var sst []string
	var sd []string
	var maskedImeisv string
	var sourceToTargetContainer *ngapType.SourceToTargetTransparentContainer
	var pDUSessionResourceSetupListHOReq *ngapType.PDUSessionResourceSetupListHOReq
	var amfUeId int64

	valueMessage := message

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.AllowedNSSAI != nil {

		valor := len(valueMessage.AllowedNSSAI.List)
		sst = make([]string, valor)
		sd = make([]string, valor)

		// list S-NSSAI(Single – Network Slice Selection Assistance Information).
		for i, items := range valueMessage.AllowedNSSAI.List {

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

	}
	if valueMessage.MaskedIMEISV != nil {

		// that field is not mandatory.
		// TODO using for mapping UE context

		maskedImeisv = fmt.Sprintf("%x", valueMessage.MaskedIMEISV.Value.Bytes)

	}
	if valueMessage.SourceToTargetTransparentContainer != nil {

		sourceToTargetContainer = valueMessage.SourceToTargetTransparentContainer
		if sourceToTargetContainer == nil {
			log.Error("[GNB][NGAP] sourceToTargetContainer is missing")
			// TODO SEND ERROR INDICATION
			return
		}

	}
	if valueMessage.PDUSessionResourceSetupListHOReq != nil {

		pDUSessionResourceSetupListHOReq = valueMessage.PDUSessionResourceSetupListHOReq
		if pDUSessionResourceSetupListHOReq == nil {
			log.Error("[GNB][NGAP] pDUSessionResourceSetupListHOReq is missing")
			// TODO SEND ERROR INDICATION
			return
		}

	}
	if valueMessage.UESecurityCapabilities != nil {

		ueSecurityCapabilities = valueMessage.UESecurityCapabilities

	}

	if sourceToTargetContainer == nil {
		log.Error("[GNB] HandoverRequest message from AMF is missing mandatory SourceToTargetTransparentContainer")
		return
	}
	// Checked before a UE is created for the handover, so a bad request leaves nothing behind.
	if pDUSessionResourceSetupListHOReq == nil {
		log.Error("[GNB] HandoverRequest message from AMF is missing mandatory PDUSessionResourceSetupListHOReq")
		return
	}

	sourceToTargetContainerBytes := sourceToTargetContainer.Value
	sourceToTargetContainerNgap := &ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
	err := ngapConvert.Unmarshal(sourceToTargetContainerBytes, sourceToTargetContainerNgap)
	if err != nil {
		log.Error("[GNB] Unable to unmarshall SourceToTargetTransparentContainer: ", err)
		return
	}
	prUeId, err := ngapConvert.VirtualUEID(sourceToTargetContainerNgap)
	if err != nil {
		log.Error("[GNB] Invalid virtual UE identity in handover: ", err)
		return
	}

	ue, err := gnb.NewGnBUe(nil, nil, prUeId, nil)
	if ue == nil || err != nil {
		log.Errorf("[GNB] HandoverFailure: %s", err)
		return
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
		err := ngapConvert.Unmarshal(handOverRequestTransferBytes, handOverRequestTransfer)
		if err != nil {
			log.Error("[GNB] Unable to unmarshall HandOverRequestTransfer: ", err)
			continue
		}

		var gtpTunnel *ngapType.GTPTunnel
		var upfIp string
		var teidUplink aper.OctetString
		for _, ie := range handOverRequestTransfer.ProtocolIEs.List {
			switch ie.Id().Value {

			case ngapType.ProtocolIEIDULNGUUPTNLInformation:
				uLNGUUPTNLInformation := ie.ULNGUUPTNLInformation

				gtpTunnel, err = ngapConvert.Tunnel(uLNGUUPTNLInformation)
				if err != nil {
					log.Error("[GNB][NGAP] Invalid uplink tunnel: ", err)
					continue
				}
				upfIp, _ = ngapConvert.IPAddressToString(gtpTunnel.TransportLayerAddress)
				teidUplink = gtpTunnel.GTPTEID.Value
			}
		}

		// Without a UL tunnel teidUplink is empty, and reading it would panic.
		if gtpTunnel == nil {
			log.Error("[GNB][NGAP] No UL NG-U UP TNL Information for PDU Session ", pduSessionId, ", skipping")
			continue
		}

		_, err = ue.CreatePduSession(pduSessionId, upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue))
		if err != nil {
			log.Error("[GNB] ", err)
		}
	}

	trigger.SendHandoverRequestAcknowledge(gnb, ue)
}

func HandlerHandoverCommand(amf *context.GNBAmf, gnb *context.GNBContext, message *ngapmsg.HandoverCommand) {
	valueMessage := message

	var amfUeId, ranUeId int64

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value

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
	connectionLost := make(chan struct{})
	if err := newGnb.QueueHandover(context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, ConnectionLost: connectionLost, PrUeId: ue.GetPrUeId(), IsHandover: true}); err != nil {
		ue.SetHandoverGnodeB(nil)
		log.Error("[GNB] Unable to connect handover target: ", err)
		return
	}

	msg := context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, ConnectionLost: connectionLost, GNBInboundChannel: newGnb.GetInboundChannel()}

	sender.SendMessageToUe(ue, msg)
}

func HandlerPaging(gnb *context.GNBContext, message *ngapmsg.Paging) {

	valueMessage := message

	var uEPagingIdentity *ngapType.UEPagingIdentity
	var tAIListForPaging *ngapType.TAIListForPaging

	if valueMessage.UEPagingIdentity != nil {

		uEPagingIdentity = valueMessage.UEPagingIdentity

	}
	if valueMessage.TAIListForPaging != nil {

		tAIListForPaging = valueMessage.TAIListForPaging

	}

	_ = tAIListForPaging

	var tmsi *ngapType.FiveGSTMSI
	if uEPagingIdentity != nil {
		tmsi, _ = uEPagingIdentity.Choice.(*ngapType.FiveGSTMSI)
	}
	if tmsi == nil {
		log.Error("[GNB][NGAP] Paging is missing mandatory UE Paging Identity")
		return
	}

	gnb.AddPagedUE(tmsi)

	log.Info("[GNB][AMF] Paging UE")
}

func HandlerErrorIndication(gnb *context.GNBContext, message *ngapmsg.ErrorIndication) {

	valueMessage := message

	var amfUeId, ranUeId int64
	var hasAmfUeId, hasRanUeId bool

	if valueMessage.AMFUENGAPID != nil {

		amfUeId = valueMessage.AMFUENGAPID.Value
		hasAmfUeId = true

	}
	if valueMessage.RANUENGAPID != nil {

		ranUeId = valueMessage.RANUENGAPID.Value
		hasRanUeId = true

	}

	log.Warn("[GNB][AMF] Received an Error Indication for UE with AMF UE ID: ", amfUeId, " RAN UE ID: ", ranUeId)

	// Find the UE by RAN UE ID or AMF UE ID
	var ue *context.GNBUe
	if hasRanUeId {
		ue, _ = gnb.GetGnbUe(ranUeId)
	}
	if ue == nil && hasAmfUeId {
		ue, _ = gnb.GetGnbUeByAmfUeId(amfUeId)
	}

	if ue == nil {
		log.Warn("[GNB] No UE context found for Error Indication")
		return
	}

	// If a release was pending, perform local release
	if ue.GetReleaseRequested() {
		ue.SetReleaseRequested(false)
		log.Warn("[GNB] Performing local release of UE context due to Error Indication (release was pending)")
		gnb.DeleteGnBUe(ue)
	}
}

func getUeFromContext(gnb *context.GNBContext, ranUeId int64, amfUeId int64) *context.GNBUe {
	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil {
		log.Warn("[GNB][NGAP] RAN UE NGAP ID ", ranUeId, " not found or already cleaned up")
		return nil
		// TODO SEND ERROR INDICATION
	}

	ue.SetAmfUeId(amfUeId)

	return ue
}

func causeToString(cause *ngapType.Cause) string {
	if cause == nil {
		return "Cause not found"
	}
	switch value := cause.Choice.(type) {
	case *ngapType.CauseRadioNetwork:
		return "radioNetwork: " + causeRadioNetworkToString(value)
	case *ngapType.CauseTransport:
		return "transport: " + causeTransportToString(value)
	case *ngapType.CauseNas:
		return "nas: " + causeNasToString(value)
	case *ngapType.CauseProtocol:
		return "protocol: " + causeProtocolToString(value)
	case *ngapType.CauseMisc:
		return "misc: " + causeMiscToString(value)
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
	case ngapType.CauseMiscPresentUnknownPLMNOrSNPN:
		return "Unknown PLMN (Public Land Mobile Network)"
	case ngapType.CauseMiscPresentUnspecified:
		return "Unspecified cause for miscellaneous"
	default:
		return "Unknown cause for miscellaneous"
	}
}
