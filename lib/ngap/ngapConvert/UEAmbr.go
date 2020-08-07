package ngapConvert

import (
	"strconv"
	"strings"
)

func UEAmbrToInt64(modelAmbr string) (ambrInt int64) {
	tok := strings.Split(modelAmbr, " ")
	ambr, _ := strconv.ParseFloat(tok[0], 64)
	ambrInt = int64(ambr * getUnit(tok[1]))
	return
}

func getUnit(unit string) float64 {
	switch unit {
	case "bps":
		return 1.0
	case "Kbps":
		return 1000.0
	case "Mbps":
		return 1000000.0
	case "Gbps":
		return 1000000000.0
	case "Tbps":
		return 1000000000000.0
	}
	return 1.0
}
