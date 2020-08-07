package nasConvert

import (
	"encoding/hex"
	"fmt"
	"math/bits"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/openapi/models"
	"strconv"
	"strings"
)

func GetTypeOfIdentity(buf byte) uint8 {
	return buf & 0x07
}

// TS 24.501 9.11.3.4
// suci(imsi) = "suci-0-${mcc}-${mnc}-${routingIndentifier}-${protectionScheme}-${homeNetworkPublicKeyIdentifier}-${schemeOutput}"
// suci(nai) = "nai-${naiString}"
func SuciToString(buf []byte) (suci string, plmnId string) {

	var mcc, mnc, routingInd, protectionScheme, homeNetworkPublicKeyIdentifier, schemeOutput string

	supiFormat := (buf[0] & 0xf0) >> 4
	if supiFormat == nasMessage.SupiFormatNai {
		suci = NaiToString(buf)
		return
	}

	// Encode buf to SUCI in supi format "IMSI"

	// Plmn(MCC + MNC)
	mccDigit3 := (buf[2] & 0x0f)
	tmpBytes := []byte{bits.RotateLeft8(buf[1], 4), (mccDigit3 << 4)}
	mcc = hex.EncodeToString(tmpBytes)
	mcc = mcc[:3] // remove rear 0

	mncDigit3 := (buf[2] & 0xf0) >> 4
	tmpBytes = []byte{bits.RotateLeft8(buf[3], 4), mncDigit3 << 4}
	mnc = hex.EncodeToString(tmpBytes)
	if mnc[2] == 'f' {
		mnc = mnc[:2] // mnc is 2 digit -> remove 'f'
	} else {
		mnc = mnc[:3] // mnc is 3 digit -> remove rear 0
	}
	plmnId = mcc + mnc

	// Routing Indicator
	var routingIndBytes []byte
	routingIndBytes = append(routingIndBytes, bits.RotateLeft8(buf[4], 4))
	routingIndBytes = append(routingIndBytes, bits.RotateLeft8(buf[5], 4))
	routingInd = hex.EncodeToString(routingIndBytes)

	if idx := strings.Index(routingInd, "f"); idx != -1 {
		routingInd = routingInd[0:idx]
	}

	// Protection Scheme
	protectionScheme = fmt.Sprintf("%x", buf[6]) // convert byte to hex string without leading 0s

	// Home Network Public Key Indentifier
	homeNetworkPublicKeyIdentifier = fmt.Sprintf("%d", buf[7])

	// Scheme output
	if protectionScheme == strconv.Itoa(nasMessage.ProtectionSchemeNullScheme) {
		// MSIN
		var msinBytes []byte
		for i := 8; i < len(buf); i++ {
			msinBytes = append(msinBytes, bits.RotateLeft8(buf[i], 4))
		}
		schemeOutput = hex.EncodeToString(msinBytes)
		if schemeOutput[len(schemeOutput)-1] == 'f' {
			schemeOutput = schemeOutput[:len(schemeOutput)-1]
		}
	} else {
		schemeOutput = hex.EncodeToString(buf[8:])
	}

	suci = strings.Join([]string{"suci", "0", mcc, mnc, routingInd, protectionScheme, homeNetworkPublicKeyIdentifier, schemeOutput}, "-")
	return
}

func NaiToString(buf []byte) (nai string) {
	prefix := "nai"
	naiBytes := buf[1:]
	naiStr := hex.EncodeToString(naiBytes)
	nai = strings.Join([]string{prefix, "1", naiStr}, "-")
	return
}

// nasType: TS 24.501 9.11.3.4
func GutiToString(buf []byte) (guami models.Guami, guti string) {

	plmnID := PlmnIDToString(buf[1:4])
	amfID := hex.EncodeToString(buf[4:7])
	tmsi5G := hex.EncodeToString(buf[7:])

	guami.PlmnId = new(models.PlmnId)
	guami.PlmnId.Mcc = plmnID[:3]
	guami.PlmnId.Mnc = plmnID[3:]
	guami.AmfId = amfID
	guti = plmnID + amfID + tmsi5G
	return
}

func GutiToNas(guti string) (gutiNas nasType.GUTI5G) {

	gutiNas.SetLen(11)
	gutiNas.SetSpare(0)
	gutiNas.SetTypeOfIdentity(nasMessage.MobileIdentity5GSType5gGuti)

	mcc1, _ := strconv.Atoi(string(guti[0]))
	mcc2, _ := strconv.Atoi(string(guti[1]))
	mcc3, _ := strconv.Atoi(string(guti[2]))
	mnc1, _ := strconv.Atoi(string(guti[3]))
	mnc2, _ := strconv.Atoi(string(guti[4]))
	mnc3 := 0x0f
	amfId := ""
	tmsi := ""
	if len(guti) == 20 {
		mnc3, _ = strconv.Atoi(string(guti[5]))
		amfId = guti[6:12]
		tmsi = guti[12:]
	} else {
		amfId = guti[5:11]
		tmsi = guti[11:]
	}
	gutiNas.SetMCCDigit1(uint8(mcc1))
	gutiNas.SetMCCDigit2(uint8(mcc2))
	gutiNas.SetMCCDigit3(uint8(mcc3))
	gutiNas.SetMNCDigit1(uint8(mnc1))
	gutiNas.SetMNCDigit2(uint8(mnc2))
	gutiNas.SetMNCDigit3(uint8(mnc3))

	amfRegionId, amfSetId, amfPointer := AmfIdToNas(amfId)
	gutiNas.SetAMFRegionID(amfRegionId)
	gutiNas.SetAMFSetID(amfSetId)
	gutiNas.SetAMFPointer(amfPointer)
	tmsiBytes, _ := hex.DecodeString(tmsi)
	copy(gutiNas.Octet[7:11], tmsiBytes[:])
	return
}

// PEI: ^(imei-[0-9]{15}|imeisv-[0-9]{16}|.+)$
func PeiToString(buf []byte) (pei string) {

	var prefix string

	typeOfIdentity := buf[0] & 0x07
	if typeOfIdentity == 0x03 {
		prefix = "imei-"
	} else {
		prefix = "imeisv-"
	}

	oddIndication := (buf[0] & 0x08) >> 3

	digit1 := (buf[0] & 0xf0)

	tmpBytes := []byte{digit1}

	for _, octet := range buf[1:] {
		digitP := octet & 0x0f
		digitP1 := octet & 0xf0

		tmpBytes[len(tmpBytes)-1] += digitP
		tmpBytes = append(tmpBytes, digitP1)
	}

	digitStr := hex.EncodeToString(tmpBytes)
	digitStr = digitStr[:len(digitStr)-1] // remove the last digit

	if oddIndication == 0 { // even digits
		digitStr = digitStr[:len(digitStr)-1] // remove the last digit
	}

	pei = prefix + digitStr
	return
}
