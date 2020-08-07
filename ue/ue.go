package main

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"my5G-RANTester/lib/CommonConsumerTestData/UDM/TestGenAuthData"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasTestpacket"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/openapi/models"
	"my5G-RANTester/test"
	"net"
	"time"
)

const ranIpAddr string = "10.200.200.2"

func getAuthSubscription() (authSubs models.AuthenticationSubscription) {
	authSubs.PermanentKey = &models.PermanentKey{
		PermanentKeyValue: TestGenAuthData.MilenageTestSet19.K,
	}
	authSubs.Opc = &models.Opc{
		OpcValue: TestGenAuthData.MilenageTestSet19.OPC,
	}
	authSubs.Milenage = &models.Milenage{
		Op: &models.Op{
			OpValue: TestGenAuthData.MilenageTestSet19.OP,
		},
	}
	authSubs.AuthenticationManagementField = "8000"

	authSubs.SequenceNumber = TestGenAuthData.MilenageTestSet19.SQN
	authSubs.AuthenticationMethod = models.AuthMethod__5_G_AKA
	return
}

func setUESecurityCapability(ue *test.RanUeContext) (UESecurityCapability *nasType.UESecurityCapability) {
	UESecurityCapability = &nasType.UESecurityCapability{
		Iei:    nasMessage.RegistrationRequestUESecurityCapabilityType,
		Len:    8,
		Buffer: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	switch ue.CipheringAlg {
	case security.AlgCiphering128NEA0:
		UESecurityCapability.SetEA0_5G(1)
	case security.AlgCiphering128NEA1:
		UESecurityCapability.SetEA1_128_5G(1)
	case security.AlgCiphering128NEA2:
		UESecurityCapability.SetEA2_128_5G(1)
	case security.AlgCiphering128NEA3:
		UESecurityCapability.SetEA3_128_5G(1)
	}

	switch ue.IntegrityAlg {
	case security.AlgIntegrity128NIA0:
		UESecurityCapability.SetIA0_5G(1)
	case security.AlgIntegrity128NIA1:
		UESecurityCapability.SetIA1_128_5G(1)
	case security.AlgIntegrity128NIA2:
		UESecurityCapability.SetIA2_128_5G(1)
	case security.AlgIntegrity128NIA3:
		UESecurityCapability.SetIA3_128_5G(1)
	}

	return
}

func ipv4HeaderChecksum(hdr *ipv4.Header) uint32 {
	var Checksum uint32
	Checksum += uint32((hdr.Version<<4|(20>>2&0x0f))<<8 | hdr.TOS)
	Checksum += uint32(hdr.TotalLen)
	Checksum += uint32(hdr.ID)
	Checksum += uint32((hdr.FragOff & 0x1fff) | (int(hdr.Flags) << 13))
	Checksum += uint32((hdr.TTL << 8) | (hdr.Protocol))

	src := hdr.Src.To4()
	Checksum += uint32(src[0])<<8 | uint32(src[1])
	Checksum += uint32(src[2])<<8 | uint32(src[3])
	dst := hdr.Dst.To4()
	Checksum += uint32(dst[0])<<8 | uint32(dst[1])
	Checksum += uint32(dst[2])<<8 | uint32(dst[3])
	return ^(Checksum&0xffff0000>>16 + Checksum&0xffff)
}

func main() {
	var n int
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)

	// RAN connect to AMF
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		fmt.Println("não fez o socket do sctp!N2 Não criada!")
		return
	}

	// RAN connect to UPF
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		fmt.Println("não fez o tunel da upf!N3 Não criada")
		return
	}

	// send NGSetupRequest Msg
	sendMsg, err = test.GetNGSetupRequest([]byte("\x00\x01\x02"), 24, "free5gc")
	if err != nil {
		fmt.Println("preparando NGSetupRequest Msg")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send NGSetupRequest Msg")
		return
	}

	// receive NGSetupResponse Msg
	n, err = conn.Read(recvMsg)
	if err != nil {
		fmt.Println("recebido NGSetupResponse Msg")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decodificado NGSetupResponse Msg")
		return
	}

	// New UE
	// ue := test.NewRanUeContext("imsi-2089300007487", 1, security.AlgCiphering128NEA2, security.AlgIntegrity128NIA2)
	ue := test.NewRanUeContext("imsi-2089300007487", 1, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2)
	ue.AmfUeNgapId = 1
	ue.AuthenticationSubs = getAuthSubscription()
	// insert UE data to MongoDB(not implemented)

	// send InitialUeMessage(Registration Request)(imsi-2089300007487)
	mobileIdentity5GS := nasType.MobileIdentity5GS{
		Len:    12, // suci
		Buffer: []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, 0x47, 0x78},
	}

	ueSecurityCapability := setUESecurityCapability(ue)
	registrationRequest := nasTestpacket.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, nil, ueSecurityCapability)
	sendMsg, err = test.GetInitialUEMessage(ue.RanUeNgapId, registrationRequest, "")
	if err != nil {
		fmt.Println("preparando initial UE Msg")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("enviando initial UE Msg")
		return
	}

	// receive NAS Authentication Request Msg
	n, err = conn.Read(recvMsg)
	if err != nil {
		fmt.Println("receive NAS Authentication Request Msg")
		return
	}
	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder NAS Authentication Request Msg")
		return
	}

	// Calculate for RES*
	nasPdu := test.GetNasPdu(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport)
	if nasPdu == nil {
		fmt.Println("nasPdu é inválido!")
		return
	}
	rand := nasPdu.AuthenticationRequest.GetRANDValue()
	resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], "5G:mnc093.mcc208.3gppnetwork.org")

	// send NAS Authentication Response
	pdu := nasTestpacket.GetAuthenticationResponse(resStat, "")
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		fmt.Println("preparando NAS Authentication Response")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("enviando NAS Authentication Response")
		return
	}

	// receive NAS Security Mode Command Msg
	n, err = conn.Read(recvMsg)
	if err != nil {
		fmt.Println("receive NAS Security Mode Command Msg ---Downlink NAS Transport")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder NAS Security Mode Command Msg ---Downlink NAS Transport")
		return
	}

	// send NAS Security Mode Complete Msg
	pdu = nasTestpacket.GetSecurityModeComplete(registrationRequest)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		fmt.Println("encode NAS Security Mode Command Msg ---Uplink NAS Transport")
		return
	}
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		fmt.Println("encode e get up  NAS Security Mode Command Msg ---Uplink NAS Transport")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send NAS Security Mode Command Msg ---Uplink NAS Transport")
		return
	}

	// receive ngap Initial Context Setup Request Msg
	n, err = conn.Read(recvMsg)
	if err != nil {
		fmt.Println("receive ngap Initial Context Setup Request Msg")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder ngap Initial Context Setup Request Msg")
		return
	}

	// send ngap Initial Context Setup Response Msg
	sendMsg, err = test.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
	if err != nil {
		fmt.Println("get ngap Initial Context Setup Response Msg")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send ngap Initial Context Setup Response Msg")
		return
	}

	// send NAS Registration Complete Msg
	pdu = nasTestpacket.GetRegistrationComplete(nil)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		fmt.Println("encode NAS Registration Complete Msg")
		return
	}
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		fmt.Println("get NAS Registration Complete Msg --- UPLINK NAS TRANSPORT")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send NAS Registration Complete Msg --- UPLINK NAS TRANSPORT")
		return
	}

	time.Sleep(100 * time.Millisecond)
	// send GetPduSessionEstablishmentRequest Msg

	// possível fonte de problema!!!!!!!!!
	sNssai := models.Snssai{
		Sst: 1,
		Sd:  "010203",
	}
	pdu = nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(10, nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", (&sNssai))

	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		fmt.Println("encode GetPduSessionEstablishmentRequest Msg")
		return
	}
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		fmt.Println("get GetPduSessionEstablishmentRequest Msg")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send GetPduSessionEstablishmentRequest Msg")
		return
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	n, err = conn.Read(recvMsg)
	if err != nil {
		fmt.Println("receive NGAP-PDU Session Resource Setup Request")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder NGAP-PDU Session Resource Setup Request")
		return
	}

	// send 14. NGAP-PDU Session Resource Setup Response
	sendMsg, err = test.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranIpAddr)
	if err != nil {
		fmt.Println("get NGAP-PDU Session Resource Setup Response")
		return
	}
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("send NGAP-PDU Session Resource Setup Response")
		return
	}

	// wait 1s
	time.Sleep(1 * time.Second)

	// Send the dummy packet
	// ping IP(tunnel IP) from 60.60.0.2(127.0.0.1) to 60.60.0.20(127.0.0.8)
	gtpHdr, err := hex.DecodeString("32ff00340000000100000000")
	if err != nil {
		return
	}
	icmpData, err := hex.DecodeString("8c870d0000000000101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")
	if err != nil {
		return
	}

	ipv4hdr := ipv4.Header{
		Version:  4,
		Len:      20,
		Protocol: 1,
		Flags:    0,
		TotalLen: 48,
		TTL:      64,
		Src:      net.ParseIP("60.60.0.1").To4(),
		Dst:      net.ParseIP("60.60.0.101").To4(),
		ID:       1,
	}
	checksum := ipv4HeaderChecksum(&ipv4hdr)
	ipv4hdr.Checksum = int(checksum)

	v4HdrBuf, err := ipv4hdr.Marshal()
	if err != nil {
		return
	}
	tt := append(gtpHdr, v4HdrBuf...)
	if err != nil {
		return
	}

	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: 12394, Seq: 1,
			Data: icmpData,
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return
	}
	b[2] = 0xaf
	b[3] = 0x88
	_, err = upfConn.Write(append(tt, b...))
	if err != nil {
		return
	}
	time.Sleep(1 * time.Second)

	conn.Close()
	fmt.Println("Deu tudo certo !!!")
}
