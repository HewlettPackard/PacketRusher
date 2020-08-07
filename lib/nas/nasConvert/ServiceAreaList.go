package nasConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/openapi/models"
)

// TS 24.501 9.11.3.49
func PartialServiceAreaListToNas(plmnID models.PlmnId, serviceAreaRestriction models.ServiceAreaRestriction) (partialServiceAreaList []byte) {

	var allowedType uint8

	if serviceAreaRestriction.RestrictionType == models.RestrictionType_ALLOWED_AREAS {
		allowedType = nasMessage.AllowedTypeAllowedArea
	} else {
		allowedType = nasMessage.AllowedTypeNonAllowedArea
	}

	numOfElements := uint8(len(serviceAreaRestriction.Areas))

	firstByte := (allowedType<<7)&0x80 + numOfElements // only support TypeOfList '00' now
	plmnIDNas := PlmnIDToNas(plmnID)

	partialServiceAreaList = append(partialServiceAreaList, firstByte)
	partialServiceAreaList = append(partialServiceAreaList, plmnIDNas...)

	for _, area := range serviceAreaRestriction.Areas {
		for _, tac := range area.Tacs {
			tacBytes, _ := hex.DecodeString(tac)
			partialServiceAreaList = append(partialServiceAreaList, tacBytes...)
		}
	}
	return
}
