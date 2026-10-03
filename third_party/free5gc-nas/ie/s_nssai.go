package ie

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

// SNSSAI is detailed in 9.11.2.8 S-NSSAI, 24.501
type SNSSAI struct {
	// Name, uint8, Bits, Octet
	SST            uint8  // 8 -> 1 ,   3 -> 3
	MappedHPLMNSST uint8  // 8 -> 1 ,   7 -> 7
	SD             string // 8 -> 1 ,   4 -> 6, HEX String -> [3]byte
	MappedHPLMNSD  string // 8 -> 1 ,   8 -> 10 HEX String -> [3]byte
}

type ConstSNSSAILen uint16

const (
	LenSST                                 ConstSNSSAILen = 0x01
	LenSST_MappedHPLMNSST                  ConstSNSSAILen = 0x02
	LenSST_SD                              ConstSNSSAILen = 0x04
	LenSST_SD_MappedHPLMNSST               ConstSNSSAILen = 0x05
	LenSST_SD_MappedHPLMNSST_MappedHPLMNSD ConstSNSSAILen = 0x08

	// Table 5.15.2.2-1, 23.501
	SST_EMBB  uint8 = 1
	SST_URLLC uint8 = 2
	SST_MIoT  uint8 = 3
	SST_V2X   uint8 = 4

	NoSDVal string = "ffffff"
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SNSSAI) UnmarshalBinary(b []byte) error {
	switch ConstSNSSAILen(len(b)) {
	case LenSST:
		i.SST = b[0]
	case LenSST_MappedHPLMNSST:
		i.SST = b[0]
		i.MappedHPLMNSST = b[1]
	case LenSST_SD:
		i.SST = b[0]
		i.SD = hex.EncodeToString(b[1:4])
	case LenSST_SD_MappedHPLMNSST:
		i.SST = b[0]
		i.SD = hex.EncodeToString(b[1:4])
		i.MappedHPLMNSST = b[4]
	case LenSST_SD_MappedHPLMNSST_MappedHPLMNSD:
		i.SST = b[0]
		i.SD = hex.EncodeToString(b[1:4])
		i.MappedHPLMNSST = b[4]
		i.MappedHPLMNSD = hex.EncodeToString(b[5:8])
	default:
		return errors.Errorf("The SNSSAI IE length(%d) is incorrect", len(b))
	}
	return nil
}

func isEmptySD(sd string) bool {
	return sd == "" || strings.EqualFold(sd, NoSDVal)
}

func isEmptySST(sst uint8) bool {
	return sst == 0
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SNSSAI) MarshalBinary() ([]byte, error) {
	b := make([]byte, LenSST_SD_MappedHPLMNSST_MappedHPLMNSD)
	switch {
	// SST only
	case !isEmptySST(i.SST) && isEmptySST(i.MappedHPLMNSST) && isEmptySD(i.SD) && isEmptySD(i.MappedHPLMNSD):
		b[0] = i.SST
		b = b[0:LenSST]
	// SST + MappedHPLMNSST
	case !isEmptySST(i.SST) && !isEmptySST(i.MappedHPLMNSST) && isEmptySD(i.SD) && isEmptySD(i.MappedHPLMNSD):
		b[0] = i.SST
		b[1] = i.MappedHPLMNSST
		b = b[0:LenSST_MappedHPLMNSST]
	// SST + SD
	case !isEmptySST(i.SST) && isEmptySST(i.MappedHPLMNSST) && !isEmptySD(i.SD) && isEmptySD(i.MappedHPLMNSD):
		b[0] = i.SST
		if _, err := hex.Decode(b[1:4], []byte(i.SD)); err != nil {
			return nil, errors.Wrap(err, "SNSSAI MarshalBinary() SD")
		}
		b = b[0:LenSST_SD]
	// SST + SD + MappedHPLMNSST
	case !isEmptySST(i.SST) && !isEmptySST(i.MappedHPLMNSST) && !isEmptySD(i.SD) && isEmptySD(i.MappedHPLMNSD):
		b[0] = i.SST
		if _, err := hex.Decode(b[1:4], []byte(i.SD)); err != nil {
			return nil, errors.Wrap(err, "SNSSAI MarshalBinary() SD")
		}
		b[4] = i.MappedHPLMNSST
		b = b[0:LenSST_SD_MappedHPLMNSST]
	// SST + SD + MappedHPLMNSST + MappedHPLMNSD
	case !isEmptySST(i.SST) && !isEmptySST(i.MappedHPLMNSST) && !isEmptySD(i.SD) && !isEmptySD(i.MappedHPLMNSD):
		b[0] = i.SST
		if _, err := hex.Decode(b[1:4], []byte(i.SD)); err != nil {
			return nil, errors.Wrap(err, "SNSSAI MarshalBinary() SD")
		}
		b[4] = i.MappedHPLMNSST
		if _, err := hex.Decode(b[5:8], []byte(i.MappedHPLMNSD)); err != nil {
			return nil, errors.Wrap(err, "SNSSAI MarshalBinary() MappedHPLMNSD")
		}
		b = b[0:LenSST_SD_MappedHPLMNSST_MappedHPLMNSD]
	default:
		return nil, errors.Errorf("the SNSSAI content combination not available")
	}

	return b, nil
}
