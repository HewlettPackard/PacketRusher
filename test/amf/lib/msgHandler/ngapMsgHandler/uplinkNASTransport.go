package ngapMsgHandler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
)

func UplinkNASTransport(req *ngapType.UplinkNASTransport, amf context.AMFContext, gnb context.GNBContext, nasHandler func(*ngapType.NASPDU, *context.UEContext) (uint8, error)) ([]byte, error) {

	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error
	var naspdu *ngapType.NASPDU
	var msg []byte
	nrLocation := models.NrLocation{}

	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ranUe, err = amf.FindUEByRanId(req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value)
			if err != nil {
				return msg, err
			}
		case ngapType.ProtocolIEIDAMFUENGAPID:
			ue, err = amf.FindUEById(req.ProtocolIEs.List[ie].Value.AMFUENGAPID.Value)
			if err != nil {
				return msg, err
			}

		case ngapType.ProtocolIEIDNASPDU:
			naspdu = req.ProtocolIEs.List[ie].Value.NASPDU

		case ngapType.ProtocolIEIDUserLocationInformation:
			UserLocationInformationNR := req.ProtocolIEs.List[ie].Value.UserLocationInformation.UserLocationInformationNR
			tai := ngapConvert.TaiToModels(UserLocationInformationNR.TAI)
			nrLocation.Tai = &tai
			ncgi := models.Ncgi{}
			ncgi.NrCellId = ngapConvert.BitStringToHex(&UserLocationInformationNR.NRCGI.NRCellIdentity.Value)
			plmn := ngapConvert.PlmnIdToModels(UserLocationInformationNR.NRCGI.PLMNIdentity)
			ncgi.PlmnId = &plmn
			nrLocation.Ncgi = &ncgi
			nrLocation.GlobalGnbId = gnb.GetGlobalRanNodeID()

		case ngapType.ProtocolIEIDRRCEstablishmentCause:

		case ngapType.ProtocolIEIDUEContextRequest:

		default:
			return msg, errors.New("[AMF][NGAP] Received unknown ie for UplinkNASTransport")
		}
	}
	if ue != ranUe {
		return msg, errors.New("[AMF][NGAP] RanUeNgapId does not match the one registred for this UE")
	}
	nastype, err := nasHandler(naspdu, ue)
	if err != nil {
		return msg, err
	}
	switch nastype {
	case nas.MsgTypeRegistrationAccept:
		msg, err = InitialContextSetupRequest(ue, amf)
	default:
		msg, err = DownlinkNASTransport(nastype, ue)
	}
	if err != nil {
		return msg, err
	}
	log.Info("[AMF][NGAP] Sent UplinkNASTransport")

	return msg, nil
}
