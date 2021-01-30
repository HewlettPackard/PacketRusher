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

	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] RAN UE NGAP ID is incorrect")
		// TODO SEND ERROR INDICATION
	}

	// update AMF UE id.
	if ue.GetAmfUeId() == 0 {
		ue.SetAmfUeId(amfUeId)
	} else {
		if ue.GetAmfUeId() != amfUeId {
			log.Fatal("[GNB][NGAP] AMF UE NGAP ID is incorrect")
		}
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
	var mobilityRestrict string
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

				if items.SNSSAI.SD != nil {
					sst[i] = fmt.Sprintf("%x", items.SNSSAI.SST.Value)
				} else {
					sst[i] = "not informed"
				}

				if items.SNSSAI.SST.Value != nil {
					sd[i] = fmt.Sprintf("%x", items.SNSSAI.SD.Value)
				} else {
					sd[i] = "not informed"
				}
			}

		case ngapType.ProtocolIEIDMobilityRestrictionList:
			// that field is not mandatory.
			if ies.Value.MobilityRestrictionList == nil {
				log.Info("[GNB][NGAP] Allowed NSSAI is missing")
			} else {
				mobilityRestrict = fmt.Sprintf("%x", ies.Value.MobilityRestrictionList.ServingPLMN.Value)
			}

		case ngapType.ProtocolIEIDMaskedIMEISV:
			// that field is not mandatory.
			// TODO using for mapping UE context
			if ies.Value.MaskedIMEISV == nil {
				log.Info("[GNB][NGAP] Masked IMEISV is missing")
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

	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Fatal("[GNB][NGAP] RAN UE NGAP ID is incorrect")
		// TODO SEND ERROR INDICATION
	}

	// check if AMF UE id.
	if ue.GetAmfUeId() != amfUeId {
		log.Fatal("[GNB][NGAP] AMF UE NGAP ID is incorrect")
		// TODO SEND ERROR INDICATION
	}

	// create UE context.
	ue.CreateUeContext(mobilityRestrict, maskedImeisv, sst, sd)

	// show UE context.
	log.Info("[GNB][UE] UE CONTEXT CREATE WITH SUCCESSFUL")
	log.Info("[GNB][UE] UE RAN ID", ue.GetRanUeId())
	log.Info("[GNB][UE] UE AMF ID", ue.GetAmfUeId())
	mcc, mnc := ue.GetUeMobility()
	log.Info("[GNB][UE] UE MOBILITY RESTRICT --PLMN-- MCC:", mcc, "MNC:", mnc)
	log.Info("[GNB][UE] UE MASKED IMEISV:", ue.GetUeMaskedImeiSv())
	for i := 0; i < ue.GetLenSlice(); i++ {
		sst, sd := ue.GetAllowedNSSAI(i)
		log.Info("[GNB][UE] ALLOWED NSSAI-- SST:", sst, "SD:", sd)
	}

	// send NAS message to UE.
	log.Info("[GNB][NAS][UE] Send Registration Accept:")
	sender.SendToUe(ue, messageNas)

	// send Initial Context Setup Response.
	log.Info("[GNB][NGAP][AMF] Send Initial Context Setup Response:")
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

				if item.PDUSessionNASPDU != nil {
					messageNas = item.PDUSessionNASPDU.Value
				} else {
					log.Fatal("[GNB][NGAP] NAS PDU is missing")
				}

				// create a PDU session( PDU SESSION ID + NSSAI).
				pduSessionId = item.PDUSessionID.Value
				// item.SNSSAI.SD.Value
				// item.SNSSAI.SST.Value

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
						sd = fmt.Sprintf("%x", slice.SNSSAI.SD.Value)
					} else {
						sd = "was not informed"
					}

					if slice.SNSSAI.SD.Value != nil {
						sst = fmt.Sprintf("%x", slice.SNSSAI.SST.Value)
					} else {
						sst = "was not informed"
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
		log.Info("[GNB][AMF] AMF NAME: ", amf.GetAmfName())
		log.Info("[GNB][AMF] State of AMF: ACTIVE")
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

	log.Fatal("[GNB][NGAP] AMF is inactive")
}
