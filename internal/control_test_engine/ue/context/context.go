package context

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/lib/UeauCommon"
	"my5G-RANTester/lib/milenage"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/nas/security"
	"my5G-RANTester/lib/openapi/models"
	"net"
	"regexp"
)

// 5GMM main states in the UE
const MM5G_DEREGISTERED = 0x00
const MM5G_REGISTERED_INITIATED = 0x01
const MM5G_REGISTERED = 0x02
const MM5G_SERVICE_REQ_INIT = 0x03
const MM5G_DEREGISTERED_INIT = 0x04

type UEContext struct {
	ueSecurity SECURITY
	Snssai     models.Snssai
	State      int
	UnixConn   net.Conn
	ueIP       string
}

type SECURITY struct {
	Supi               string
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
}

func (ue *UEContext) NewRanUeContext(imsi string,
	cipheringAlg, integrityAlg uint8,
	k, opc, op, amf, mcc, mnc string,
	sst int32, sd string) {

	// added SUPI.
	ue.ueSecurity.Supi = imsi

	// TODO ue.amfUENgap is received by AMF in authentication request.(? changed this).
	// ue.AmfUeNgapId = ranUeNgapId

	// added ciphering algorithm.
	ue.ueSecurity.CipheringAlg = cipheringAlg

	// added integrity algorithm.
	ue.ueSecurity.IntegrityAlg = integrityAlg

	// added key, AuthenticationManagementField and opc or op.
	ue.SetAuthSubscription(k, opc, op, amf)

	// added suci
	suciV1, suciV2, suciV3 := ue.EncodeUeSuci()

	// added mcc and mnc
	ue.ueSecurity.mcc = mcc
	ue.ueSecurity.mnc = mnc

	// added network slice
	ue.Snssai.Sd = sd
	ue.Snssai.Sst = sst

	// encode mcc and mnc for mobileIdentity5Gs.
	resu := ue.GetMccAndMncInOctets()

	// added suci to mobileIdentity5GS
	if len(ue.ueSecurity.Supi) == 18 {
		ue.ueSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    12, // suci
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, 0x00, suciV3, suciV2, suciV1},
		}
	} else {
		ue.ueSecurity.Suci = nasType.MobileIdentity5GS{
			Len:    13, // suci
			Buffer: []uint8{0x01, resu[0], resu[1], resu[2], 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, suciV3, suciV2, suciV1},
		}
	}

	// added snn.
	ue.ueSecurity.Snn = ue.deriveSNN()

	// added initial state(DEREGISTERED)
	ue.SetState(0x00)
}

func (ue *UEContext) GetSuci() nasType.MobileIdentity5GS {
	return ue.ueSecurity.Suci
}

func (ue *UEContext) SetState(state int) {
	ue.State = state
}

func (ue *UEContext) GetState() int {
	return ue.State
}

func (ue *UEContext) SetUnixConn(conn net.Conn) {
	ue.UnixConn = conn
}

func (ue *UEContext) GetUnixConn() net.Conn {
	return ue.UnixConn
}

func (ue *UEContext) SetIp(ip [12]uint8) {
	ue.ueIP = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ue *UEContext) GetIp() string {
	return ue.ueIP
}

func (ue *UEContext) deriveSNN() string {
	// 5G:mnc093.mcc208.3gppnetwork.org
	var resu string
	if len(ue.ueSecurity.mnc) == 2 {
		resu = "5G:mnc0" + ue.ueSecurity.mnc + ".mcc" + ue.ueSecurity.mcc + ".3gppnetwork.org"
	} else {
		resu = "5G:mnc" + ue.ueSecurity.mnc + ".mcc" + ue.ueSecurity.mcc + ".3gppnetwork.org"
	}

	return resu
}

func (ue *UEContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(ue.ueSecurity.mcc)
	mnc := reverse(ue.ueSecurity.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(ue.ueSecurity.mnc) == 2 {
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

func (ue *UEContext) EncodeUeSuci() (uint8, uint8, uint8) {

	// reverse imsi string.
	aux := reverse(ue.ueSecurity.Supi)

	// calculate decimal value.
	suci, error := hex.DecodeString(aux[:6])
	if error != nil {
		return 0, 0, 0
	}

	// return decimal value
	// Function worked fine.
	return uint8(suci[0]), uint8(suci[1]), uint8(suci[2])
}

func (ue *UEContext) deriveSQN(autn []byte, ak []uint8) []byte {
	sqn := make([]byte, 6)

	// get SQNxorAK
	SQNxorAK := autn[0:6]

	// get sqn
	for i := 0; i < len(SQNxorAK); i++ {
		sqn[i] = SQNxorAK[i] ^ ak[i]
	}

	// return sqn
	return sqn
}

func (ue *UEContext) DeriveRESstarAndSetKey(authSubs models.AuthenticationSubscription, RAND []byte, snNmae string, AUTN []byte) []byte {

	// SQN, _ := hex.DecodeString(authSubs.SequenceNumber)

	// get management field.
	AMF, _ := hex.DecodeString(authSubs.AuthenticationManagementField)

	// Run milenage
	// TODO: verify MAC
	MAC_A, MAC_S := make([]byte, 8), make([]byte, 8)
	CK, IK := make([]byte, 16), make([]byte, 16)
	RES := make([]byte, 8)
	AK, AKstar := make([]byte, 6), make([]byte, 6)

	// generate OPC, K.
	OPC, _ := hex.DecodeString(authSubs.Opc.OpcValue)
	K, _ := hex.DecodeString(authSubs.PermanentKey.PermanentKeyValue)

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// Generate SQN.
	SQN := ue.deriveSQN(AUTN, AK)

	// Generate MAC_A, MAC_S
	milenage.F1_Test(OPC, K, RAND, SQN, AMF, MAC_A, MAC_S)

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345_Test(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// derive RES*
	key := append(CK, IK...)
	FC := UeauCommon.FC_FOR_RES_STAR_XRES_STAR_DERIVATION
	P0 := []byte(snNmae)
	P1 := RAND
	P2 := RES

	ue.DerivateKamf(key, snNmae, SQN, AK)
	ue.DerivateAlgKey()
	kdfVal_for_resStar := UeauCommon.GetKDFValue(key, FC, P0, UeauCommon.KDFLen(P0), P1, UeauCommon.KDFLen(P1), P2, UeauCommon.KDFLen(P2))
	return kdfVal_for_resStar[len(kdfVal_for_resStar)/2:]

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
	groups := supiRegexp.FindStringSubmatch(ue.ueSecurity.Supi)

	P0 = []byte(groups[1])
	L0 := UeauCommon.KDFLen(P0)
	P1 = []byte{0x00, 0x00}
	L1 := UeauCommon.KDFLen(P1)

	ue.ueSecurity.Kamf = UeauCommon.GetKDFValue(Kseaf, UeauCommon.FC_FOR_KAMF_DERIVATION, P0, L0, P1, L1)
}

// Algorithm key Derivation function defined in TS 33.501 Annex A.9
func (ue *UEContext) DerivateAlgKey() {
	// Security Key
	P0 := []byte{security.NNASEncAlg}
	L0 := UeauCommon.KDFLen(P0)
	P1 := []byte{ue.ueSecurity.CipheringAlg}
	L1 := UeauCommon.KDFLen(P1)

	kenc := UeauCommon.GetKDFValue(ue.ueSecurity.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.ueSecurity.KnasEnc[:], kenc[16:32])

	// Integrity Key
	P0 = []byte{security.NNASIntAlg}
	L0 = UeauCommon.KDFLen(P0)
	P1 = []byte{ue.ueSecurity.IntegrityAlg}
	L1 = UeauCommon.KDFLen(P1)

	kint := UeauCommon.GetKDFValue(ue.ueSecurity.Kamf, UeauCommon.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	copy(ue.ueSecurity.KnasInt[:], kint[16:32])
}

func (ue *UEContext) SetAuthSubscription(k, opc, op, amf string) {
	ue.ueSecurity.AuthenticationSubs.PermanentKey = &models.PermanentKey{
		PermanentKeyValue: k,
	}
	ue.ueSecurity.AuthenticationSubs.Opc = &models.Opc{
		OpcValue: opc,
	}
	ue.ueSecurity.AuthenticationSubs.Milenage = &models.Milenage{
		Op: &models.Op{
			OpValue: op,
		},
	}
	ue.ueSecurity.AuthenticationSubs.AuthenticationManagementField = amf

	//ue.ueSecurity.AuthenticationSubs.SequenceNumber = TestGenAuthData.MilenageTestSet19.SQN
	ue.ueSecurity.AuthenticationSubs.AuthenticationMethod = models.AuthMethod__5_G_AKA
}

func SetUESecurityCapability(ue *UEContext) (UESecurityCapability *nasType.UESecurityCapability) {
	UESecurityCapability = &nasType.UESecurityCapability{
		Iei:    nasMessage.RegistrationRequestUESecurityCapabilityType,
		Len:    8,
		Buffer: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	switch ue.ueSecurity.CipheringAlg {
	case security.AlgCiphering128NEA0:
		UESecurityCapability.SetEA0_5G(1)
	case security.AlgCiphering128NEA1:
		UESecurityCapability.SetEA1_128_5G(1)
	case security.AlgCiphering128NEA2:
		UESecurityCapability.SetEA2_128_5G(1)
	case security.AlgCiphering128NEA3:
		UESecurityCapability.SetEA3_128_5G(1)
	}

	switch ue.ueSecurity.IntegrityAlg {
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

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux

}
