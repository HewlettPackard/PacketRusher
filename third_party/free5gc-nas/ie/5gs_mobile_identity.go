package ie

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// MobileId5GS is detailed in 9.11.3.4 5GS mobile identity, 24.501
// This structure handles various 5GS mobile identity types including SUCI, GUTI, IMEI, etc.
type MobileId5GS struct {
	TypeOfId     uint8 // IdType_5GS_xxx, see 5gs_identity_type.go
	AllOneBits   uint8
	MAURI        uint8
	OddEvenIndic uint8

	AMFRegionID uint8
	AMFPointer  uint8
	AMFSetID    uint16

	AMFId string // Hex String

	ProtectionSchemeId uint8
	HomeNwPubKeyId     uint8

	EUI64      [8]uint8
	IMEISV     [16]uint8 // 6.2.2 Composition of IMEISV, 23.003
	IMEI       [15]uint8 // 6.2.1 Composition of IMEI, 23.003
	SUPIFormat uint8

	PlmnId

	RoutingIndDigit [4]uint8
	TMSI5G          [4]byte

	MACAddr    [6]uint8
	MSINDigits [10]uint8 // MSIN digits stored as individual BCD values (0-9)
	MSINLength int       // Actual number of MSIN digits (excluding padding)

	SchemeOutput []byte
	ECCEphPubKey []byte
	CipherVal    []byte
	MACTag       []byte
	SUCINAI      []uint8
}

type SUPIType uint8

const (
	SupiIMSI                 SUPIType = 0x00 // IMSI format as defined in 23.003
	SupiNwSpecificIdentifier SUPIType = 0x01 // Network specific identifier
	SupiGCI                  SUPIType = 0x02 // Global Cable Identifier
	SupiGLI                  SUPIType = 0x03 // Global Line Identifier

	EvenNumOfIdDigit uint8 = 0x0 // Even number of identity digits
	OddNumOfIdDigit  uint8 = 0x1 // Odd number of identity digits

	NullScheme          uint8 = 0x00 // No protection scheme (plain text)
	ECIESSchemeProfileA uint8 = 0x01 // ECIES protection scheme profile A
	ECIESSchemeProfileB uint8 = 0x02 // ECIES protection scheme profile B
)

var idTypeStr = map[uint8]string{
	IdType_5GS_None:       "None",
	IdType_5GS_SUCI:       "SUCI",
	IdType_5GS_GUTI:       "5G-GUTI",
	IdType_5GS_IMEI:       "IMEI",
	IdType_5GS_TMSI:       "5G-S-TMSI",
	IdType_5GS_IMEISV:     "IMEISV",
	IdType_5GS_MACAddress: "MACAddress",
	IdType_5GS_EUI64:      "EUI64",
}

func IdType5GSStr(id uint8) string {
	str, ok := idTypeStr[id]
	if !ok {
		str = fmt.Sprintf("unknown ID type(%d)", id)
	}
	return str
}

func (i *MobileId5GS) IdStr() string {
	switch i.TypeOfId {
	case IdType_5GS_GUTI:
		return i.GUTIStr()
	case IdType_5GS_IMEI:
		fallthrough
	case IdType_5GS_IMEISV:
		return i.PEIStr()
	case IdType_5GS_TMSI:
		return i.FiveGSTMSIStr()
	case IdType_5GS_MACAddress:
		return i.MACAddrStr()
	case IdType_5GS_EUI64:
		return i.EUI64Str()
	case IdType_5GS_SUCI:
		fallthrough
	default:
		return i.SUCIStr()
	}
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *MobileId5GS) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("MobileId5GS IE len(%d) < 1", len(b))
	}
	i.TypeOfId = Get3Bits31(b[0])

	switch i.TypeOfId {
	case IdType_5GS_None:
		return i.unmarshalBinary_None(b)
	case IdType_5GS_SUCI:
		return i.unmarshalBinary_SUCI(b)
	case IdType_5GS_GUTI:
		return i.unmarshalBinary_5GGUTI(b)
	case IdType_5GS_IMEI:
		return i.unmarshalBinary_IMEI(b)
	case IdType_5GS_TMSI:
		return i.unmarshalBinary_5GSTMSI(b)
	case IdType_5GS_IMEISV:
		return i.unmarshalBinary_IMEISV(b)
	case IdType_5GS_MACAddress:
		return i.unmarshalBinary_MACAddr(b)
	case IdType_5GS_EUI64:
		return i.unmarshalBinary_EUI64(b)
	default:
		return errors.Errorf("MobileId5GS: Unknown type %d\n", i.TypeOfId)
	}
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MobileId5GS) MarshalBinary() ([]byte, error) {
	switch i.TypeOfId {
	case IdType_5GS_None:
		return i.marshalBinary_None()
	case IdType_5GS_SUCI:
		return i.marshalBinary_SUCI()
	case IdType_5GS_GUTI:
		return i.marshalBinary_5GGUTI()
	case IdType_5GS_IMEI:
		return i.marshalBinary_IMEI()
	case IdType_5GS_TMSI:
		return i.marshalBinary_5GSTMSI()
	case IdType_5GS_IMEISV:
		return i.marshalBinary_IMEISV()
	case IdType_5GS_MACAddress:
		return i.marshalBinary_MACAddr()
	case IdType_5GS_EUI64:
		return i.marshalBinary_EUI64()
	default:
		return nil, errors.Errorf("MobileId5GS: Unknown type %d\n", i.TypeOfId)
	}
}

/*************** NoIdentity ***************/

func (i *MobileId5GS) unmarshalBinary_None(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("unmarshalBinary_None: incorrect IE length(%d)", len(b))
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MobileId5GS) marshalBinary_None() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	return b, nil
}

/*************** SUCI ***************/

// getMSIN decodes MSIN digits from SchemeOutput according to 3GPP TS 24.501
// MSIN is encoded using BCD (Binary-Coded Decimal) format where each octet contains two digits
// For odd-length MSIN, the last nibble is padded with 0xF as per 3GPP specification
func (i *MobileId5GS) getMSIN() error {
	// Reset MSINLength and clear MSINDigits array to avoid data contamination
	i.MSINLength = 0
	for j := range i.MSINDigits {
		i.MSINDigits[j] = 0
	}

	// According to 3GPP spec: MSIN uses BCD encoding with two digits packed per octet
	// If total number of octets is odd, the last nibble is padded with 0xF
	totalBytes := len(i.SchemeOutput)
	if totalBytes == 0 {
		return nil
	}

	// Decode all complete octets (two digits each)
	for soi := 0; soi < totalBytes-1; soi++ {
		if len(i.MSINDigits)-1 < soi*2+1 {
			return errors.Errorf(
				"getMSIN(): soi*2+1 (=%d) overflows i.MSINDigits (len=%d)", soi*2+1, len(i.MSINDigits),
			)
		}
		i.MSINDigits[soi*2+1], i.MSINDigits[soi*2] = GetHalfValue(i.SchemeOutput[soi])
		i.MSINLength += 2
	}

	// Handle the last octet (may contain padding)
	lastByte := i.SchemeOutput[totalBytes-1]
	highNibble, lowNibble := GetHalfValue(lastByte)

	// Always add the low nibble (always valid)
	if len(i.MSINDigits)-1 < (totalBytes-1)*2+1 {
		return errors.Errorf(
			"getMSIN(): (totalBytes-1)*2+1 (=%d) overflows i.MSINDigits (len=%d)",
			(totalBytes-1)*2+1, len(i.MSINDigits),
		)
	}
	i.MSINDigits[(totalBytes-1)*2+1], i.MSINDigits[(totalBytes-1)*2] = highNibble, lowNibble
	i.MSINLength++

	// Check if high nibble is padding value 0xF
	// If not 0xF, it's a valid digit; if 0xF, it's padding and should not be counted
	if highNibble != 0x0f {
		i.MSINLength++
	}

	return nil
}

// unmarshalBinary_SUCISupiImsi unmarshals SUCI with IMSI format according to 3GPP TS 24.501
// SUCI structure: Type(1) + SUPI Format(1) + PLMN ID(3) + Routing Indicator(2) +
// Protection Scheme ID(1) + Home Network Public Key ID(1) + Scheme Output(variable)
func (i *MobileId5GS) unmarshalBinary_SUCISupiImsi(b []byte) error {
	blen := len(b)
	if blen < 9 {
		return errors.Errorf("unmarshalBinary_SUCISupiImsi: incorrect IE length(%d)", blen)
	}

	// Decode PLMN ID (MCC + MNC) from octets 1-3
	if err := i.PlmnId.UnmarshalBinary(b[1:4]); err != nil {
		return errors.Wrap(err, "unmarshalBinary_SUCISupiImsi")
	}

	// Decode Routing Indicator digits (4 digits packed into 2 octets)
	i.RoutingIndDigit[1], i.RoutingIndDigit[0] = GetHalfValue(b[4])
	i.RoutingIndDigit[3], i.RoutingIndDigit[2] = GetHalfValue(b[5])
	_, i.ProtectionSchemeId = GetHalfValue(b[6])

	// Decode Home Network Public Key Identifier
	i.HomeNwPubKeyId = b[7]
	i.SchemeOutput = b[8:]

	switch i.ProtectionSchemeId {
	case NullScheme:
		// For Null scheme, Scheme Output contains MSIN digits in BCD format
		if err := i.getMSIN(); err != nil {
			return errors.Wrap(err, "unmarshalBinary_SUCISupiImsi fail")
		}
	case ECIESSchemeProfileA:
		// ECIES Profile A: 32-byte ephemeral public key + variable cipher + 8-byte MAC
		if len(i.SchemeOutput) < 40 {
			return errors.Errorf("unmarshalBinary_SUCISupiImsi: SchemeOutput length (%d), < 40 octets",
				len(i.SchemeOutput))
		}
		i.ECCEphPubKey = i.SchemeOutput[:32]
		i.CipherVal = i.SchemeOutput[32 : len(i.SchemeOutput)-8]
		i.MACTag = i.SchemeOutput[len(i.SchemeOutput)-8:]
	case ECIESSchemeProfileB:
		// ECIES Profile B: 33-byte ephemeral public key + variable cipher + 8-byte MAC
		if len(i.SchemeOutput) < 41 {
			return errors.Errorf("unmarshalBinary_SUCISupiImsi: SchemeOutput length (%d), < 41 octets",
				len(i.SchemeOutput))
		}
		i.ECCEphPubKey = i.SchemeOutput[:33]
		i.CipherVal = i.SchemeOutput[33 : len(i.SchemeOutput)-8]
		i.MACTag = i.SchemeOutput[len(i.SchemeOutput)-8:]
	default:
	}

	i.SchemeOutput = nil
	return nil
}

func (i *MobileId5GS) unmarshalBinary_SUCI(b []byte) error {
	blen := len(b)
	if blen < 1 {
		return errors.Errorf("unmarshalBinary_SUCI: incorrect IE length(%d)", blen)
	}
	i.SUPIFormat = Get3Bits75(b[0])
	switch SUPIType(i.SUPIFormat) {
	case SupiIMSI:
		if err := i.unmarshalBinary_SUCISupiImsi(b); err != nil {
			return errors.Wrap(err, "unmarshalBinary_SUCI failed")
		}
	case SupiNwSpecificIdentifier:
		fallthrough
	case SupiGCI:
		fallthrough
	case SupiGLI:
		i.SUCINAI = make([]byte, blen-1)
		if nil == i.SUCINAI {
			return errors.Errorf("unmarshalBinary_SUCI :Failed to make %d []byte", blen-1)
		}
		copy(i.SUCINAI, b[1:])
	default:
	}
	return nil
}

// marshalBinary_SUCISupiImsi marshals SUCI with IMSI format according to 3GPP TS 24.501
// The function handles different protection schemes and encodes MSIN using BCD format
func (i *MobileId5GS) marshalBinary_SUCISupiImsi() ([]byte, error) {
	blen := 0
	switch i.ProtectionSchemeId {
	case NullScheme:
		// Each byte can store 2 BCD digits, so (MSINLength + 1) / 2 bytes are needed
		// This handles both even and odd MSIN lengths correctly
		blen = 8 + (i.MSINLength+1)/2
	case ECIESSchemeProfileA:
		if len(i.ECCEphPubKey) != 32 || len(i.MACTag) != 8 {
			return nil, errors.Errorf("Profile A requires a 32-byte ephemeral key and 8-byte MAC")
		}
		blen = 8 + 32 + 8 + len(i.CipherVal)
	case ECIESSchemeProfileB:
		if len(i.ECCEphPubKey) != 33 || len(i.MACTag) != 8 {
			return nil, errors.Errorf("Profile B requires a 33-byte ephemeral key and 8-byte MAC")
		}
		blen = 8 + 33 + 8 + len(i.CipherVal)
	default:
		return nil, errors.Errorf("marshalBinary_SUCISupiImsi : bad i.ProtectionSchemeId(%d)", i.ProtectionSchemeId)
	}
	b := make([]byte, blen)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	b[0] = Set3Bits75(b[0], i.SUPIFormat)
	if err := i.PlmnId.MarshalBinary(b[1:4]); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_SUCISupiImsi PlmnId err")
	}
	b[4] = SetHalfValue(i.RoutingIndDigit[1], i.RoutingIndDigit[0])
	b[5] = SetHalfValue(i.RoutingIndDigit[3], i.RoutingIndDigit[2])
	b[6] = Set4Bits41(b[6], i.ProtectionSchemeId)
	b[7] = i.HomeNwPubKeyId

	ofs := 8 // offset
	switch i.ProtectionSchemeId {
	case NullScheme:
		// Encode MSIN digits using BCD format (two digits per octet)
		msinLen := i.MSINLength
		j := 0
		for ; msinLen > 1; j, msinLen = j+1, msinLen-2 {
			if len(b)-1 < ofs {
				return nil, errors.Errorf(
					"marshalBinary_SUCISupiImsi(): ofs (%d) overflows b (len=%d)", ofs, len(b))
			}
			b[ofs] = SetHalfValue(i.MSINDigits[j*2+1], i.MSINDigits[j*2])
			ofs++
		}
		if msinLen == 1 {
			// For odd-length MSIN, pad the last nibble with 0xF as per 3GPP spec
			if len(b)-1 < ofs {
				return nil, errors.Errorf(
					"marshalBinary_SUCISupiImsi(): ofs (%d) overflows b (len=%d)", ofs, len(b))
			}
			b[ofs] = SetHalfValue(0xff, i.MSINDigits[j*2])
		}
	case ECIESSchemeProfileA:
		if len(b)-1 < ofs {
			return nil, errors.Errorf(
				"marshalBinary_SUCISupiImsi(): ofs (%d) overflows b (len=%d)", ofs, len(b))
		}
		copy(b[ofs:], i.ECCEphPubKey)
		ofs += 32
		copy(b[ofs:], i.CipherVal)
		ofs += len(i.CipherVal)
		copy(b[ofs:], i.MACTag)
	case ECIESSchemeProfileB:
		if len(b)-1 < ofs {
			return nil, errors.Errorf(
				"marshalBinary_SUCISupiImsi(): ofs (%d) overflows b (len=%d)", ofs, len(b))
		}
		copy(b[ofs:], i.ECCEphPubKey)
		ofs += 33
		copy(b[ofs:], i.CipherVal)
		ofs += len(i.CipherVal)
		copy(b[ofs:], i.MACTag)
	}
	return b, nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MobileId5GS) marshalBinary_SUCI() ([]byte, error) {
	switch SUPIType(i.SUPIFormat) {
	case SupiIMSI:
		return i.marshalBinary_SUCISupiImsi()
	case SupiNwSpecificIdentifier:
		fallthrough
	case SupiGCI:
		fallthrough
	case SupiGLI:
		b := make([]byte, len(i.SUCINAI)+1)
		b[0] = Set3Bits31(b[0], i.TypeOfId)
		b[0] = Set3Bits75(b[0], i.SUPIFormat)
		copy(b[1:], i.SUCINAI)
		return b, nil
	default:
	}
	return nil, nil
}

// rtIndDigitsStr converts routing indicator digits to string format
// Routing indicator is 4 digits that identify the routing information
func (i *MobileId5GS) rtIndDigitsStr() string {
	var str string
	for _, d := range i.RoutingIndDigit {
		if d < 10 {
			str = str + strconv.Itoa(int(d))
		}
	}
	if len(str) == 0 {
		return "0"
	}
	return str
}

// SUCIStr formats the SUCI as a human-readable string according to 3GPP TS 24.501
// Format: suci-<type>-<mcc>-<mnc>-<routing-indicator>-<protection-scheme>-<home-network-pub-key-id>-<scheme-output>
func (i *MobileId5GS) SUCIStr() string {
	if IdType_5GS_SUCI != i.TypeOfId {
		return ""
	}
	var suci string
	switch i.ProtectionSchemeId {
	case NullScheme:
		// For Null scheme, convert MSIN digits to string representation
		msinStr := []byte{}
		for _, m := range i.MSINDigits[:i.MSINLength] {
			msinStr = append(msinStr[:], '0'+m)
		}
		suci = fmt.Sprintf("suci-0-%s-%s-%s-%d-%d-%s",
			i.MCC, i.MNC, i.rtIndDigitsStr(),
			i.ProtectionSchemeId, i.HomeNwPubKeyId,
			string(msinStr))
	case ECIESSchemeProfileA:
		fallthrough
	case ECIESSchemeProfileB:
		// For ECIES schemes, encode scheme output as hex strings
		suci = fmt.Sprintf("suci-0-%s-%s-%s-%d-%d-%s%s%s",
			i.MCC, i.MNC, i.rtIndDigitsStr(),
			i.ProtectionSchemeId, i.HomeNwPubKeyId,
			hex.EncodeToString(i.ECCEphPubKey),
			hex.EncodeToString(i.CipherVal),
			hex.EncodeToString(i.MACTag))
	}
	return suci
}

/*************** 5G GUTI ***************/

// unmarshalBinary_5GGUTI unmarshals 5G-GUTI according to 3GPP TS 24.501
// 5G-GUTI structure: Type(1) + AllOneBits(1) + PLMN ID(3) + AMF Region ID(1) +
// AMF Set ID + AMF Pointer(2) + 5G-TMSI(4)
func (i *MobileId5GS) unmarshalBinary_5GGUTI(b []byte) error {
	var err error
	if len(b) != 11 {
		return errors.Errorf("UnmarshalBinary_5GGUTI: incorrect IE length(%d)", len(b))
	}

	// First octet must start with 0xF0 as per 3GPP spec
	i.AllOneBits = Get4Bits85(b[0])
	if i.AllOneBits != 0x0f {
		return errors.Errorf("UnmarshalBinary_5GGUTI: octet4 not start with 0xF0")
	}

	// Decode PLMN ID (MCC + MNC)
	if err = i.PlmnId.UnmarshalBinary(b[1:4]); err != nil {
		return errors.Wrap(err, "unmarshalBinary_5GGUTI")
	}
	i.AMFRegionID = b[4]
	if i.AMFSetID, err = GetMaskedValue16(b[5:7], BIT16, BIT7); err != nil {
		return errors.Wrap(err, "unmarshalBinary_5GGUTI")
	}
	i.AMFPointer = Get6Bits61(b[6])
	i.AMFId = hex.EncodeToString(b[4:7])
	slTMSI5G := i.TMSI5G[:]
	copy(slTMSI5G, b[7:11])
	return nil
}

// marshalBinary_5GGUTI marshals 5G-GUTI according to 3GPP TS 24.501
// Encodes all components into the standard 11-octet format
func (i *MobileId5GS) marshalBinary_5GGUTI() ([]byte, error) {
	var err error
	b := make([]byte, 11)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	b[0] = Set4Bits85(b[0], 0xF) // Set AllOneBits to 0xF as per spec

	// Encode PLMN ID
	if err = i.PlmnId.MarshalBinary(b[1:4]); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_5GGUTI")
	}
	b[4] = i.AMFRegionID

	// Encode AMF Set ID and AMF Pointer in the same octets
	if err = SetMaskedValue16(b[5:7], i.AMFSetID, BIT16, BIT7); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_5GGUTI AMFSetID")
	}
	b[6] = Set6Bits61(b[6], i.AMFPointer)
	slTMSI5G := i.TMSI5G[:]
	copy(b[7:11], slTMSI5G)
	return b, nil
}

func (i *MobileId5GS) GUTIStr() string {
	return i.MCC + i.MNC + i.AMFId + i.TMSIStr()
}

func (i *MobileId5GS) FromGUTIStr(s string) error {
	if len(s) != 20 && len(s) != 19 {
		return errors.Errorf("FromGUTIStr: Bad GUTI String len=%d", len(s))
	}
	i.TypeOfId = IdType_5GS_GUTI
	i.AllOneBits = 0x0f
	i.MCC = s[0:3]
	if len(s) == 20 {
		i.MNC = s[3:6]
		s = s[6:]
	} else {
		i.MNC = s[3:5]
		s = s[5:]
	}
	// AMFId
	i.AMFId = s[0:6]
	if amfId, err := hex.DecodeString(i.AMFId); err != nil {
		return errors.Wrap(err, "FromGUTIStr: Decode amfid")
	} else {
		i.AMFRegionID = amfId[0]
		if i.AMFSetID, err = GetMaskedValue16(amfId[1:3], BIT16, BIT7); err != nil {
			return errors.Wrap(err, "FromGUTIStr: Get AMFSetID")
		}
		i.AMFPointer = Get6Bits61(amfId[2])
	}

	// TMSI
	if decodedTMSI, err := hex.DecodeString(s[6:14]); err != nil {
		return errors.Wrap(err, "FromGUTIStr: decode TMSI")
	} else {
		sl := i.TMSI5G[:]
		copy(sl, decodedTMSI)
	}
	return nil
}

/*************** Common function for IMEI / IMEISV ***************/

// getImeiImeisv14 decodes 14 digits of IMEI/IMEISV from BCD-encoded octets
// IMEI/IMEISV uses BCD encoding where each octet contains two digits
// The last octet may contain padding (0xF) for odd-length identifiers
func getImeiImeisv14(sl, b []byte) error {
	sl[0], _ = GetHalfValue(b[0])
	sl[2], sl[1] = GetHalfValue(b[1])
	sl[4], sl[3] = GetHalfValue(b[2])
	sl[6], sl[5] = GetHalfValue(b[3])
	sl[8], sl[7] = GetHalfValue(b[4])
	sl[10], sl[9] = GetHalfValue(b[5])
	sl[12], sl[11] = GetHalfValue(b[6])
	_, sl[13] = GetHalfValue(b[7])

	return nil
}

// setImeiImeisv14 encodes 14 digits of IMEI/IMEISV into BCD-encoded octets
// For odd-length IMEI, the last nibble is padded with 0xF as per 3GPP spec
func setImeiImeisv14(sl, b []byte) error {
	b[0] = Set4Bits85(b[0], sl[0])
	b[1] = SetHalfValue(sl[2], sl[1])
	b[2] = SetHalfValue(sl[4], sl[3])
	b[3] = SetHalfValue(sl[6], sl[5])
	b[4] = SetHalfValue(sl[8], sl[7])
	b[5] = SetHalfValue(sl[10], sl[9])
	b[6] = SetHalfValue(sl[12], sl[11])
	b[7] = Set4Bits41(sl[14], sl[13])
	return nil
}

/*************** IMEI ***************/

// unmarshalBinary_IMEI unmarshals IMEI according to 3GPP TS 24.501
// IMEI can be 14 or 15 digits, with odd/even indication in the first octet
// For odd-length IMEI, the last nibble is padded with 0xF
func (i *MobileId5GS) unmarshalBinary_IMEI(b []byte) error {
	var err error
	if len(b) != 8 {
		return errors.Errorf("unmarshalBinary_IMEI: incorrect IE length(%d)", len(b))
	}
	i.OddEvenIndic = GetBit4(b[0])

	sl := i.IMEI[:]
	if err = getImeiImeisv14(sl, b); err != nil {
		return errors.Wrap(err, "unmarshalBinary_IMEI get imei")
	}

	if OddNumOfIdDigit == i.OddEvenIndic {
		// Odd-length IMEI: extract the 15th digit from the last nibble
		sl[14] = Get4Bits85(b[7])
		// According to 3GPP TS 23.501: "shall" be filled with 1111 for even-length
		// For odd-length, this nibble contains the actual 15th digit
	} else {
		// Even-length IMEI: 15th digit position is unused
		sl[14] = 0
	}

	return nil
}

// marshalBinary_IMEI marshals IMEI according to 3GPP TS 24.501
// Handles both odd and even length IMEI with appropriate padding
func (i *MobileId5GS) marshalBinary_IMEI() ([]byte, error) {
	var err error
	b := make([]byte, 8)
	b[0] = Set3Bits31(b[0], i.TypeOfId)

	sl := i.IMEI[:]
	if err = setImeiImeisv14(sl, b); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_IMEI set imeisv")
	}
	if sl[14] == 0 {
		// Even-length IMEI: pad the last nibble with 0xF as per 3GPP spec
		b[7] = Set4Bits85(b[7], 0xf)
	} else {
		// Odd-length IMEI: set odd/even indicator and use the actual 15th digit
		b[0] = SetBit4(b[0], OddNumOfIdDigit)
		b[7] = Set4Bits85(b[7], sl[14])
	}
	return b, nil
}

// PEIStr formats the Permanent Equipment Identifier as a human-readable string
// For IMEI: "imei-" + 14 or 15 digits
// For IMEISV: "imeisv-" + 16 digits (including software version)
func (i *MobileId5GS) PEIStr() string {
	switch i.TypeOfId {
	case IdType_5GS_IMEI:
		s := i.IMEI[:]
		var str string
		if OddNumOfIdDigit == i.OddEvenIndic {
			// Odd-length IMEI: include all 15 digits
			str = fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d%d%d%d%d",
				s[0], s[1], s[2], s[3],
				s[4], s[5], s[6], s[7],
				s[8], s[9], s[10], s[11],
				s[12], s[13], s[14])
		} else {
			// Even-length IMEI: only 14 digits (15th position is padding)
			str = fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d%d%d%d",
				s[0], s[1], s[2], s[3],
				s[4], s[5], s[6], s[7],
				s[8], s[9], s[10], s[11],
				s[12], s[13])
		}
		return "imei-" + str

	case IdType_5GS_IMEISV:
		// IMEISV always has 16 digits including software version
		s := i.IMEISV[:]
		str := fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d%d%d%d%d%d",
			s[0], s[1], s[2], s[3],
			s[4], s[5], s[6], s[7],
			s[8], s[9], s[10], s[11],
			s[12], s[13], s[14], s[15])
		return "imeisv-" + str
	}
	return ""
}

/*************** IMEISV ***************/

func (i *MobileId5GS) unmarshalBinary_IMEISV(b []byte) error {
	var err error
	if len(b) != 9 {
		return errors.Errorf("unmarshalBinary_IMEISV: incorrect IE length(%d)", len(b))
	}

	sl := i.IMEISV[:]
	if err = getImeiImeisv14(sl, b); err != nil {
		return errors.Wrap(err, "unmarshalBinary_IMEISV()")
	}
	sl[14] = Get4Bits85(b[7])
	sl[15] = Get4Bits41(b[8])

	return nil
}

func (i *MobileId5GS) marshalBinary_IMEISV() ([]byte, error) {
	var err error
	b := make([]byte, 9)
	b[0] = Set3Bits31(b[0], i.TypeOfId)

	sl := i.IMEISV[:]
	if err = setImeiImeisv14(sl, b); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_IMEISV set imei")
	}
	b[7] = Set4Bits85(b[7], sl[14])
	b[8] = Set4Bits41(b[8], sl[15])
	b[8] = Set4Bits85(b[8], 0x0f)
	return b, nil
}

/*************** 5GS TSMI ***************/

func (i *MobileId5GS) unmarshalBinary_5GSTMSI(b []byte) error {
	var err error
	if len(b) != 7 {
		return errors.Errorf("unmarshalBinary_5GSTMSI: incorrect IE length(%d)", len(b))
	}

	i.AllOneBits = Get4Bits85(b[0])
	if i.AllOneBits != 0x0f {
		return errors.Errorf("unmarshalBinary_5GSTMSI: octet4 not start with 0xF0")
	}

	if i.AMFSetID, err = GetMaskedValue16(b[1:3], BIT16, BIT7); err != nil {
		return errors.Wrap(err, "unmarshalBinary_5GSTMSI get AMFSetID")
	}
	i.AMFPointer = Get6Bits61(b[2])
	sl := i.TMSI5G[:]
	copy(sl, b[3:7])
	return nil
}

func (i *MobileId5GS) marshalBinary_5GSTMSI() ([]byte, error) {
	b := make([]byte, 7)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	b[0] = Set4Bits85(b[0], 0x0F)
	if err := SetMaskedValue16(b[1:3], i.AMFSetID, BIT16, BIT7); err != nil {
		return nil, errors.Wrap(err, "marshalBinary_5GSTMSI set AMFSetID")
	}
	b[2] = Set6Bits61(b[2], i.AMFPointer)
	sl := i.TMSI5G[:]
	copy(b[3:7], sl)
	return b, nil
}

func (i *MobileId5GS) FiveGSTMSIStr() string {
	// <5G-S-TMSI> := <AMF Set ID><AMF Pointer><5G-TMSI>
	amfSetPtr := fmt.Sprintf("%04x", (i.AMFSetID<<6)|(uint16(i.AMFPointer)&0x3F))
	tmsi := hex.EncodeToString(i.TMSI5G[:])
	return amfSetPtr + tmsi
}

func (i *MobileId5GS) TMSIStr() string {
	return hex.EncodeToString(i.TMSI5G[:])
}

/*************** MAC Address ***************/

func (i *MobileId5GS) unmarshalBinary_MACAddr(b []byte) error {
	if len(b) != 7 {
		return errors.Errorf("unmarshalBinary_MACAddr: incorrect IE length(%d)", len(b))
	}

	i.MAURI = GetBit4(b[0])

	sl := i.MACAddr[:]
	copy(sl, b[1:7])

	return nil
}

func (i *MobileId5GS) marshalBinary_MACAddr() ([]byte, error) {
	b := make([]byte, 7)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	b[0] = SetBit4(b[0], i.MAURI)
	sl := i.MACAddr[:]
	copy(b[1:7], sl)
	return b, nil
}

func (i *MobileId5GS) MACAddrStr() string {
	return hex.EncodeToString(i.MACAddr[:])
}

/*************** EUI64 ***************/

func (i *MobileId5GS) unmarshalBinary_EUI64(b []byte) error {
	if len(b) != 9 {
		return errors.Errorf("unmarshalBinary_EUI64: incorrect IE length(%d)", len(b))
	}
	slEUI64 := i.EUI64[:]
	copy(slEUI64, b[1:9])
	return nil
}

func (i *MobileId5GS) marshalBinary_EUI64() ([]byte, error) {
	b := make([]byte, 9)
	b[0] = Set3Bits31(b[0], i.TypeOfId)
	sl := i.EUI64[:]
	copy(b[1:9], sl)
	return b, nil
}

func (i *MobileId5GS) EUI64Str() string {
	return hex.EncodeToString(i.EUI64[:])
}
