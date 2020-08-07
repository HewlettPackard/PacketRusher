package nasConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/nas/nasType"
	"my5G-RANTester/lib/openapi/models"
)

func SnssaiToModels(nasSnssai *nasType.SNSSAI) (snssai models.Snssai) {
	sD := nasSnssai.GetSD()
	snssai.Sd = hex.EncodeToString([]uint8(sD[:]))
	snssai.Sst = int32(nasSnssai.GetSST())
	return
}

func SnssaiToNas(snssai models.Snssai) (buf []uint8) {
	if snssai.Sd == "" {
		buf = append(buf, 0x01)
		buf = append(buf, uint8(snssai.Sst))
	} else {
		buf = append(buf, 0x04)
		buf = append(buf, uint8(snssai.Sst))
		byteArray, _ := hex.DecodeString(snssai.Sd)
		buf = append(buf, byteArray...)
	}
	return
}

func RejectedSnssaiToNas(snssai models.Snssai, rejectCause uint8) (rejectedSnssai []uint8) {

	if snssai.Sd == "" {
		rejectedSnssai = append(rejectedSnssai, (0x01<<4)+rejectCause)
		rejectedSnssai = append(rejectedSnssai, uint8(snssai.Sst))
	} else {
		rejectedSnssai = append(rejectedSnssai, (0x04<<4)+rejectCause)
		rejectedSnssai = append(rejectedSnssai, uint8(snssai.Sst))
		sDBytes, _ := hex.DecodeString(snssai.Sd)
		rejectedSnssai = append(rejectedSnssai, sDBytes...)
	}

	return
}
