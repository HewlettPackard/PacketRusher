package context

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"my5G-RANTester/lib/UeauCommon"
	"my5G-RANTester/lib/milenage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/openapi/models"
	"net"
	"reflect"
	"regexp"
	"sync"
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
	id           uint8
	UeSecurity   SECURITY
	StateMM      int
	StateSM      int
	gnbRx        chan context.UEMessage
	gnbTx        chan context.UEMessage
	PduSession   [16]*PDUSession
	amfInfo      Amf

	// TODO: Modify config so you can configure these parameters per PDUSession
	Dnn           string
	Snssai        models.Snssai
	TunnelEnabled bool

	// Sync primitive
	scenarioChan chan scenario.ScenarioMessage

	lock sync.Mutex
}

type Amf struct {
	amfRegionId uint8
	amfSetId    uint16
	amfPointer  uint8
	amfUeId     int64
}

type PDUSession struct {
	Id         uint8
	ueIP       string
	ueGnbIP    net.IP
	tun        netlink.Link
	routeTun   *netlink.Route
	vrf        *netlink.Vrf
	stopSignal chan bool
	Wait       chan bool
}

type SECURITY struct {
	Supi                 string
	Msin                 string
	mcc                  string
	mnc                  string
	ULCount              security.Count
	DLCount              security.Count
	UeSecurityCapability *nasType.UESecurityCapability
	IntegrityAlg         uint8
	CipheringAlg         uint8
	Snn                  string
	KnasEnc              [16]uint8
	KnasInt              [16]uint8
	Kamf                 []uint8
	AuthenticationSubs   models.AuthenticationSubscription
	Suci                 nasType.MobileIdentity5GS
	RoutingIndicator     string
	Guti                 [4]byte
}

func (ue *UEContext) NewRanUeContext(msin string,
	ueSecurityCapability *nasType.UESecurityCapability,
	k, opc, op, amf, sqn, mcc, mnc, routingIndicator, dnn string,
	sst int32, sd string, tunnelEnabled bool, scenarioChan chan scenario.ScenarioMessage,
	id uint8) {

	// added SUPI.
	ue.UeSecurity.Msin = msin

	// added ciphering algorithm.
	ue.UeSecurity.UeSecurityCapability = ueSecurityCapability
	// set the algorithms of integrity
	if ueSecurityCapability.GetIA0_5G() == 1 {
		ue.UeSecurity.IntegrityAlg = security.AlgIntegrity128NIA0
	} else if ueSecurityCapability.GetIA1_128_5G() == 1 {
		ue.UeSecurity.IntegrityAlg = security.AlgIntegrity128NIA1
	} else if ueSecurityCapability.GetIA2_128_5G() == 1 {
		ue.UeSecurity.IntegrityAlg = security.AlgIntegrity128NIA2
	}

	// set the algorithms of ciphering
	if ueSecurityCapability.GetEA0_5G() == 1 {
		ue.UeSecurity.CipheringAlg = security.AlgCiphering128NEA0
	} else if ueSecurityCapability.GetEA1_128_5G() == 1 {
		ue.UeSecurity.CipheringAlg = security.AlgCiphering128NEA1
	} else if ueSecurityCapability.GetEA2_128_5G() == 1 {
		ue.UeSecurity.CipheringAlg = security.AlgCiphering128NEA2
	}
	// added key, AuthenticationManagementField and opc or op.
	ue.SetAuthSubscription(k, opc, op, amf, sqn)

	// added suci
	suciV1, suciV2, suciV3, suciV4, suciV5 := ue.EncodeUeSuci()

	// added mcc and mnc
	ue.UeSecurity.mcc = mcc
	ue.UeSecurity.mnc = mnc

	// added routing indidcator
	ue.UeSecurity.RoutingIndicator = routingIndicator

	// added supi
	ue.UeSecurity.Supi = fmt.Sprintf("imsi-%s%s%s", mcc, mnc, msin)

	// added UE id.
	ue.id = id

	// added network slice
	ue.Snssai.Sd = sd
	ue.Snssai.Sst = sst

	// added Domain Network Name.
	ue.Dnn = dnn
	ue.TunnelEnabled = tunnelEnabled

	ue.gnbRx = make(chan context.UEMessage, 1)
	ue.gnbTx = make(chan context.UEMessage, 1)

	// encode mcc and mnc for mobileIdentity5Gs.
	resu := ue.GetMccAndMncInOctets()
	encodedRoutingIndicator := ue.GetRoutingIndicatorInOctets()

	// added suci to mobileIdentity5GS
	if len(ue.UeSecurity.Msin) == 8 {
		ue.UeSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    12,
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], encodedRoutingIndicator[0], encodedRoutingIndicator[1], 0x00, 0x00, suciV4, suciV3, suciV2, suciV1},
		}
	} else if len(ue.UeSecurity.Msin) == 10 {
		ue.UeSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    13,
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], encodedRoutingIndicator[0], encodedRoutingIndicator[1], 0x00, 0x00, suciV5, suciV4, suciV3, suciV2, suciV1},
		}
	}

	// added snn.
	ue.UeSecurity.Snn = ue.deriveSNN()

	ue.scenarioChan = scenarioChan

	// added initial state for MM(NULL)
	ue.StateMM = MM5G_NULL

	// added initial state for SM(INACTIVE)
	ue.SetStateSM_PDU_SESSION_INACTIVE()
}

func (ue *UEContext) CreatePDUSession() (*PDUSession, error) {
	pduSessionIndex := -1
	for i, pduSession := range ue.PduSession {
		if pduSession == nil {
			pduSessionIndex = i
			break
		}
	}

	if pduSessionIndex == -1 {
		return nil, errors.New("unable to create an additional PDU Session, we already created the max number of PDU Session")
	}

	pduSession := &PDUSession{}
	pduSession.Id = uint8(pduSessionIndex + 1)
	pduSession.Wait = make(chan bool)

	ue.PduSession[pduSessionIndex] = pduSession

	return pduSession, nil
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
	ue.scenarioChan <- scenario.ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UEContext) SetStateMM_MM5G_SERVICE_REQ_INIT() {
	ue.StateMM = MM5G_SERVICE_REQ_INIT
	ue.scenarioChan <- scenario.ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UEContext) SetStateMM_REGISTERED_INITIATED() {
	ue.StateMM = MM5G_REGISTERED_INITIATED
	ue.scenarioChan <- scenario.ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UEContext) SetStateMM_REGISTERED() {
	ue.StateMM = MM5G_REGISTERED
	ue.scenarioChan <- scenario.ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UEContext) SetStateMM_NULL() {
	ue.StateMM = MM5G_NULL
}

func (ue *UEContext) SetStateMM_DEREGISTERED() {
	ue.StateMM = MM5G_DEREGISTERED
	ue.scenarioChan <- scenario.ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UEContext) GetStateSM() int {
	return ue.StateSM
}

func (ue *UEContext) GetStateMM() int {
	return ue.StateMM
}

func (ue *UEContext) SetGnbRx(gnbRx chan context.UEMessage) {
	ue.gnbRx = gnbRx
}

func (ue *UEContext) SetGnbTx(gnbTx chan context.UEMessage) {
	ue.gnbTx = gnbTx
}

func (ue *UEContext) GetGnbRx() chan context.UEMessage {
	return ue.gnbRx
}

func (ue *UEContext) GetGnbTx() chan context.UEMessage {
	return ue.gnbTx
}

func (ue *UEContext) Lock() {
	ue.lock.Lock()
}

func (ue *UEContext) Unlock() {
	ue.lock.Unlock()
}

func (ue *UEContext) IsTunnelEnabled() bool {
	return ue.TunnelEnabled
}

func (ue *UEContext) GetPduSession(pduSessionid uint8) (*PDUSession, error) {
	if pduSessionid > 15 || ue.PduSession[pduSessionid-1] == nil {
		return nil, errors.New("Unable to find PDUSession ID " + string(pduSessionid))
	}
	return ue.PduSession[pduSessionid-1], nil
}

func (ue *UEContext) DeletePduSession(pduSessionid uint8) error {
	if pduSessionid > 15 || ue.PduSession[pduSessionid-1] == nil {
		return errors.New("Unable to find PDUSession ID " + string(pduSessionid))
	}
	pduSession := ue.PduSession[pduSessionid-1]
	close(pduSession.Wait)
	stopSignal := pduSession.GetStopSignal()
	if stopSignal != nil {
		stopSignal <- true
	}
	ue.PduSession[pduSessionid-1] = nil
	return nil
}

func (pduSession *PDUSession) SetIp(ip [12]uint8) {
	pduSession.ueIP = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (pduSession *PDUSession) GetIp() string {
	return pduSession.ueIP
}

func (pduSession *PDUSession) SetGnbIp(ip net.IP) {
	pduSession.ueGnbIP = ip
}

func (pduSession *PDUSession) GetGnbIp() net.IP {
	return pduSession.ueGnbIP
}

func (pduSession *PDUSession) SetStopSignal(stopSignal chan bool) {
	pduSession.stopSignal = stopSignal
}

func (pduSession *PDUSession) GetStopSignal() chan bool {
	return pduSession.stopSignal
}

func (pduSession *PDUSession) GetPduSesssionId() uint8 {
	return pduSession.Id
}

func (pduSession *PDUSession) SetTunInterface(tun netlink.Link) {
	pduSession.tun = tun
}

func (pduSession *PDUSession) GetTunInterface() netlink.Link {
	return pduSession.tun
}

func (pduSession *PDUSession) SetTunRoute(route *netlink.Route) {
	pduSession.routeTun = route
}

func (pduSession *PDUSession) GetTunRoute() *netlink.Route {
	return pduSession.routeTun
}

func (pduSession *PDUSession) SetVrfDevice(vrf *netlink.Vrf) {
	pduSession.vrf = vrf
}

func (pduSession *PDUSession) GetVrfDevice() *netlink.Vrf {
	return pduSession.vrf
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

func (ue *UEContext) GetUeSecurityCapability() *nasType.UESecurityCapability {
	return ue.UeSecurity.UeSecurityCapability
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

// TS 24.501 9.11.3.4.1
// Routing Indicator shall consist of 1 to 4 digits. The coding of this field is the
// responsibility of home network operator but BCD coding shall be used. If a network
// operator decides to assign less than 4 digits to Routing Indicator, the remaining digits
// shall be coded as "1111" to fill the 4 digits coding of Routing Indicator (see NOTE 2). If
// no Routing Indicator is configured in the USIM, the UE shall coxde bits 1 to 4 of octet 8
// of the Routing Indicator as "0000" and the remaining digits as “1111".
func (ue *UEContext) GetRoutingIndicatorInOctets() []byte {
	if len(ue.UeSecurity.RoutingIndicator) == 0 {
		ue.UeSecurity.RoutingIndicator = "0"
	}

	if len(ue.UeSecurity.RoutingIndicator) > 4 {
		log.Fatal("[UE][CONFIG] Routing indicator must be 4 digits maximum, ", ue.UeSecurity.RoutingIndicator, " is invalid")
	}

	routingIndicator := []byte(ue.UeSecurity.RoutingIndicator)
	for len(routingIndicator) < 4 {
		routingIndicator = append(routingIndicator, 'F')
	}

	// Reverse the bytes in group of two
	for i := 1; i < len(routingIndicator); i += 2 {
		tmp := routingIndicator[i-1]
		routingIndicator[i-1] = routingIndicator[i]
		routingIndicator[i] = tmp
	}

	// BCD conversion
	encodedRoutingIndicator, err := hex.DecodeString(string(routingIndicator))
	if err != nil {
		log.Fatal("[UE][CONFIG] Unable to encode routing indicator ", err)
	}

	return encodedRoutingIndicator
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

func (ue *UEContext) SetAmfUeId(id int64) {
	ue.amfInfo.amfUeId = id
}

func (ue *UEContext) GetAmfUeId() int64 {
	return ue.amfInfo.amfUeId
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
	OPC, err := hex.DecodeString(authSubs.Opc.OpcValue)
	if err != nil {
		log.Fatal("[UE] OPC error: ", err, authSubs.Opc.OpcValue)
	}
	K, err := hex.DecodeString(authSubs.PermanentKey.PermanentKeyValue)
	if err != nil {
		log.Fatal("[UE] K error: ", err, authSubs.PermanentKey.PermanentKeyValue)
	}
	sqnUe, err := hex.DecodeString(authSubs.SequenceNumber)
	if err != nil {
		log.Fatal("[UE] sqn error: ", err, authSubs.SequenceNumber)
	}
	AMF, err := hex.DecodeString(authSubs.AuthenticationManagementField)
	if err != nil {
		log.Fatal("[UE] AuthenticationManagementField error: ", err, authSubs.AuthenticationManagementField)
	}

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
	log.WithFields(log.Fields{
		"k":   k,
		"opc": opc,
		"op":  op,
		"amf": amf,
		"sqn": sqn,
	}).Info("[UE] Authentification parameters:")

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

func (ue *UEContext) Terminate() {
	ue.SetStateMM_NULL()

	// clean all context of tun interface
	for _, pduSession := range ue.PduSession {
		if pduSession != nil {
			ueTun := pduSession.GetTunInterface()
			ueRoute := pduSession.GetTunRoute()
			ueVrf := pduSession.GetVrfDevice()

			if ueTun != nil {
				_ = netlink.LinkSetDown(ueTun)
				_ = netlink.LinkDel(ueTun)
			}

			if ueRoute != nil {
				_ = netlink.RouteDel(ueRoute)
			}

			if ueVrf != nil {
				_ = netlink.LinkSetDown(ueVrf)
				_ = netlink.LinkDel(ueVrf)
			}
		}
	}

	ue.Lock()
	close(ue.gnbRx)
	ue.gnbRx = nil
	ue.Unlock()
	close(ue.scenarioChan)

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
