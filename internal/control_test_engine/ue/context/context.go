package context

import (
	"bytes"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/lib/UeauCommon"
	"my5G-RANTester/lib/milenage"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/openapi/models"
	"net"
	"reflect"
	"regexp"
)

// 5GMM main states in the UE.
const MM5G_NULL = 0x00
const MM5G_DEREGISTERED = 0x01
const MM5G_REGISTERED_INITIATED = 0x02
const MM5G_REGISTERED = 0x03
const MM5G_SERVICE_REQ_INIT = 0x04
const MM5G_DEREGISTERED_INIT = 0x05

// 5GSM main states in the UE.
const SM5G_PDU_SESSION_INACTIVE = 0x06
const SM5G_PDU_SESSION_ACTIVE_PENDING = 0x07
const SM5G_PDU_SESSION_ACTIVE = 0x08

type UEContext struct {
	id         uint8
	UeSecurity SECURITY
	StateMM    int
	StateSM    int
	UnixConn   net.Conn
	PduSession PDUSession
	amfInfo    Amf
}

type Amf struct {
	amfRegionId uint8
	amfSetId    uint16
	amfPointer  uint8
}

type PDUSession struct {
	Id        uint8
	ueIP      string
	ueGnbIP   net.IP
	Dnn       string
	Snssai    models.Snssai
	gatewayIP net.IP
	tun       netlink.Link
	routeTun  *netlink.Route
	ruleTun   *netlink.Rule
}

type SECURITY struct {
	Supi               string
	Msin               string
	mcc                string
	mnc                string
	ULCount            security.Count
	DLCount            security.Count
	CipheringAlg       uint8
	IntegrityAlg       uint8
	Snn                string
	KnasEnc            [16]uint8
	KnasInt            [16]uint8
	Kamf               []uint8
	AuthenticationSubs models.AuthenticationSubscription
	Suci               nasType.MobileIdentity5GS
	Guti               [4]byte
}

func (ue *UEContext) NewRanUeContext(msin string,
	cipheringAlg, integrityAlg uint8,
	k, opc, op, amf, sqn, mcc, mnc, dnn string,
	sst int32, sd string, id uint8) {

	// added SUPI.
	ue.UeSecurity.Msin = msin

	// added ciphering algorithm.
	ue.UeSecurity.CipheringAlg = cipheringAlg

	// added integrity algorithm.
	ue.UeSecurity.IntegrityAlg = integrityAlg

	// added key, AuthenticationManagementField and opc or op.
	ue.SetAuthSubscription(k, opc, op, amf, sqn)

	// added suci
	suciV1, suciV2, suciV3, suciV4, suciV5 := ue.EncodeUeSuci()

	// added mcc and mnc
	ue.UeSecurity.mcc = mcc
	ue.UeSecurity.mnc = mnc

	// added supi
	ue.UeSecurity.Supi = fmt.Sprintf("imsi-%s%s%s", mcc, mnc, msin)

	// added PDU Session id
	ue.PduSession.Id = id

	// added UE id.
	ue.id = id

	// added network slice
	ue.PduSession.Snssai.Sd = sd
	ue.PduSession.Snssai.Sst = sst

	// added Domain Network Name.
	ue.PduSession.Dnn = dnn

	// added gateway ip.
	ue.PduSession.gatewayIP = net.ParseIP("127.0.0.2").To4()

	// encode mcc and mnc for mobileIdentity5Gs.
	resu := ue.GetMccAndMncInOctets()

	// added suci to mobileIdentity5GS
	if len(ue.UeSecurity.Msin) == 8 {
		ue.UeSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    12,
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, suciV4, suciV3, suciV2, suciV1},
		}
	} else if len(ue.UeSecurity.Msin) == 10 {
		ue.UeSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    13,
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, suciV5, suciV4, suciV3, suciV2, suciV1},
		}
	}

	// added snn.
	ue.UeSecurity.Snn = ue.deriveSNN()

	// added initial state for MM(NULL)
	ue.SetStateMM_NULL()

	// added initial state for SM(INACTIVE)
	ue.SetStateSM_PDU_SESSION_INACTIVE()

}

func (ue *UEContext) GetUeId() uint8 {
	return ue.id
}

func (ue *UEContext) GetSuci() nasType.MobileIdentity5GS {
	return ue.UeSecurity.Suci
}

func (ue *UEContext) GetMsin() string {
	return ue.UeSecurity.Msin
}

func (ue *UEContext) GetSupi() string {
	return ue.UeSecurity.Supi
}

func (ue *UEContext) SetStateSM_PDU_SESSION_INACTIVE() {
	ue.StateSM = SM5G_PDU_SESSION_INACTIVE
}

func (ue *UEContext) SetStateSM_PDU_SESSION_ACTIVE() {
	ue.StateSM = SM5G_PDU_SESSION_ACTIVE
}

func (ue *UEContext) SetStateSM_PDU_SESSION_PENDING() {
	ue.StateSM = SM5G_PDU_SESSION_ACTIVE_PENDING
}

func (ue *UEContext) SetStateMM_DEREGISTERED_INITIATED() {
	ue.StateMM = MM5G_DEREGISTERED_INIT
}

func (ue *UEContext) SetStateMM_MM5G_SERVICE_REQ_INIT() {
	ue.StateMM = MM5G_SERVICE_REQ_INIT
}

func (ue *UEContext) SetStateMM_REGISTERED_INITIATED() {
	ue.StateMM = MM5G_REGISTERED_INITIATED
}

func (ue *UEContext) SetStateMM_REGISTERED() {
	ue.StateMM = MM5G_REGISTERED
}

func (ue *UEContext) SetStateMM_NULL() {
	ue.StateMM = MM5G_NULL
}

func (ue *UEContext) SetStateMM_DEREGISTERED() {
	ue.StateMM = MM5G_DEREGISTERED
}

func (ue *UEContext) GetStateSM() int {
	return ue.StateSM
}

func (ue *UEContext) GetStateMM() int {
	return ue.StateMM
}

func (ue *UEContext) SetUnixConn(conn net.Conn) {
	ue.UnixConn = conn
}

func (ue *UEContext) GetUnixConn() net.Conn {
	return ue.UnixConn
}

func (ue *UEContext) SetIp(ip [12]uint8) {
	ue.PduSession.ueIP = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ue *UEContext) GetIp() string {
	return ue.PduSession.ueIP
}

func (ue *UEContext) GetGatewayIp() net.IP {
	return ue.PduSession.gatewayIP
}

func (ue *UEContext) SetGnbIp(ip net.IP) {
	ue.PduSession.ueGnbIP = ip
}

func (ue *UEContext) GetGnbIp() net.IP {
	return ue.PduSession.ueGnbIP
}

func (ue *UEContext) GetPduSesssionId() uint8 {
	return ue.PduSession.Id
}

func (ue *UEContext) deriveSNN() string {
	// 5G:mnc093.mcc208.3gppnetwork.org
	var resu string
	if len(ue.UeSecurity.mnc) == 2 {
		resu = "5G:mnc0" + ue.UeSecurity.mnc + ".mcc" + ue.UeSecurity.mcc + ".3gppnetwork.org"
	} else {
		resu = "5G:mnc" + ue.UeSecurity.mnc + ".mcc" + ue.UeSecurity.mcc + ".3gppnetwork.org"
	}

	return resu
}

func (ue *UEContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(ue.UeSecurity.mcc)
	mnc := reverse(ue.UeSecurity.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(ue.UeSecurity.mnc) == 2 {
		oct6 = "f" + string(mcc[0])
		oct7 = mnc
	} else {
		oct6 = string(mnc[0]) + string(mcc[0])
		oct7 = mnc[1:3]
	}

	// changed for bytes.
	resu, err := hex.DecodeString(oct5 + oct6 + oct7)
	if err != nil {
		fmt.Println(err)
	}

	return resu
}

func (ue *UEContext) EncodeUeSuci() (uint8, uint8, uint8, uint8, uint8) {

	// reverse imsi string.
	aux := reverse(ue.UeSecurity.Msin)

	// calculate decimal value.
	suci, error := hex.DecodeString(aux)
	if error != nil {
		return 0, 0, 0, 0, 0
	}

	// return decimal value
	// Function worked fine.
	if len(ue.UeSecurity.Msin) == 8 {
		return uint8(suci[0]), uint8(suci[1]), uint8(suci[2]), uint8(suci[3]), 0
	} else {
		return uint8(suci[0]), uint8(suci[1]), uint8(suci[2]), uint8(suci[3]), uint8(suci[4])
	}
}

func (ue *UEContext) SetAmfRegionId(amfRegionId uint8) {
	ue.amfInfo.amfRegionId = amfRegionId
}

func (ue *UEContext) GetAmfRegionId() uint8 {
	return ue.amfInfo.amfRegionId
}

func (ue *UEContext) SetAmfPointer(amfPointer uint8) {
	ue.amfInfo.amfPointer = amfPointer
}

func (ue *UEContext) GetAmfPointer() uint8 {
	return ue.amfInfo.amfPointer
}

func (ue *UEContext) SetAmfSetId(amfSetId uint16) {
	ue.amfInfo.amfSetId = amfSetId
}

func (ue *UEContext) GetAmfSetId() uint16 {
	return ue.amfInfo.amfSetId
}

func (ue *UEContext) Get5gGuti() [4]uint8 {
	return ue.UeSecurity.Guti
}

func (ue *UEContext) Set5gGuti(guti [4]uint8) {
	ue.UeSecurity.Guti = guti
}

func (ue *UEContext) deriveAUTN(autn []byte, ak []uint8) ([]byte, []byte, []byte) {

	sqn := make([]byte, 6)

	// get SQNxorAK
	SQNxorAK := autn[0:6]
	amf := autn[6:8]
	mac_a := autn[8:]

	// get SQN
	for i := 0; i < len(SQNxorAK); i++ {
		sqn[i] = SQNxorAK[i] ^ ak[i]
	}

	// return SQN, amf, mac_a
	return sqn, amf, mac_a
}

func (ue *UEContext) DeriveRESstarAndSetKey(authSubs models.AuthenticationSubscription,
	RAND []byte,
	snNmae string,
	AUTN []byte) ([]byte, string) {

	// parameters for authentication challenge.
	mac_a, mac_s := make([]byte, 8), make([]byte, 8)
	CK, IK := make([]byte, 16), make([]byte, 16)
	RES := make([]byte, 8)
	AK, AKstar := make([]byte, 6), make([]byte, 6)

	// Get OPC, K, SQN, AMF from USIM.
	OPC, _ := hex.DecodeString(authSubs.Opc.OpcValue)
	K, _ := hex.DecodeString(authSubs.PermanentKey.PermanentKeyValue)
	sqnUe, _ := hex.DecodeString(authSubs.SequenceNumber)
	AMF, _ := hex.DecodeString(authSubs.AuthenticationManagementField)

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// Get SQN, MAC_A, AMF from AUTN
	sqnHn, _, mac_aHn := ue.deriveAUTN(AUTN, AK)

	// Generate MAC_A, MAC_S
	milenage.F1_Test(OPC, K, RAND, sqnHn, AMF, mac_a, mac_s)

	// MAC verification.
	if !reflect.DeepEqual(mac_a, mac_aHn) {
		return nil, "MAC failure"
	}

	// Verification of sequence number freshness.
	if bytes.Compare(sqnUe, sqnHn) > 0 {

		// get AK*
		milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

		// From the standard, AMF(0x0000) should be used in the synch failure.
		amfSynch, _ := hex.DecodeString("0000")

		// get mac_s using sqn ue.
		milenage.F1_Test(OPC, K, RAND, sqnUe, amfSynch, mac_a, mac_s)

		sqnUeXorAK := make([]byte, 6)
		for i := 0; i < len(sqnUe); i++ {
			sqnUeXorAK[i] = sqnUe[i] ^ AKstar[i]
		}

		failureParam := append(sqnUeXorAK, mac_s...)

		return failureParam, "SQN failure"
	}

	// updated sqn value.
	authSubs.SequenceNumber = fmt.Sprintf("%x", sqnHn)

	// derive RES*
	key := append(CK, IK...)
	FC := UeauCommon.FC_FOR_RES_STAR_XRES_STAR_DERIVATION
	P0 := []byte(snNmae)
	P1 := RAND
	P2 := RES

	ue.DerivateKamf(key, snNmae, sqnHn, AK)
	ue.DerivateAlgKey()
	kdfVal_for_resStar := UeauCommon.GetKDFValue(key, FC, P0, UeauCommon.KDFLen(P0), P1, UeauCommon.KDFLen(P1), P2, UeauCommon.KDFLen(P2))
	return kdfVal_for_resStar[len(kdfVal_for_resStar)/2:], "successful"
}

func (ue *UEContext) DerivateKamf(key []byte, snName string, SQN, AK []byte) {

	FC := UeauCommon.FC_FOR_KAUSF_DERIVATION
	P0 := []byte(snName)
	SQNxorAK := make([]byte, 6)
	for i := 0; i < len(SQN); i++ {
		SQNxorAK[i] = SQN[i] ^ AK[i]
	}
	P1 := SQNxorAK
	Kausf := UeauCommon.GetKDFValue(key, FC, P0, UeauCommon.KDFLen(P0), P1, UeauCommon.KDFLen(P1))
	P0 = []byte(snName)
	Kseaf := UeauCommon.GetKDFValue(Kausf, UeauCommon.FC_FOR_KSEAF_DERIVATION, P0, UeauCommon.KDFLen(P0))

	supiRegexp, _ := regexp.Compile("(?:imsi|supi)-([0-9]{5,15})")
	groups := supiRegexp.FindStringSubmatch(ue.UeSecurity.Supi)

	P0 = []byte(groups[1])
	L0 := UeauCommon.KDFLen(P0)
	P1 = []byte{0x00, 0x00}
	L1 := UeauCommon.KDFLen(P1)

	ue.UeSecurity.Kamf = UeauCommon.GetKDFValue(Kseaf, UeauCommon.FC_FOR_KAMF_DERIVATION, P0, L0, P1, L1)
}

// Algorithm key Derivation function defined in TS 33.501 Annex A.9
func (ue *UEContext) DerivateAlgKey() {
	// Security Key
	P0 := []byte{security.NNASEncAlg}
	L0 := UeauCommon.KDFLen(P0)
	P1 := []byte{ue.UeSecurity.CipheringAlg}
	L1 := UeauCommon.KDFLen(P1)

	kenc := UeauCommon.GetKDFValue(ue.UeSecurity.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.UeSecurity.KnasEnc[:], kenc[16:32])

	// Integrity Key
	P0 = []byte{security.NNASIntAlg}
	L0 = UeauCommon.KDFLen(P0)
	P1 = []byte{ue.UeSecurity.IntegrityAlg}
	L1 = UeauCommon.KDFLen(P1)

	kint := UeauCommon.GetKDFValue(ue.UeSecurity.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.UeSecurity.KnasInt[:], kint[16:32])
}

func (ue *UEContext) SetAuthSubscription(k, opc, op, amf, sqn string) {
	ue.UeSecurity.AuthenticationSubs.PermanentKey = &models.PermanentKey{
		PermanentKeyValue: k,
	}
	ue.UeSecurity.AuthenticationSubs.Opc = &models.Opc{
		OpcValue: opc,
	}
	ue.UeSecurity.AuthenticationSubs.Milenage = &models.Milenage{
		Op: &models.Op{
			OpValue: op,
		},
	}
	ue.UeSecurity.AuthenticationSubs.AuthenticationManagementField = amf

	ue.UeSecurity.AuthenticationSubs.SequenceNumber = sqn
	ue.UeSecurity.AuthenticationSubs.AuthenticationMethod = models.AuthMethod__5_G_AKA
}

func SetUESecurityCapability(ue *UEContext) (UESecurityCapability *nasType.UESecurityCapability) {
	UESecurityCapability = &nasType.UESecurityCapability{
		Iei:    nasMessage.RegistrationRequestUESecurityCapabilityType,
		Len:    8,
		Buffer: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	switch ue.UeSecurity.CipheringAlg {
	case security.AlgCiphering128NEA0:
		UESecurityCapability.SetEA0_5G(1)
	case security.AlgCiphering128NEA1:
		UESecurityCapability.SetEA1_128_5G(1)
	case security.AlgCiphering128NEA2:
		UESecurityCapability.SetEA2_128_5G(1)
	case security.AlgCiphering128NEA3:
		UESecurityCapability.SetEA3_128_5G(1)
	}

	switch ue.UeSecurity.IntegrityAlg {
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

func (ue *UEContext) SetTunInterface(tun netlink.Link) {
	ue.PduSession.tun = tun
}

func (ue *UEContext) GetTunInterface() netlink.Link {
	return ue.PduSession.tun
}

func (ue *UEContext) SetTunRoute(route *netlink.Route) {
	ue.PduSession.routeTun = route
}

func (ue *UEContext) GetTunRoute() *netlink.Route {
	return ue.PduSession.routeTun
}

func (ue *UEContext) SetTunRule(rule *netlink.Rule) {
	ue.PduSession.ruleTun = rule
}

func (ue *UEContext) GetTunRule() *netlink.Rule {
	return ue.PduSession.ruleTun
}

func (ue *UEContext) Terminate() {

	// clean all context of tun interface
	ueTun := ue.GetTunInterface()
	ueRoute := ue.GetTunRoute()
	ueRule := ue.GetTunRule()
	ueUnix := ue.GetUnixConn()

	if ueTun != nil {
		_ = netlink.LinkSetDown(ueTun)
		_ = netlink.LinkDel(ueTun)
	}

	if ueRoute != nil {
		_ = netlink.RouteDel(ueRoute)
	}

	if ueRule != nil {
		_ = netlink.RuleDel(ueRule)
	}

	if ueUnix != nil {
		ueUnix.Close()
	}

	log.Info("[UE] UE Terminated")

}

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux

}
