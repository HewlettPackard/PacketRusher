package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/nas_control"
	"my5G-RANTester/internal/control_test_engine/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ngap_control/nas_transport"
	"my5G-RANTester/internal/control_test_engine/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/ngap_control/ue_context_management"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/openapi/models"
	"time"
)

func RegistrationUE(connN2 *sctp.SCTPConn, imsi string, ranUeId int64, ranIpAddr string) (string, error) {
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)
	var n int

	// instance new ue.
	ue := &nas_control.RanUeContext{}

	// make initial UE message.
	err := nas_transport.InitialUEMessage(connN2, ue, imsi, ranUeId)
	if err != nil {
		fmt.Println(err)
	}

	/*
		n, err = connN2.Read(recvMsg)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error receiving %s ue nas authentication request message")
		}
		ngapMsg, err := ngap.Decoder(recvMsg[:n])
		if err != nil {
			return ue.Supi, fmt.Errorf("Error decoding %s ue nas authentication request message")

		}
	*/

	// receive NAS Authentication Request Msg
	ngapMsg, err := nas_transport.DownlinkNasTransport(connN2)
	if err != nil {
		fmt.Println(err)
	}

	/*

		// Calculate for RES*
		nasPdu := nas_control.GetNasPdu(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport)
		if nasPdu == nil {
			return ue.Supi, fmt.Errorf("Invalid NAS PDU")
		}

		rand := nasPdu.AuthenticationRequest.GetRANDValue()
		resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], "5G:mnc093.mcc208.3gppnetwork.org")

		// send NAS Authentication Response
		pdu := mm_5gs.GetAuthenticationResponse(resStat, "")
		sendMsg, err = nas_transport.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error getting %s NAS Authentication Response", ue.Supi)
		}
		_, err = connN2.Write(sendMsg)
			if err != nil {
				return ue.Supi, fmt.Errorf("Error sending %s NAS Authentication Response", ue.Supi)
			}
	*/

	// send NAS Authentication Response
	pdu, err := mm_5gs.AuthenticationResponse(ue, ngapMsg)
	if err != nil {
		fmt.Println(err)
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu)

	/*
		n, err = connN2.Read(recvMsg)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error reading %s NAS Security Mode Command Message", ue.Supi)
		}
		_, err = ngap.Decoder(recvMsg[:n])
		if err != nil {
			return ue.Supi, fmt.Errorf("Error decoding %s NAS Security Mode Command Message", ue.Supi)
		}
	*/

	// receive NAS Security Mode Command Msg
	_, err = nas_transport.DownlinkNasTransport(connN2)
	if err != nil {
		fmt.Println(err)
	}

	// send NAS Security Mode Complete Msg
	/*
		pdu = mm_5gs.GetSecurityModeComplete(registrationRequest)
		pdu, err = nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error encoding %s ue NAS Security Mode Complete Message", ue.Supi)
		}
	*/

	// send NAS Security Mode Complete Msg
	pdu, err = mm_5gs.SecurityModeComplete(ue)
	if err != nil {
		fmt.Println(err)
	}
	err = nas_transport.UplinkNasTransport(connN2, ue.AmfUeNgapId, ue.RanUeNgapId, pdu)

	/*
		sendMsg, err = nas_transport.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error getting %s ue NAS Security Mode Complete Message", ue.Supi)
		}
		_, err = connN2.Write(sendMsg)
		if err != nil {
			return ue.Supi, fmt.Errorf("Error sending %s ue NAS Security Mode Complete Message", ue.Supi)
		}
	*/

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
	sendMsg, err = ue_context_management.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
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
	sendMsg, err = nas_transport.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
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

	sendMsg, err = nas_transport.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
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
	sendMsg, err = pdu_session_management.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranIpAddr)
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
