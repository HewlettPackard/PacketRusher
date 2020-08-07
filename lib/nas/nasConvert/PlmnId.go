package nasConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/openapi/models"
	"strconv"
)

func PlmnIDToNas(plmnID models.PlmnId) (plmnNas []uint8) {

	mccDigit1, _ := strconv.Atoi(string(plmnID.Mcc[0]))
	mccDigit2, _ := strconv.Atoi(string(plmnID.Mcc[1]))
	mccDigit3, _ := strconv.Atoi(string(plmnID.Mcc[2]))

	mncDigit1, _ := strconv.Atoi(string(plmnID.Mnc[0]))
	mncDigit2, _ := strconv.Atoi(string(plmnID.Mnc[1]))
	mncDigit3 := 0x0f
	if len(plmnID.Mnc) == 3 {
		mncDigit3, _ = strconv.Atoi(string(plmnID.Mnc[2]))
	}

	plmnNas = []uint8{
		uint8((mccDigit2 << 4) | mccDigit1),
		uint8((mncDigit3 << 4) | mccDigit3),
		uint8((mncDigit2 << 4) | mncDigit1),
	}

	return
}

func PlmnIDToString(nasBuf []byte) (plmnID string) {

	mccDigit1 := nasBuf[0] & 0x0f
	mccDigit2 := (nasBuf[0] & 0xf0) >> 4
	mccDigit3 := (nasBuf[1] & 0x0f)

	mncDigit1 := (nasBuf[2] & 0x0f)
	mncDigit2 := (nasBuf[2] & 0xf0) >> 4
	mncDigit3 := (nasBuf[1] & 0xf0) >> 4

	tmpBytes := []byte{(mccDigit1 << 4) | mccDigit2, (mccDigit3 << 4) | mncDigit1, (mncDigit2 << 4) | mncDigit3}

	plmnID = hex.EncodeToString(tmpBytes)
	if plmnID[5] == 'f' {
		plmnID = plmnID[:5] // get plmnID[0~4]
	}
	return
}
