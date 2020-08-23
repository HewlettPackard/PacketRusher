package main

import (
	"encoding/hex"
	"fmt"
	"git.cs.nctu.edu.tw/calee/sctp"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"my5G-RANTester/lib/CommonConsumerTestData/UDM/TestGenAuthData"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasTestpacket"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapSctp"
	"my5G-RANTester/lib/openapi/models"
	"my5G-RANTester/test"
	"net"
	"time"
)

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

func connectToAmf(amfIP, ranIP string, amfPort, ranPort int) (*sctp.SCTPConn, error) {
	amfAddr, ranAddr, err := getNgapIp(amfIP, ranIP, amfPort, ranPort)
	if err != nil {
		return nil, err
	}
	conn, err := sctp.DialSCTP("sctp", ranAddr, amfAddr)
	if err != nil {
		return nil, err
	}
	info, _ := conn.GetDefaultSentParam()
	info.PPID = ngapSctp.NGAP_PPID
	err = conn.SetDefaultSentParam(info)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getNgapIp(amfIP, ranIP string, amfPort, ranPort int) (amfAddr, ranAddr *sctp.SCTPAddr, err error) {
	ips := []net.IPAddr{}
	// se der um erro != nill entra no if.
	if ip, err1 := net.ResolveIPAddr("ip", amfIP); err1 != nil {
		err = fmt.Errorf("Error resolving address '%s': %v", amfIP, err1)
		return
	} else {
		ips = append(ips, *ip)
	}
	amfAddr = &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    amfPort,
	}
	ips = []net.IPAddr{}
	if ip, err1 := net.ResolveIPAddr("ip", ranIP); err1 != nil {
		err = fmt.Errorf("Error resolving address '%s': %v", ranIP, err1)
		return
	} else {
		ips = append(ips, *ip)
	}
	ranAddr = &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    ranPort,
	}
	return
}

func connectToUpf(enbIP, upfIP string, gnbPort, upfPort int) (*net.UDPConn, error) {
	upfAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", upfIP, upfPort))
	if err != nil {
		return nil, err
	}
	gnbAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", enbIP, gnbPort))
	if err != nil {
		return nil, err
	}
	return net.DialUDP("udp", gnbAddr, upfAddr)
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

// registration testing code.

// registration and authentication to a GNB
func registrationGNB(connN2 *sctp.SCTPConn, gnbId []byte, nameGNB string) error {
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)
	var n int

	// authentication and authorization for GNB.

	// send NGSetupRequest Msg
	// sendMsg, err := test.GetNGSetupRequest([]byte("\x00\x01\x02"), 24, "free5gc")
	sendMsg, err := test.GetNGSetupRequest(gnbId, 24, nameGNB)
	if err != nil {
		fmt.Println("get NGSetupRequest Msg")
		return fmt.Errorf("Error getting GNB %s NGSetup Request Msg", nameGNB)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		fmt.Println("send NGSetupRequest Msg")
		return fmt.Errorf("Error sending GNB %s NGSetup Request Msg", nameGNB)
	}

	// receive NGSetupResponse Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		fmt.Println("read NGSetupResponse Msg")
		return fmt.Errorf("Error reading GNB %s NGSetup Response Msg", nameGNB)
	}

	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder NGSetupResponse Msg")
		return fmt.Errorf("Error decoding GNB %s NGSetup Response Msg", nameGNB)
	}

	// function worked fine.
	return nil
}

// registration and authentication to a UE.
func registrationUE(connN2 *sctp.SCTPConn, imsiSupi string, ranUeId int64, suciV1 uint8, suciV2 uint8, ranIpAddr string) error {
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)
	var n int

	// new UE Context
	ue := test.NewRanUeContext(imsiSupi, ranUeId, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2)
	ue.AmfUeNgapId = ranUeId
	ue.AuthenticationSubs = getAuthSubscription()

	// send InitialUeMessage(Registration Request)(imsi-2089300007487)
	mobileIdentity5GS := nasType.MobileIdentity5GS{
		Len:    12, // suci
		Buffer: []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, suciV1, suciV2},
	}

	ueSecurityCapability := setUESecurityCapability(ue)
	registrationRequest := nasTestpacket.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, nil, ueSecurityCapability)
	sendMsg, err := test.GetInitialUEMessage(ue.RanUeNgapId, registrationRequest, "")
	if err != nil {
		return fmt.Errorf("Error getting %s ue initial message", imsiSupi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue initial message", imsiSupi)
	}

	// receive NAS Authentication Request Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return fmt.Errorf("Error receiving %s ue nas authentication request message")
	}
	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return fmt.Errorf("Error decoding %s ue nas authentication request message")

	}

	// Calculate for RES*
	nasPdu := test.GetNasPdu(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport)
	if nasPdu == nil {
		return fmt.Errorf("Invalid NAS PDU")
	}

	rand := nasPdu.AuthenticationRequest.GetRANDValue()
	resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], "5G:mnc093.mcc208.3gppnetwork.org")

	// send NAS Authentication Response
	pdu := nasTestpacket.GetAuthenticationResponse(resStat, "")
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return fmt.Errorf("Error getting %s NAS Authentication Response", imsiSupi)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s NAS Authentication Response", imsiSupi)
	}

	// receive NAS Security Mode Command Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return fmt.Errorf("Error reading %s NAS Security Mode Command Message", imsiSupi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return fmt.Errorf("Error decoding %s NAS Security Mode Command Message", imsiSupi)
	}

	// send NAS Security Mode Complete Msg
	pdu = nasTestpacket.GetSecurityModeComplete(registrationRequest)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		return fmt.Errorf("Error encoding %s ue NAS Security Mode Complete Message", imsiSupi)
	}
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return fmt.Errorf("Error getting %s ue NAS Security Mode Complete Message", imsiSupi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue NAS Security Mode Complete Message", imsiSupi)
	}

	// receive ngap Initial Context Setup Request Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return fmt.Errorf("Error receiving %s ue ngap Initial Context Setup Request Msg", imsiSupi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return fmt.Errorf("Error decoding %s ue ngap Initial Context Setup Request Msg", imsiSupi)
	}

	// send ngap Initial Context Setup Response Msg
	sendMsg, err = test.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
	if err != nil {
		return fmt.Errorf("Error getting %s ue ngap Initial Context Setup Response Msg", imsiSupi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue ngap Initial Context Setup Response Msg", imsiSupi)
	}

	// send NAS Registration Complete Msg
	pdu = nasTestpacket.GetRegistrationComplete(nil)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return fmt.Errorf("Error encoding %s ue NAS Registration Complete Msg", imsiSupi)
	}
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return fmt.Errorf("Error getting %s ue NAS Registration Complete Msg", imsiSupi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue NAS Registration Complete Msg", imsiSupi)
	}

	time.Sleep(100 * time.Millisecond)

	// send GetPduSessionEstablishmentRequest Msg

	// called Single Network Slice Selection Assistance Information (S-NSSAI).
	sNssai := models.Snssai{
		Sst: 1, //The SST part of the S-NSSAI is mandatory and indicates the type of characteristics of the Network Slice.
		Sd:  "010203",
	}

	pdu = nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(uint8(ranUeId), nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", (&sNssai))
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return fmt.Errorf("Error encoding %s ue PduSession Establishment Request Msg", imsiSupi)
	}

	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		return fmt.Errorf("Error getting %s ue PduSession Establishment Request Msg", imsiSupi)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue PduSession Establishment Request Msg", imsiSupi)
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	n, err = connN2.Read(recvMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue NGAP-PDU Session Establishment Setup accept", imsiSupi)
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		return fmt.Errorf("Error decoding %s ue NGAP-PDU Session Establishment Setup accept", imsiSupi)
	}

	// send 14. NGAP-PDU Session Resource Setup Response.
	sendMsg, err = test.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranIpAddr)
	if err != nil {
		return fmt.Errorf("Error getting %s ue NGAP-PDU Session Resource Setup Response", imsiSupi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue NGAP-PDU Session Resource Setup Response", imsiSupi)
	}

	// wait 1s
	time.Sleep(1 * time.Second)

	// function worked fine.
	return nil
}

// UE ping.
func pingUE(connN3 *net.UDPConn, ranUeId int, ipUe string) error {
	var valorGtp int
	var auxGtp string

	// generates some GTP-TEIDs for the RAN-UPF tunnels(uplink) in order to make the GTP header.
	valorGtp = 1 + 2*(ranUeId-1)
	if valorGtp < 16 {
		auxGtp = "32ff00340000000" + fmt.Sprintf("%x", valorGtp) + "00000000"
	} else if valorGtp < 256 {
		auxGtp = "32ff0034000000" + fmt.Sprintf("%x", valorGtp) + "00000000"
	} else {
		auxGtp = "32ff003400000" + fmt.Sprintf("%x", valorGtp) + "00000000"
	}

	gtpHdr, err := hex.DecodeString(auxGtp)
	if err != nil {
		return fmt.Errorf("Error getting gtp header")
	}
	icmpData, err := hex.DecodeString("8c870d0000000000101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")
	if err != nil {
		return fmt.Errorf("Error getting icmp data")
	}

	// included ping source.
	basePing := ipUe
	fmt.Println(basePing)

	ipv4hdr := ipv4.Header{
		Version:  4,
		Len:      20,
		Protocol: 1,
		Flags:    0,
		TotalLen: 48,
		TTL:      64,
		Src:      net.ParseIP(basePing).To4(),
		Dst:      net.ParseIP("60.60.0.101").To4(),
		ID:       1,
	}
	checksum := ipv4HeaderChecksum(&ipv4hdr)
	ipv4hdr.Checksum = int(checksum)

	v4HdrBuf, err := ipv4hdr.Marshal()
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}
	tt := append(gtpHdr, v4HdrBuf...)
	if err != nil {
		return fmt.Errorf("Error in ping!")
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
		return fmt.Errorf("Error in ping!")
	}
	b[2] = 0xaf
	b[3] = 0x88

	_, err = connN3.Write(append(tt, b...))
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}
	time.Sleep(1 * time.Second)

	// function worked fine.
	return nil
}

// registration and authentication to a UE.
func testAttachUe() error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a UE.
	err = registrationUE(conn, "imsi-2089300000001", 1, 0x00, 0x10, ranIpAddr)
	if err != nil {
		return fmt.Errorf("The test failed when UE tried to attach! Error:%s", err)
	}

	// using the data plane UE
	err = pingUE(upfConn, 1, "60.60.0.1")
	if err != nil {
		return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	return nil
}
