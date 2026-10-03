package ie

import (
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// TrackingAreaIdList5GS is detailed in 9.11.3.9 5GS tracking area identity list, 24.501
type TrackingAreaIdList5GS struct {
	TAI []TrackingAreaId5GS
}

const (
	ConstSamePlmnIdNonConsecTac uint8 = 0
	ConstSamePlmnIdConsecTac    uint8 = 1
	ConstDiffPlmnId             uint8 = 2

	ConstMaxTAI uint8 = 16
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *TrackingAreaIdList5GS) UnmarshalBinary(b []byte) error {
	var err error
	var umBytes uint8
	ofs := uint8(0) // Offset
	if len(b) < 7 {
		return errors.Errorf("TrackingAreaIdList5GS: bad len (%d) to unmarshal", len(b))
	}

	i.TAI = make([]TrackingAreaId5GS, 0, ConstMaxTAI)

	for int(ofs) < len(b) && len(i.TAI) < int(ConstMaxTAI) {
		listType := Get2Bits76(b[ofs])
		switch listType {
		case ConstSamePlmnIdNonConsecTac:
			umBytes, err = i.unmarshalSamePlmnId_NonConsecTac(b[ofs:])
		case ConstSamePlmnIdConsecTac:
			umBytes, err = i.unmarshalSamePlmnId_ConsecTac(b[ofs:])
		case ConstDiffPlmnId:
			umBytes, err = i.unmarshalDiffPlmnId(b[ofs:])
		default:
			return errors.Errorf("TrackingAreaIdList5GS: unknown listType %d", listType)
		}
		if err != nil {
			return err
		}
		ofs += umBytes
	}
	return nil
}

func (i *TrackingAreaIdList5GS) unmarshalSamePlmnId_NonConsecTac(b []byte) (uint8, error) {
	nrOfElem := Get5Bits51(b[0]) + 1
	if nrOfElem > 16 {
		nrOfElem = 16
	}
	umBytes := 4 + 3*nrOfElem
	if len(b) < int(umBytes) {
		return 0, errors.Errorf("TrackingAreaIdList5GS-NonConsecTac: bad len (%d) to unmarshal, nrOfElem=%d, umB=%d",
			len(b), nrOfElem, umBytes)
	}
	tai := TrackingAreaId5GS{}
	if err := tai.UnmarshalPlmnId(b[1:4]); err != nil {
		return 0, errors.Wrap(err, "unmarshalSamePlmnId_NonConsecTac plmnId")
	}

	for j := uint8(0); j < nrOfElem; j++ {
		ofs := 3*j + 4
		if err := tai.UnmarshalTac(b[ofs : ofs+3]); err != nil {
			return 0, errors.Wrap(err, "unmarshalSamePlmnId_NonConsecTac tac")
		}
		i.TAI = append(i.TAI, tai)
	}
	return umBytes, nil
}

func (i *TrackingAreaIdList5GS) unmarshalSamePlmnId_ConsecTac(b []byte) (uint8, error) {
	nrOfElem := Get5Bits51(b[0]) + 1
	if nrOfElem > 16 {
		nrOfElem = 16
	}
	umBytes := uint8(7)
	if len(b) < int(umBytes) {
		return 0, errors.Errorf("TrackingAreaIdList5GS-ConsecTac: bad len (%d) to unmarshal, nrOfElem=%d",
			len(b), nrOfElem)
	}
	tai := TrackingAreaId5GS{}
	if err := tai.UnmarshalPlmnId(b[1:4]); err != nil {
		return 0, errors.Wrap(err, "unmarshalSamePlmnId_ConsecTac plmnId")
	}

	tmp32 := uint32(b[4])<<16 + uint32(b[5])<<8 + uint32(b[6])

	for j := uint8(0); j < nrOfElem; j++ {
		tmpTac := []byte{0, 0, 0}
		tmpTac[0] = uint8((tmp32 & 0x00ff0000) >> 16)
		tmpTac[1] = uint8((tmp32 & 0x0000ff00) >> 8)
		tmpTac[2] = uint8((tmp32 & 0x000000ff))
		if err := tai.UnmarshalTac(tmpTac); err != nil {
			return 0, errors.Wrap(err, "unmarshalSamePlmnId_ConsecTac tac")
		}
		i.TAI = append(i.TAI, tai)
		tmp32++
	}
	return umBytes, nil
}

func (i *TrackingAreaIdList5GS) unmarshalDiffPlmnId(b []byte) (uint8, error) {
	nrOfElem := Get5Bits51(b[0]) + 1
	if nrOfElem > 16 {
		nrOfElem = 16
	}
	umBytes := 1 + 6*nrOfElem
	if len(b) < int(umBytes) {
		return 0, errors.Errorf("TrackingAreaIdList5GS-NonConsecTac: bad len (%d) to unmarshal, nrOfElem=%d",
			len(b), nrOfElem)
	}
	tai := TrackingAreaId5GS{}
	for j := uint8(0); j < nrOfElem; j++ {
		ofs := 6*j + 1
		if err := tai.UnmarshalPlmnId(b[ofs : ofs+3]); err != nil {
			return 0, errors.Wrap(err, "unmarshalDiffPlmnId plmnId")
		}
		if err := tai.UnmarshalTac(b[ofs+3 : ofs+6]); err != nil {
			return 0, errors.Wrap(err, "unmarshalDiffPlmnId tac")
		}
		i.TAI = append(i.TAI, tai)
	}
	return umBytes, nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *TrackingAreaIdList5GS) MarshalBinary() ([]byte, error) {
	nrOfTai := len(i.TAI)
	if nrOfTai > 16 {
		nrOfTai = 16
	}
	mBytes := uint8(nrOfTai*(6) + 1)
	b := make([]byte, mBytes)

	listType := ConstSamePlmnIdNonConsecTac
	if nrOfTai > 1 {
		for j := 1; j < nrOfTai; j++ {
			if !reflect.DeepEqual(i.TAI[0].PlmnId, i.TAI[j].PlmnId) {
				listType = ConstDiffPlmnId
				break
			}
		}
		if listType != ConstDiffPlmnId {
			// All have same PLMN; check if consecutive TAC
			consec := true
			base, err1 := strconv.ParseInt(i.TAI[0].TAC, 16, 64)
			if err1 != nil {
				return nil, errors.Errorf("TrackingAreaIdList5GS: invalid TAC=%s", i.TAI[0].TAC)
			}
			for j := 1; j < nrOfTai; j++ {
				inc, err2 := strconv.ParseInt(i.TAI[j].TAC, 16, 64)
				if err2 != nil {
					return nil, errors.Errorf("TrackingAreaIdList5GS: invalid TAC=%s", i.TAI[j].TAC)
				}
				if base+1 != inc {
					consec = false
					break
				}
				base = inc
			}
			if consec {
				listType = ConstSamePlmnIdConsecTac
			}
		}
	}
	b[0] = Set2Bits76(b[0], listType)

	b[0] = Set5Bits51(b[0], uint8(nrOfTai)-1) // 0: 1 element, 15: 16 elements
	for j := 0; j < nrOfTai; j++ {
		ofs := 6*j + 1
		if err := i.TAI[j].MarshalPlmnId(b[ofs : ofs+3]); err != nil {
			return nil, errors.Wrap(err, "unmarshalDiffPlmnId plmnId")
		}
		if err := i.TAI[j].MarshalTac(b[ofs+3 : ofs+6]); err != nil {
			return nil, errors.Wrap(err, "unmarshalDiffPlmnId tac")
		}
	}
	return b, nil
}
