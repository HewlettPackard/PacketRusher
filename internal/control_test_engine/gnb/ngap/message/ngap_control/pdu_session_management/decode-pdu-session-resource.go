package pdu_session_management

import (
	"my5G-RANTester/lib/aper"
	"my5G-RANTester/lib/ngap/ngapType"
)

func getGtpTeidFromNgUpUpTnlInformation(payload []byte) ([]byte, error) {

	// per aligned PDUSessionResourceSetupRequestTransfer value( octet string )
	pdu := &ngapType.PDUSessionResourceSetupRequestTransfer{}
	err := aper.UnmarshalWithParams(payload, pdu, "valueExt")
	if err != nil {
		return nil, err
	}
	for _, ies := range pdu.ProtocolIEs.List {
		// get UL NGUP UP TNL Information.
		if ies.Id.Value == ngapType.ProtocolIEIDULNGUUPTNLInformation {
			gtpTeid := []byte(ies.Value.ULNGUUPTNLInformation.GTPTunnel.GTPTEID.Value)
			return gtpTeid, nil
		}
	}

	return nil, nil
}

func getResourceSetupTransferFromPDUSessionResourceSetupRequest(msg *ngapType.PDUSessionResourceSetupRequest) []byte {
	for _, ie := range msg.ProtocolIEs.List {
		if ie.Id.Value == ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq {
			// get ie PDUSessionResourceSetupListSUReq
			pDUSessionResourceSetupList := ie.Value.PDUSessionResourceSetupListSUReq
			for _, item := range pDUSessionResourceSetupList.List {
				// get PDUSessionResourceSetupRequestTransfer value( octet string )
				payload := []byte(item.PDUSessionResourceSetupRequestTransfer)
				return payload
			}
		}
	}
	return nil
}

func GetGtpTeid(ngap *ngapType.NGAPPDU) ([]byte, error) {
	ngapResource := getResourceSetupTransferFromPDUSessionResourceSetupRequest(ngap.InitiatingMessage.Value.PDUSessionResourceSetupRequest)
	gtpTeid, err := getGtpTeidFromNgUpUpTnlInformation(ngapResource)
	return gtpTeid, err
}
