package nasConvert

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/openapi/models"
)

func RequestedNssaiToModels(nasNssai *nasType.RequestedNSSAI) (nssai []models.Snssai) {

	buf := nasNssai.GetSNSSAIValue()
	lengthOfBuf := int(nasNssai.GetLen())
	offset := 0
	for offset < lengthOfBuf {
		snssaiValue := buf[offset:]
		snssai, readLength := requestedSnssaiToModels(snssaiValue)
		nssai = append(nssai, snssai)
		offset += readLength
	}

	return

}

func requestedSnssaiToModels(buf []byte) (snssai models.Snssai, length int) {

	lengthOfSnssaiContents := buf[0]
	switch lengthOfSnssaiContents {
	case 0x01: // sst
		snssai.Sst = int32(buf[1])
		length = 2
	case 0x04: // sst + sd
		snssai.Sst = int32(buf[1])
		snssai.Sd = hex.EncodeToString(buf[2:5])
		length = 5
	default:
		fmt.Printf("Not Supported length: %d\n", lengthOfSnssaiContents)
	}

	return
}

func RejectedNssaiToNas(rejectedNssaiInPlmn []models.Snssai, rejectedNssaiInTa []models.Snssai) (rejectedNssaiNas nasType.RejectedNSSAI) {

	var byteArray []uint8
	for _, rejectedSnssai := range rejectedNssaiInPlmn {
		byteArray = append(byteArray, RejectedSnssaiToNas(rejectedSnssai, nasMessage.RejectedSnssaiCauseNotAvailableInCurrentPlmn)...)
	}
	for _, rejectedSnssai := range rejectedNssaiInTa {
		byteArray = append(byteArray, RejectedSnssaiToNas(rejectedSnssai, nasMessage.RejectedSnssaiCauseNotAvailableInCurrentRegistrationArea)...)
	}

	rejectedNssaiNas.SetLen(uint8(len(byteArray)))
	rejectedNssaiNas.SetRejectedNSSAIContents(byteArray)
	return
}
