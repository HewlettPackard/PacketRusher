package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/context"
	"my5G-RANTester/internal/control_test_engine/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/nas_control/sm_5gs"
	"my5G-RANTester/internal/control_test_engine/ngap_control/nas_transport"
	"my5G-RANTester/internal/control_test_engine/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/ngap_control/ue_context_management"
	"my5G-RANTester/internal/logging"
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

	// make initial UE message.
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, ue.Suci, nil, nil, ue)
	err := nas_transport.InitialUEMessage(connN2, registrationRequest, ue, gnb)
	msg := fmt.Sprintf("[NGAP/NAS][UE%d][%s]SEND INITIAL UE MESSAGE/REGISTRATION REQUEST", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// receive NAS Authentication Request Msg
	ngapMsg, err := nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]RECEIVE DOWNLINK NAS TRANSPORT/AUTHENTICATION REQUEST", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// send NAS Authentication Response
	pdu, err := mm_5gs.AuthenticationResponse(ue, ngapMsg)
	msg = fmt.Sprintf("[NAS][UE%d][%s]MAKE AUTHENTICATION RESPONSE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// get UeAmfNgapId from DownlinkNasTransport message.
	ue.SetAmfNgapId(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport.ProtocolIEs.List[0].Value.AMFUENGAPID.Value)

	// send Nas Authentication response within UplinkNasTransport.
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]SEND UPLINK NAS TRANSPORT/AUTHENTICATION RESPONSE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// receive NAS Security Mode Command Msg
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]RECEIVE DOWNLINK NAS TRANSPORT/SECURITY MODE COMMAND", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// send NAS Security Mode Complete from UplinkNasTransport
	pdu, err = mm_5gs.SecurityModeComplete(ue)
	msg = fmt.Sprintf("[NAS][UE%d][%s]MAKE SECURITY MODE COMPLETE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]SEND UPLINK NAS TRANSPORT/SECURITY MODE COMPLETE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// receive ngap Initial Context Setup Request Msg.
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]RECEIVE INITIAL CONTEXT SETUP REQUEST/REGISTRATION ACCEPT", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// send ngap Initial Context Setup Response Msg
	err = ue_context_management.InitialContextSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi)
	msg = fmt.Sprintf("[NGAP][UE%d][%s]SEND INITIAL CONTEXT SETUP RESPONSE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// send NAS Registration Complete Msg
	pdu, err = mm_5gs.RegistrationComplete(ue)
	msg = fmt.Sprintf("[NAS][UE%d][%s]MAKE REGISTRATION COMPLETE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]SEND UPLINK NAS TRANSPORT/REGISTRATION COMPLETE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// included configuration update command here.
	if strings.ToLower(conf.AMF.Name) == "open5gs" {
		_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
		msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]RECEIVE DOWNLINK NAS TRANSPORT/CONFIGURATION UPDATE COMMAND", ranUeId, imsi)
		if logging.Check_error(err, msg) {
			return nil, err
		}
	}

	// send PduSessionEstablishmentRequest Msg
	pdu, err = mm_5gs.UlNasTransport(ue, uint8(ue.AmfUeNgapId), nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &ue.Snssai)
	msg = fmt.Sprintf("[NAS][UE%d][%s]MAKE UL NAS TRANSPORT/PDU SESSION ESTABLISHMENT REQUEST", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]SEND UPLINK NAS TRANSPORT/UL NAS TRANSPORT/PDU SESSION ESTABLISHMENT REQUEST", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	ngapMsg, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	msg = fmt.Sprintf("[NGAP/NAS][UE%d][%s]RECEIVE PDU SESSION RESOURCE SETUP REQUEST/DL NAS TRANSPORT/PDU SESSION ESTABLISHMENT ACCEPT", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}
	nasPdu, err := sm_5gs.DecodeNasPduAccept(ngapMsg)
	if err != nil {
		return nil, err
	}
	//if logging.Check_error(err, "decode NasPduAccept worked fine!") {
	//return nil, err
	// }
	gtpTeid, err := pdu_session_management.GetGtpTeid(ngapMsg)
	if err != nil {
		return nil, err
	}
	//if logging.Check_error(err, "decode PDUSessionResourceSetupRequest worked fine!") {
	//return nil, err
	// }

	// got ip address for ue.
	ue.SetIp(sm_5gs.GetPduAdress(nasPdu))

	// got gtp teid for ue.
	ue.SetUeTeid(gtpTeid[3])

	// send 14. NGAP-PDU Session Resource Setup Response.
	err = pdu_session_management.PDUSessionResourceSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi, conf.GNodeB.DataIF.Ip)
	msg = fmt.Sprintf("[NGAP][UE%d][%s]SEND PDU SESSION RESOURCE SETUP RESPONSE", ranUeId, imsi)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// time.Sleep(1 * time.Second)
	time.Sleep(100 * time.Millisecond)

	msg = fmt.Sprintf("[UE%d][%s] RECEIVE IP:%s AND TEID:0x0000000%x", ranUeId, imsi, ue.GetIp(), ue.GetUeTeid())
	fmt.Println(msg)
	fmt.Println("REGISTRATION FINISHED")

	// function worked fine.
	return ue, nil
}
