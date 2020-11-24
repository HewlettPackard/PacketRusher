package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/context"
	"my5G-RANTester/internal/control_test_engine/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/nas_control/sm_5gs"
	"my5G-RANTester/internal/control_test_engine/ngap_control/nas_transport"
	"my5G-RANTester/internal/control_test_engine/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/ngap_control/ue_context_management"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/security"
	"strings"
	"time"
)

func RegistrationUE(connN2 *sctp.SCTPConn, imsi string, ranUeId int64, conf config.Config, gnb *context.RanGnbContext, mcc, mnc string) (*context.RanUeContext, error) {

	// instance new ue.
	ue := &context.RanUeContext{}

	// new UE Context
	ue.NewRanUeContext(imsi, ranUeId, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2, conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, mcc, mnc, int32(conf.Ue.Snssai.Sd), conf.Ue.Snssai.Sst)

	log.Info("-------------------------------------")
	log.Info("UE INFORMATION:")
	log.Info("1-IMSI: ", ue.Supi)
	log.Info("2-OPc: ", conf.Ue.Opc)
	log.Info("2-Key: ", conf.Ue.Key)
	log.Info("3-Amf: ", conf.Ue.Amf)
	log.Info("-------------------------------------")

	// make registration request.
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, ue.Suci, nil, nil, ue)

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("UE[%d]", ranUeId),
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "REGISTRATION REQUEST",
	}).Info("Sending message")

	// make initial ue message.
	err := nas_transport.InitialUEMessage(connN2, registrationRequest, ue, gnb)
	if err != nil {
		log.Fatal("Error sending initial ue message: ", err)
	}

	// receive NAS Authentication Request Msg
	ngapMsg, err := nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if err != nil {
		log.Fatal("Error sending Downlink Nas transport: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      "AMF",
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "DOWNLINK NAS TRANSPORT",
	}).Info("Receiving message")

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": fmt.Sprintf("UE[%d]", ranUeId),
		"message":     "AUTHENTICATION REQUEST",
	}).Info("Sending message")

	// send NAS Authentication Response
	pdu, err := mm_5gs.AuthenticationResponse(ue, ngapMsg)
	if err != nil {
		log.Fatal("Error sending Authentication Response: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("UE[%d]", ranUeId),
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "AUTHENTICATION RESPONSE",
	}).Info("Sending message")

	// get UeAmfNgapId from DownlinkNasTransport message.
	ue.SetAmfNgapId(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport.ProtocolIEs.List[0].Value.AMFUENGAPID.Value)

	// send Nas Authentication response within UplinkNasTransport.
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if err != nil {
		log.Fatal("Error sending Uplink Nas transport: ", err)
	}

	// receive NAS Security Mode Command Msg
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if err != nil {
		log.Fatal("Error receive Downlink Nas transport: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      "AMF",
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "DOWNLINK NAS TRANSPORT",
	}).Info("Receiving message")

	// decode nas security mode complete here.
	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": fmt.Sprintf("UE[%d]", ranUeId),
		"message":     "SECURITY MODE COMMAND",
	}).Info("Sending message")

	// send NAS Security Mode Complete from UplinkNasTransport
	pdu, err = mm_5gs.SecurityModeComplete(ue)
	if err != nil {
		log.Fatal("Error sending Security Mode Complete: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("UE[%d]", ranUeId),
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "SECURITY MODE COMPLETE",
	}).Info("Sending message")

	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if err != nil {
		log.Fatal("Error receiving Uplink Nas transport: ", err)
	}

	// receive ngap Initial Context Setup Request Msg.
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if err != nil {
		log.Fatal("Error receive Initial Context Setup Request: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      "AMF",
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "INITIAL CONTEXT SETUP REQUEST",
	}).Info("Receiving message")

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": fmt.Sprintf("UE[%d]", ranUeId),
		"message":     "REGISTRATION ACCEPT",
	}).Info("Sending message")

	// send ngap Initial Context Setup Response Msg
	err = ue_context_management.InitialContextSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi)
	if err != nil {
		log.Fatal("Error sending Initial Context Setup Response: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": "AMF",
		"message":     "INITIAL CONTEXT SETUP RESPONSE",
	}).Info("Sending message")

	// send NAS Registration Complete Msg
	pdu, err = mm_5gs.RegistrationComplete(ue)
	if err != nil {
		log.Fatal("Error sending Registration Complete: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("UE[%d]", ranUeId),
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "REGISTRATION COMPLETE",
	}).Info("Sending message")

	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if err != nil {
		log.Fatal("Error receiving Uplink Nas transport: ", err)
	}

	// included configuration update command here.
	if strings.ToLower(conf.AMF.Name) == "open5gs" {
		_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
		if err != nil {
			log.Fatal("Error receiving Downlink Nas transport: ", err)
		}
		log.WithFields(log.Fields{
			"protocol":    "NGAP",
			"source":      "AMF",
			"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
			"message":     "DOWNLINK NAS TRANSPORT",
		}).Info("Receiving message")

		log.WithFields(log.Fields{
			"protocol":    "NAS",
			"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
			"destination": fmt.Sprintf("UE[%d]", ranUeId),
			"message":     "CONFIGURATION UPDATE COMMAND",
		}).Info("Sending message")
	}

	// send PduSessionEstablishmentRequest Msg
	pdu, err = mm_5gs.UlNasTransport(ue, uint8(ue.AmfUeNgapId), nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &ue.Snssai)
	if err != nil {
		log.Fatal("Error sending PDU Session request: ", err)
	}
	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("UE[%d]", ranUeId),
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "UL NAS TRANSPORT/PDU SESSSION ESTABLISHMENT REQUEST",
	}).Info("Sending message")

	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if err != nil {
		log.Fatal("Error sending Uplink Nas transport: ", err)
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	ngapMsg, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if err != nil {
		log.Fatal("Error receiving Downlink Nas transport: ", err)
	}
	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      "AMF",
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "PDU SESSION RESOURCE SETUP REQUEST",
	}).Info("Receiving message")

	log.WithFields(log.Fields{
		"protocol":    "NAS",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": fmt.Sprintf("UE[%d]", ranUeId),
		"message":     "DL NAS TRANSPORT/PDU SESSION ESTABLISHMENT ACCEPT",
	}).Info("Sending message")

	// decode IE Nas.
	nasPdu, err := sm_5gs.DecodeNasPduAccept(ngapMsg)
	if err != nil {
		return nil, err
	}

	// decode IE NGAP
	gtpTeid, err := pdu_session_management.GetGtpTeid(ngapMsg)
	if err != nil {
		return nil, err
	}

	// got ip address for ue.
	ue.SetIp(sm_5gs.GetPduAdress(nasPdu))

	// got gtp teid for ue.
	ue.SetUeTeid(gtpTeid[3])

	// send 14. NGAP-PDU Session Resource Setup Response.
	err = pdu_session_management.PDUSessionResourceSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi, conf.GNodeB.DataIF.Ip)
	if err != nil {
		log.Fatal("Error sending PDUSessionResourceSetupResponse: ", err)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": "AMF",
		"message":     "PDU SESSION RESOURCE SETUP RESPONSE",
	}).Info("Sending message")

	// time.Sleep(1 * time.Second)
	time.Sleep(100 * time.Millisecond)

	msg := fmt.Sprintf("[UE%d][%s] RECEIVE IP:%s AND UP-TEID:0x0000000%x DL-TEID:x0000000%x", ranUeId, imsi, ue.GetIp(), ue.GetUeTeid(), ue.AmfUeNgapId)
	log.Info(msg)
	log.Info("REGISTRATION FINISHED")

	// function worked fine.
	return ue, nil
}
