package nasMsgHandler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/amf/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func NASTransport(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext) (err error) {

	ulNasTransport := nasReq.ULNASTransport

	switch ulNasTransport.GetPayloadContainerType() {
	// TS 24.501 5.4.5.2.3 case a)
	case nasMessage.PayloadContainerTypeN1SMInfo:
		return transport5GSMMessage(ue, ulNasTransport, amf)
	default:
		return fmt.Errorf("[AMF][NAS] Payload Container type not implemented")
	}
}

func transport5GSMMessage(ue *context.UEContext, ulNasTransport *nasMessage.ULNASTransport, amf *context.AMFContext) error {
	requestType := ulNasTransport.RequestType
	smMessage := ulNasTransport.PayloadContainer.GetPayloadContainerContents()
	var pduSessionID int32

	if requestType == nil {
		return errors.New("[AMF][NAS] ulNasTransport Request type is nil")
	}

	if id := ulNasTransport.PduSessionID2Value; id != nil {
		pduSessionID = int32(id.GetPduSessionID2Value())
	} else {
		return errors.New("PDU Session ID is nil")
	}

	switch requestType.GetRequestTypeValue() {
	// case iii) if the AMF does not have a PDU session routing context for the PDU session ID and the UE
	// and the Request type IE is included and is set to "initial request"
	case nasMessage.ULNASTransportRequestTypeInitialRequest:
		err := context.CreatePDUSession(ulNasTransport, ue, amf, pduSessionID, smMessage)
		return err
	default:
		return errors.New("[AMF][NAS] Unimplemented ulNasTransport Request type")
	}
}
