package multipleUes

import (
	"encoding/hex"
	"fmt"
	"free5gc/lib/CommonConsumerTestData/UDM/TestGenAuthData"
	"free5gc/lib/nas"
	"free5gc/lib/nas/nasMessage"
	"free5gc/lib/nas/nasTestpacket"
	"free5gc/lib/nas/nasType"
	"free5gc/lib/nas/security"
	"free5gc/lib/ngap"
	"free5gc/lib/openapi/models"
	"free5gc/src/test"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"strconv"
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

	// looping com a autenticação de vários ues.
	for i := 1; i <= 52; i++ {

		// criando vários imsi diferentes para autenticação de varios ues.
		var base string
		switch true {
		case i < 10:
			base = "imsi-208930000000"
		case i < 100:
			base = "imsi-20893000000"
		case i >= 100:
			base = "imsi-2089300000"
		}

		imsi := base + strconv.Itoa(i)
		fmt.Println(imsi)
		ueId := int64(i)

		ue := test.NewRanUeContext(imsi, ueId, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2)
		ue.AmfUeNgapId = ueId // mesmo valor que o ueId por isso o uso aqui!!!!
		ue.AuthenticationSubs = getAuthSubscription()
		// insert UE data to MongoDB(not implemented)

		// colocando os dados de varios imsi para autenticação de vários ues.
		var valor = []uint8{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80, 0x90, 0x01,
			0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0x02, 0x12, 0x22, 0x32, 0x42, 0x52,
			0x62, 0x72, 0x82, 0x92, 0x03, 0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0x04, 0x14,
			0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0x05, 0x15, 0x25, 0x35, 0x45, 0x55, 0x65,
			0x75, 0x85, 0x95, 0x06, 0x16, 0x26, 0x36, 0x46, 0x56, 0x66, 0x76, 0x86, 0x96, 0x07, 0x17, 0x27,
			0x37, 0x47, 0x57, 0x67, 0x77, 0x87, 0x97, 0x08, 0x18, 0x28, 0x38, 0x48, 0x58, 0x68, 0x78, 0x88,
			0x98, 0x09, 0x19, 0x29, 0x39, 0x49, 0x59, 0x69, 0x79, 0x89, 0x99, 0x00, 0x10, 0x20, 0x30, 0x40, 0x50,
			0x60, 0x70, 0x80, 0x90, 0x01, 0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0x02}

		// adicionando a lógica da próximo uint8.
		var valor2 uint8
		if i < 100 {
			valor2 = 0x00
		} else {
			valor2 = 0x10
		}

		// send InitialUeMessage(Registration Request)(imsi-2089300007487)
		mobileIdentity5GS := nasType.MobileIdentity5GS{
			Len:    12, // suci
			Buffer: []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, valor2, valor[i-1]},
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
		pdu = nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(10, nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &sNssai)
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

		// alterando o destino dos pings.
		basePing := "60.60.0."
		srcPing := basePing + strconv.Itoa(i)
		fmt.Println(srcPing)

		ipv4hdr := ipv4.Header{
			Version:  4,
			Len:      20,
			Protocol: 1,
			Flags:    0,
			TotalLen: 48,
			TTL:      64,
			Src:      net.ParseIP(srcPing).To4(),
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

	}

	conn.Close()
}
