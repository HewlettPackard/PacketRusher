package nasConvert

import (
	"encoding/hex"
)

func AmfIdToNas(amfId string) (amfRegionId uint8, amfSetId uint16, amfPointer uint8) {

	amfIdBytes, _ := hex.DecodeString(amfId)

	amfRegionId = uint8(amfIdBytes[0])
	amfSetId = uint16(amfIdBytes[1])<<2 + (uint16(amfIdBytes[2])&0x00c0)>>6
	amfPointer = uint8(amfIdBytes[2]) & 0x3f
	return
}

func AmfIdToModels(amfRegionId uint8, amfSetId uint16, amfPointer uint8) (amfId string) {

	tmpBytes := []uint8{amfRegionId, uint8(amfSetId>>2) & 0xff, uint8(amfSetId&0x03) + amfPointer&0x3f}
	amfId = hex.EncodeToString(tmpBytes)
	return
}
