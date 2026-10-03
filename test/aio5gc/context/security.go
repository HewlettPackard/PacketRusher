/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/mohae/deepcopy"
	"my5G-RANTester/internal/common/auth"

	nasType "github.com/free5gc/nas/ie"
	security "github.com/free5gc/nas/message"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/ueauth"
	log "github.com/sirupsen/logrus"
)

type SecurityContext struct {
	supi               string
	msin               string
	ulCount            security.Count
	dlCount            security.Count
	integrityAlg       uint8
	cipheringAlg       uint8
	xresStar           string
	knasEnc            [16]uint8
	knasInt            [16]uint8
	kamf               string
	authenticationSubs models.Udr_DR_AuthenticationSubscription
	suci               string
	kseaf              string
	kgnb               []uint8
	abba               []uint8
	NH                 []byte
}

func (s *SecurityContext) GetAuthSubscription() models.Udr_DR_AuthenticationSubscription {
	return deepcopy.Copy(s.authenticationSubs).(models.Udr_DR_AuthenticationSubscription)
}

func (s *SecurityContext) SetAuthSubscription(k, opc, op, amf, sqn string) {
	// The upstream model names the key fields Enc*, matching free5GC's local provisioning convention.
	s.authenticationSubs = models.Udr_DR_AuthenticationSubscription{
		EncPermanentKey:               k,
		EncOpcKey:                     opc,
		AuthenticationManagementField: amf,
		SequenceNumber:                &models.Udr_DR_SequenceNumber{Sqn: sqn},
		AuthenticationMethod:          models.Udr_DR_AuthMethod_5_G_AKA,
	}

}

func (s *SecurityContext) GetMsin() string {
	return s.msin
}

func (s *SecurityContext) SetMsin(msin string) {
	s.msin = msin
}

func (s *SecurityContext) SetSuci(suci string) {
	s.suci = suci
}

func (s *SecurityContext) SetSupi(supi string) {
	s.supi = supi
}

func (s *SecurityContext) SetXresStar(xresStar string) {
	s.xresStar = xresStar
}

func (s *SecurityContext) GetXresStar() string {
	return s.xresStar
}

func (s *SecurityContext) GetULCount() security.Count {
	return s.ulCount
}

func (s *SecurityContext) SetULCount(ulCount security.Count) {
	s.ulCount = ulCount
}

func (s *SecurityContext) GetDLCount() security.Count {
	return s.dlCount
}

func (s *SecurityContext) SetDLCount(dlCount security.Count) {
	s.dlCount = dlCount
}

func (s *SecurityContext) GetIntegrityAlg() uint8 {
	return s.integrityAlg
}

func (s *SecurityContext) SetIntegrityAlg(integrityAlg uint8) {
	s.integrityAlg = integrityAlg
}

func (s *SecurityContext) GetCipheringAlg() uint8 {
	return s.cipheringAlg
}

func (s *SecurityContext) SetCipheringAlg(cipheringAlg uint8) {
	s.cipheringAlg = cipheringAlg
}

func (s *SecurityContext) SetKseaf(kseaf string) {
	s.kseaf = kseaf
}

func (s *SecurityContext) GetKseaf() string {
	return s.kseaf
}

func (s *SecurityContext) SetAbba(abba []uint8) {
	s.abba = append([]uint8(nil), abba...)
}

func (s *SecurityContext) GetKnasInt() [16]uint8 {
	return s.knasInt
}

func (s *SecurityContext) GetKnasEnc() [16]uint8 {
	return s.knasEnc
}

// Access Network key Derivation function defined in TS 33.501 Annex A.9
func (s *SecurityContext) DerivateAnKey() {
	accessType := security.AccessType3GPP // Defalut 3gpp
	P0 := make([]byte, 4)
	binary.BigEndian.PutUint32(P0, s.ulCount.Get())
	L0 := ueauth.KDFLen(P0)
	P1 := []byte{uint8(accessType)}
	L1 := ueauth.KDFLen(P1)

	KamfBytes, err := hex.DecodeString(s.kamf)
	if err != nil {
		log.Error(err)
		return
	}
	key, err := ueauth.GetKDFValue(KamfBytes, ueauth.FC_FOR_KGNB_KN3IWF_DERIVATION, P0, L0, P1, L1)
	if err != nil {
		log.Error(err)
		return
	}
	s.kgnb = key
}

// NH Derivation function defined in TS 33.501 Annex A.10
func (s *SecurityContext) DerivateNH(syncInput []byte) {
	P0 := syncInput
	L0 := ueauth.KDFLen(P0)

	KamfBytes, err := hex.DecodeString(s.kamf)
	if err != nil {
		log.Error(err)
		return
	}
	s.NH, err = ueauth.GetKDFValue(KamfBytes, ueauth.FC_FOR_NH_DERIVATION, P0, L0)
	if err != nil {
		log.Error(err)
		return
	}
}

func (s *SecurityContext) UpdateSecurityContext() {

	s.DerivateAnKey()
	s.DerivateNH(s.kgnb)
}

func (s *SecurityContext) GetKGNB() []uint8 {
	return s.kgnb
}

func (s *SecurityContext) DerivateAlgKey() {

	KamfBytes, err := hex.DecodeString(s.kamf)
	if err != nil {
		log.Printf("[5GC] Kamf decode failed: %v", err)
		return
	}

	err = auth.AlgorithmKeyDerivation(s.cipheringAlg,
		KamfBytes,
		&s.knasEnc,
		s.integrityAlg,
		&s.knasInt)

	if err != nil {
		log.Printf("[5GC] Algorithm key derivation failed  %v", err)
	}
}

func (s *SecurityContext) NASSecurityContext() *security.SecCtx {
	return &security.SecCtx{
		Side: security.CoreNetworkSide, Bearer: security.Bearer3GPP,
		UplinkCount: &s.ulCount, DownlinkCount: &s.dlCount,
		CipheringAlg: nasType.AlgCiphering(s.cipheringAlg), IntegrityAlg: nasType.AlgIntegrity(s.integrityAlg),
		KnasEnc: s.knasEnc, KnasInt: s.knasInt,
	}
}

// clone isolates provisioned secrets and pointer/slice fields from callers and
// from each new registration's live NAS counter context.
func (s SecurityContext) clone() SecurityContext {
	out := s
	out.authenticationSubs = deepcopy.Copy(s.authenticationSubs).(models.Udr_DR_AuthenticationSubscription)
	out.abba = append([]uint8(nil), s.abba...)
	out.kgnb = append([]uint8(nil), s.kgnb...)
	out.NH = append([]byte(nil), s.NH...)
	return out
}
