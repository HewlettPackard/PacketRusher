package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control-test-engine/nas-control"
	"my5G-RANTester/internal/control-test-engine/nas-control/mm_5gs"
	"my5G-RANTester/internal/control-test-engine/ngap-control"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/openapi/models"
	"time"
)

func RegistrationUE(connN2 *sctp.SCTPConn, ueId int, ranUeId int64, ranIpAddr string) (string, error) {
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)
	var n int

	// instance new ue.
	ue := &nas_control.RanUeContext{}

	// new UE Context
	ue.NewRanUeContext(ueId, ranUeId, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2, "5122250214c33e723a5dd523fc145fc0", "981d464c7c52eb6e5036234984ad0bcf", "c9e8763286b5b9ffbdf56e1297d0887b", "8000")

	// ue.amfUENgap is received by AMF in authentication request.(? changed this).
	ue.AmfUeNgapId = ranUeId

	// send InitialUeMessage(Registration Request)(imsi-2089300007487)

	// generate suci for authentication.
	suciV2, suciV1, err := ue.EncodeUeSuci()
	if err != nil {
		return ue.Supi, fmt.Errorf("The test failed when SUCI was created! Error:%s", err)
	}

	mobileIdentity5GS := nasType.MobileIdentity5GS{
		Len:    12, // suci
		Buffer: []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, suciV1, suciV2},
	}
	ueSecurityCapability := nas_control.SetUESecurityCapability(ue)
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, nil, ueSecurityCapability)
	sendMsg, err = ngap_control.GetInitialUEMessage(ue.RanUeNgapId, registrationRequest, "")
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue initial message", ue.Supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue initial message", ue.Supi)
	}

	// receive NAS Authentication Request Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error receiving %s ue nas authentication request message")
	}
	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return ue.Supi, fmt.Errorf("Error decoding %s ue nas authentication request message")

	}

	// Calculate for RES*
	nasPdu := nas_control.GetNasPdu(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport)
	if nasPdu == nil {
		return ue.Supi, fmt.Errorf("Invalid NAS PDU")
	}

	rand := nasPdu.AuthenticationRequest.GetRANDValue()
	resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], "5G:mnc093.mcc208.3gppnetwork.org")

	// send NAS Authentication Response
	pdu := mm_5gs.GetAuthenticationResponse(resStat, "")
	sendMsg, err = ngap_control.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s NAS Authentication Response", ue.Supi)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s NAS Authentication Response", ue.Supi)
	}

	// receive NAS Security Mode Command Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error reading %s NAS Security Mode Command Message", ue.Supi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return ue.Supi, fmt.Errorf("Error decoding %s NAS Security Mode Command Message", ue.Supi)
	}

	// send NAS Security Mode Complete Msg
	pdu = mm_5gs.GetSecurityModeComplete(registrationRequest)
	pdu, err = nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error encoding %s ue NAS Security Mode Complete Message", ue.Supi)
	}
	sendMsg, err = ngap_control.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue NAS Security Mode Complete Message", ue.Supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue NAS Security Mode Complete Message", ue.Supi)
	}

	// receive ngap Initial Context Setup Request Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error receiving %s ue ngap Initial Context Setup Request Msg", ue.Supi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return ue.Supi, fmt.Errorf("Error decoding %s ue ngap Initial Context Setup Request Msg", ue.Supi)
	}

	// send ngap Initial Context Setup Response Msg
	sendMsg, err = ngap_control.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue ngap Initial Context Setup Response Msg", ue.Supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue ngap Initial Context Setup Response Msg", ue.Supi)
	}

	// send NAS Registration Complete Msg
	pdu = mm_5gs.GetRegistrationComplete(nil)
	pdu, err = nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error encoding %s ue NAS Registration Complete Msg", ue.Supi)
	}
	sendMsg, err = ngap_control.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue NAS Registration Complete Msg", ue.Supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue NAS Registration Complete Msg", ue.Supi)
	}

	time.Sleep(100 * time.Millisecond)

	// send GetPduSessionEstablishmentRequest Msg

	// called Single Network Slice Selection Assistance Information (S-NSSAI).
	sNssai := models.Snssai{
		Sst: 1, //The SST part of the S-NSSAI is mandatory and indicates the type of characteristics of the Network Slice.
		Sd:  "010203",
	}

	pdu = mm_5gs.GetUlNasTransport_PduSessionEstablishmentRequest(uint8(ranUeId), nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", (&sNssai))
	pdu, err = nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error encoding %s ue PduSession Establishment Request Msg", ue.Supi)
	}

	sendMsg, err = ngap_control.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue PduSession Establishment Request Msg", ue.Supi)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue PduSession Establishment Request Msg", ue.Supi)
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error reading %s ue NGAP-PDU Session Establishment Setup accept", ue.Supi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return ue.Supi, fmt.Errorf("Error decoding %s ue NGAP-PDU Session Establishment Setup accept", ue.Supi)
	}

	// send 14. NGAP-PDU Session Resource Setup Response.
	sendMsg, err = ngap_control.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranIpAddr)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error getting %s ue NGAP-PDU Session Resource Setup Response", ue.Supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return ue.Supi, fmt.Errorf("Error sending %s ue NGAP-PDU Session Resource Setup Response", ue.Supi)
	}

	// wait 1s
	// time.Sleep(1 * time.Second)
	time.Sleep(100 * time.Millisecond)

	// function worked fine.
	return ue.Supi, nil
}
