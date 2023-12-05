package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func UlNasTransport(nasReq *nas.Message, fgc *context.Aio5gc, ue *context.UEContext) error {

	ulNasTransport := nasReq.ULNASTransport
	var err error
	var smMessage *context.SmContext

	switch ulNasTransport.GetPayloadContainerType() {
	// TS 24.501 5.4.5.2.3 case a)
	case nasMessage.PayloadContainerTypeN1SMInfo:
		smMessage, err = transport5GSMMessage(ue, ulNasTransport, fgc)

	default:
		err = fmt.Errorf("[5GC][NAS] Payload Container type not implemented")
	}

	if err != nil {
		return err
	}
	msg.SendPDUSessionEstablishmentAccept(fgc, ue, smMessage)
	return nil
}

func transport5GSMMessage(ue *context.UEContext, ulNasTransport *nasMessage.ULNASTransport, fgc *context.Aio5gc) (*context.SmContext, error) {
	requestType := ulNasTransport.RequestType
	smMessage := ulNasTransport.PayloadContainer.GetPayloadContainerContents()
	var pduSessionID int32

	if requestType == nil {
		return nil, errors.New("[5GC][NAS] ulNasTransport Request type is nil")
	}

	if id := ulNasTransport.PduSessionID2Value; id != nil {
		pduSessionID = int32(id.GetPduSessionID2Value())
	} else {
		return nil, errors.New("[5GC][NAS] PDU Session ID is nil")
	}

	switch requestType.GetRequestTypeValue() {
	// case iii) if the AMF does not have a PDU session routing context for the PDU session ID and the UE
	// and the Request type IE is included and is set to "initial request"
	case nasMessage.ULNASTransportRequestTypeInitialRequest:
		return context.CreatePDUSession(ulNasTransport, ue, fgc, pduSessionID, smMessage)
	default:
		return nil, errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
}
