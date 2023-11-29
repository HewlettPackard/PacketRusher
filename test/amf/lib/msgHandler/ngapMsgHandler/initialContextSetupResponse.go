package ngapMsgHandler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"
)

func InitialContextSetupResponse(req *ngapType.InitialContextSetupResponse, amf context.AMFContext) error {
	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error

	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ranUe, err = amf.FindUEByRanId(req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value)
			if err != nil {
				return err
			}
		case ngapType.ProtocolIEIDAMFUENGAPID:
			ue, err = amf.FindUEById(req.ProtocolIEs.List[ie].Value.AMFUENGAPID.Value)
			if err != nil {
				return err
			}
		}
	}
	if ue != ranUe {
		return errors.New("[AMF][NGAP] RanUeNgapId does not match the one registred for this UE")
	}
	return nil
}
