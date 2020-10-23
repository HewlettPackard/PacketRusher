package mm_5gs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/nas_control/sm_5gs"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/openapi/models"
)

func GetUlNasTransport_PduSessionEstablishmentRequest(pduSessionId uint8, requestType uint8, dnnString string, sNssai *models.Snssai) (nasPdu []byte) {

	pduSessionEstablishmentRequest := sm_5gs.GetPduSessionEstablishmentRequest(pduSessionId)

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeULNASTransport)

	ulNasTransport := nasMessage.NewULNASTransport(0)
	ulNasTransport.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	ulNasTransport.SetMessageType(nas.MsgTypeULNASTransport)
	ulNasTransport.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	ulNasTransport.PduSessionID2Value = new(nasType.PduSessionID2Value)
	ulNasTransport.PduSessionID2Value.SetIei(nasMessage.ULNASTransportPduSessionID2ValueType)
	ulNasTransport.PduSessionID2Value.SetPduSessionID2Value(pduSessionId)
	ulNasTransport.RequestType = new(nasType.RequestType)
	ulNasTransport.RequestType.SetIei(nasMessage.ULNASTransportRequestTypeType)
	ulNasTransport.RequestType.SetRequestTypeValue(requestType)
	if dnnString != "" {
		dnn := []byte(dnnString)
		ulNasTransport.DNN = new(nasType.DNN)
		ulNasTransport.DNN.SetIei(nasMessage.ULNASTransportDNNType)
		ulNasTransport.DNN.SetLen(uint8(len(dnn)))
		ulNasTransport.DNN.SetDNN(dnn)
	}
	if sNssai != nil {
		var sdTemp [3]uint8
		sd, _ := hex.DecodeString(sNssai.Sd)
		copy(sdTemp[:], sd)
		ulNasTransport.SNSSAI = nasType.NewSNSSAI(nasMessage.ULNASTransportSNSSAIType)
		ulNasTransport.SNSSAI.SetLen(4)
		ulNasTransport.SNSSAI.SetSST(uint8(sNssai.Sst))
		ulNasTransport.SNSSAI.SetSD(sdTemp)
	}

	ulNasTransport.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(nasMessage.PayloadContainerTypeN1SMInfo)
	ulNasTransport.PayloadContainer.SetLen(uint16(len(pduSessionEstablishmentRequest)))
	ulNasTransport.PayloadContainer.SetPayloadContainerContents(pduSessionEstablishmentRequest)

	m.GmmMessage.ULNASTransport = ulNasTransport

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
