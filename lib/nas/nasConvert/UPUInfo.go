package nasConvert

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/lib/openapi/models"
)

//  subclause 9.11.3.53A in 3GPP TS 24.501
func UpuInfoToNas(upuInfo models.UpuInfo) (buf []uint8) {
	// set upu Header
	buf = append(buf, upuInfoGetHeader(upuInfo.UpuRegInd, upuInfo.UpuAckInd))
	// Set UPU-MAC-IAUSF
	byteArray, _ := hex.DecodeString(upuInfo.UpuMacIausf)
	buf = append(buf, byteArray...)
	// Set Counter UPU
	byteArray, _ = hex.DecodeString(upuInfo.CounterUpu)
	buf = append(buf, byteArray...)
	// Set UE parameters update list
	for _, data := range upuInfo.UpuDataList {
		if data.SecPacket != "" {
			buf = append(buf, 0x01)
			byteArray, _ = hex.DecodeString(data.SecPacket)
		} else {
			buf = append(buf, 0x02)
			byteArray = []byte{}
			for _, snssai := range data.DefaultConfNssai {
				snssaiData := SnssaiToNas(snssai)
				byteArray = append(byteArray, snssaiData...)
			}
		}
		buf = append(buf, uint8(len(byteArray)))
		buf = append(buf, byteArray...)
	}
	return
}

func upuInfoGetHeader(reg bool, ack bool) (buf uint8) {
	var regValue, ackValue uint8
	if reg {
		regValue = 1
	}
	if ack {
		ackValue = 1
	}
	buf = regValue<<2 + ackValue<<1
	return
}

func UpuAckToModels(buf []uint8) (string, error) {
	if (buf[0] != 0x01) || (len(buf) != 17) {
		return "", fmt.Errorf("NAS UPU Ack is not valid")
	}
	return hex.EncodeToString(buf[1:]), nil
}
