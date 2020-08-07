package nasConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/openapi/models"
	"reflect"
)

// TS 24.501 9.11.3.9
func TaiListToNas(taiList []models.Tai) (taiListNas []uint8) {

	typeOfList := 0x00

	plmnId := taiList[0].PlmnId
	for _, tai := range taiList {
		if !reflect.DeepEqual(plmnId, tai.PlmnId) {
			typeOfList = 0x02
		}
	}

	numOfElementsNas := uint8(len(taiList)) - 1

	taiListNas = append(taiListNas, uint8(typeOfList<<5)+numOfElementsNas)

	switch typeOfList {
	case 0x00:
		plmnNas := PlmnIDToNas(*plmnId)
		taiListNas = append(taiListNas, plmnNas...)

		for _, tai := range taiList {
			tacBytes, _ := hex.DecodeString(tai.Tac)
			taiListNas = append(taiListNas, tacBytes...)
		}
	case 0x02:
		for _, tai := range taiList {
			plmnNas := PlmnIDToNas(*tai.PlmnId)
			tacBytes, _ := hex.DecodeString(tai.Tac)
			taiListNas = append(taiListNas, plmnNas...)
			taiListNas = append(taiListNas, tacBytes...)
		}
	}

	return
}
