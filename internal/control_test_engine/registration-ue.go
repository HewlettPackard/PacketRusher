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
	"time"
)

func RegistrationUE(connN2 *sctp.SCTPConn, imsi string, ranUeId int64, conf config.Config, gnb *context.RanGnbContext, mcc, mnc string) (string, error, string) {

	// instance new ue.
	ue := &context.RanUeContext{}

	// new UE Context
	ue.NewRanUeContext(imsi, ranUeId, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2, conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, mcc, mnc, int32(conf.Ue.Snssai.Sd), conf.Ue.Snssai.Sst)

	// make initial UE message.
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, ue.Suci, nil, nil, ue)
	err := nas_transport.InitialUEMessage(connN2, registrationRequest, ue, gnb)
	if logging.Check_error(err, "send Initial Ue Message") {
		return ue.Supi, err, ""
	}

	// receive NAS Authentication Request Msg
	ngapMsg, err := nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if logging.Check_error(err, "receive DownlinkNasTransport/authentication request") {
		return ue.Supi, err, ""
	}

	// send NAS Authentication Response
	pdu, err := mm_5gs.AuthenticationResponse(ue, ngapMsg)
	if logging.Check_error(err, "Authentication response worked fine") {
		return ue.Supi, err, ""
	}

	// get UeAmfNgapId from DownlinkNasTransport message.
	ue.SetAmfNgapId(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport.ProtocolIEs.List[0].Value.AMFUENGAPID.Value)

	// send Nas Authentication response within UplinkNasTransport.
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if logging.Check_error(err, "send UplinkNasTransport/Authentication Response") {
		return ue.Supi, err, ""
	}

	// receive NAS Security Mode Command Msg
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if logging.Check_error(err, "receive DownlinkNasTransport/Security Mode Command") {
		return ue.Supi, err, ""
	}

	// send NAS Security Mode Complete from UplinkNasTransport
	pdu, err = mm_5gs.SecurityModeComplete(ue)
	if logging.Check_error(err, "Security Mode Complete worked fine!") {
		return ue.Supi, err, ""
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if logging.Check_error(err, "send UplinkNasTransport/Security Mode Complete Msg!") {
		return ue.Supi, err, ""
	}

	// receive ngap Initial Context Setup Request Msg.
	_, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if logging.Check_error(err, "receive NGAP/Initial Context Setup Request") {
		return ue.Supi, err, ""
	}

	// send ngap Initial Context Setup Response Msg
	err = ue_context_management.InitialContextSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi)
	if logging.Check_error(err, "send NGAP/Initial context setup response message") {
		return ue.Supi, err, ""
	}

	// send NAS Registration Complete Msg
	pdu, err = mm_5gs.RegistrationComplete(ue)
	if logging.Check_error(err, "NAS registration complete worked fine") {
		return ue.Supi, err, ""
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if logging.Check_error(err, "send UplinkNasTransport/registration complete") {
		return ue.Supi, err, ""
	}

	// included configuration update command here.
	confUpdate, err := nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if logging.Check_error(err, "") {
		return ue.Supi, err, ""
	}
	if logging.Check_Ngap(confUpdate, "receive DownlinkNasTransport/ConfigurationUpdateCommand") {
		fmt.Println("does not receive receive DownlinkNasTransport/ConfigurationUpdateCommand")
	}
	//time.Sleep(100 * time.Millisecond)

	// send PduSessionEstablishmentRequest Msg
	pdu, err = mm_5gs.UlNasTransport(ue, uint8(ue.AmfUeNgapId), nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &ue.Snssai)
	if logging.Check_error(err, "NAS UlNasTransport worked fine!") {
		return ue.Supi, err, ""
	}

	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu, gnb)
	if logging.Check_error(err, "send UplinkNasTransport/Ul Nas Transport/PduSession Establishment request") {
		return ue.Supi, err, ""
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	ngapMsg, err = nas_transport.DownlinkNasTransport(connN2, ue.Supi)
	if logging.Check_error(err, "receive PDU Session Resource Setup Request/Dl Nas Transport/PDU establishment accept") {
		return ue.Supi, err, ""
	}
	nasPdu, err := sm_5gs.DecodeNasPduAccept(ue, ngapMsg)
	if logging.Check_error(err, "decodeNasPduAccept worked fine!") {
		return ue.Supi, err, ""
	}

	// got ip address for ue.
	ue.SetIp(sm_5gs.GetPduAdress(nasPdu))

	// send 14. NGAP-PDU Session Resource Setup Response.
	err = pdu_session_management.PDUSessionResourceSetupResponse(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, ue.Supi, conf.GNodeB.DataIF.Ip)
	if logging.Check_error(err, "send PDU Session Resource Setup Response") {
		return ue.Supi, err, ""
	}

	// time.Sleep(1 * time.Second)
	time.Sleep(100 * time.Millisecond)

	// function worked fine.
	return ue.Supi, nil, ue.GetIp()
}
