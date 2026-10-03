package ie

import (
	"github.com/pkg/errors"
)

// LADNInfo is detailed in 9.11.3.30 LADN information, 24.501
type LADNInfo struct {
	DnnTai []DNN_TAI
}

type DNN_TAI struct {
	DNN
	TrackingAreaIdList5GS
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *LADNInfo) UnmarshalBinary(b []byte) error {
	ttlLen := uint8(len(b))
	ofs := uint8(0)

	for ofs < ttlLen {
		var tmpDT DNN_TAI

		if dnnLen := b[ofs]; dnnLen == 0 || (ofs+1+dnnLen) > ttlLen {
			return errors.Errorf("LADNInfo UnmarshalBinary: ofs=%d, dnnLen=%d, ttlLen=%d",
				ofs, dnnLen, ttlLen)
		} else {
			ofs += 1
			if err := tmpDT.DNN.UnmarshalBinary(b[ofs : ofs+dnnLen]); err != nil {
				return errors.Wrap(err, "LADNInfo DNN UnmarshalBinary")
			}
			ofs += dnnLen
		}

		if taiLen := b[ofs]; taiLen == 0 || (ofs+1+taiLen) > ttlLen {
			return errors.Errorf("LADNInfo UnmarshalBinary: ofs=%d, taiLen=%d, ttlLen=%d",
				ofs, taiLen, ttlLen)
		} else {
			ofs += 1
			if err := tmpDT.TrackingAreaIdList5GS.UnmarshalBinary(b[ofs : ofs+taiLen]); err != nil {
				return errors.Wrap(err, "LADNInfo TAID List UnmarshalBinary")
			}
			ofs += taiLen
		}

		i.DnnTai = append(i.DnnTai, tmpDT)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *LADNInfo) MarshalBinary() ([]byte, error) {
	var b []byte
	for _, dnnTai := range i.DnnTai {
		var bDnn, bTai []byte
		var err error

		if bDnn, err = dnnTai.DNN.MarshalBinary(); err != nil {
			return nil, errors.Wrap(err, "LADNInfo DNN MarshalBinary")
		}

		if bTai, err = dnnTai.TrackingAreaIdList5GS.MarshalBinary(); err != nil {
			return nil, errors.Wrap(err, "LADNInfo TAID List MarshalBinary")
		}
		b = append(b, uint8(len(bDnn)))
		b = append(b, bDnn...)

		b = append(b, uint8(len(bTai)))
		b = append(b, bTai...)
	}
	return b, nil
}
