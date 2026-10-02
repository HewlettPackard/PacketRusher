// SPDX-License-Identifier: Apache-2.0
// Package naswire constructs independent NAS test packets without native NAS codecs.
package naswire

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"github.com/aead/cmac"
)

// Protect applies NEA2/NIA2 directly to exact inner bytes (TS 33.501, Annex D).
// direction is 0 for uplink and 1 for downlink; NAS 3GPP bearer is 1.
func Protect(plain []byte, enc, integrity [16]byte, count uint32, direction, header byte) []byte {
	counter := make([]byte, 16)
	binary.BigEndian.PutUint32(counter, count)
	counter[4] = 1<<3 | direction<<2
	payload := append([]byte(nil), plain...)
	if header == 2 || header == 4 {
		block, _ := aes.NewCipher(enc[:])
		cipher.NewCTR(block, counter).XORKeyStream(payload, payload)
	}
	seq := byte(count)
	authenticated := append(append([]byte(nil), counter[:8]...), seq)
	authenticated = append(authenticated, payload...)
	block, _ := aes.NewCipher(integrity[:])
	mac, err := cmac.Sum(authenticated, block, 16)
	if err != nil {
		panic(err)
	}
	wire := []byte{0x7e, header, mac[0], mac[1], mac[2], mac[3], seq}
	return append(wire, payload...)
}

func Hex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// WrapSM builds a DL NAS Transport envelope for an independently specified N1 SM message.
func WrapSM(inner []byte) []byte {
	wire := []byte{0x7e, 0, 0x68, 1, byte(len(inner) >> 8), byte(len(inner))}
	wire = append(wire, inner...)
	return append(wire, 0x12, 1)
}

// Fixture identifies the optional native field whose value must be retained.
type Fixture struct {
	Name  string
	Wire  string
	Field string
}

// OptionalIEs covers every v1.1.3-supported optional IE that v1.3.0 left as a
// stub in dispatched MM or nested SM responses. Bytes follow TS 24.501/24.008.
// Opaque containers exercise envelope retention, not their independent procedures.
var OptionalIEs = []Fixture{
	{"registration/mico", "7e00420101b3", "MICOInd"},
	{"registration/emergency", "7e0042010134030201f1", "EmergNumList"},
	{"registration/extended-emergency", "7e004201017a00040001f100", "ExtendedEmergNumList"},
	{"registration/sor", "7e0042010173001300000000000000000000000000000000000000", "SORTransparentCntr"},
	{"registration/nssai-mode", "7e00420101a1", "NSSAIInclusionMode"},
	{"registration/operator-categories", "7e00420101760000", "OperatorDefinedAccessCategoryDefs"},
	{"registration/non3gpp-policies", "7e00420101d1", "Non3GppNWPolicies"},
	{"registration/eps-status", "7e0042010160022000", "EPSBearerCtxStatus"},
	{"configuration/mico", "7e0054b0", "MICOInd"},
	{"configuration/operator-categories", "7e0054760000", "OperatorDefinedAccessCategoryDefs"},
	{"configuration/full-name-ucs2", "7e00544307900046005200e9", "FullNameForNw"},
	{"configuration/short-name-ucs2", "7e005445039000e9", "ShortNameForNw"},
	{"configuration/sms", "7e0054f1", "SMSInd"},
	{"security/eps-algorithms", "7e005d220102a0205722", "SelectedEPSNASSecAlgos"},
	{"security/s1-capabilities", "7e005d220102a0201902a020", "ReplayedS1UESecCapabilities"},
	{"transport/additional-information", "7e00680100052e0101c31f240100", "AdditionalInfo"},
	{"session/always-on", "2e0101c211000601000320ff01060103e80103e881", "AlwaysonPDUSessInd"},
	{"session/mapped-eps", "2e0101c211000601000320ff01060103e80103e875000450000180", "MappedEPSBearerCtxs"},
	{"session/allowed-ssc", "2e0101c31ff3", "AllowedSSCMode"},
	{"session/reject-congestion", "2e0101c31f610103", "CongestionReattemptIndicator5GSM"},
	{"session/release-congestion", "2e0101d324610103", "CongestionReattemptIndicator5GSM"},
}
