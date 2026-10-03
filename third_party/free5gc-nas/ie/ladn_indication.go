package ie

import "github.com/pkg/errors"

// LADNInd is detailed in 9.11.3.29 LADN indication, 24.501
type LADNInd struct {
	// Name, uint8, Bits, Octet
	DNNs []DNN
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *LADNInd) UnmarshalBinary(b []byte) error {
	ofs := uint16(0) // offset
	for int(ofs) < len(b) {
		dnn := DNN{}
		dnnLen := b[ofs]
		ofs++
		if ofs+uint16(dnnLen) > uint16(len(b)) {
			i.DNNs = nil
			return errors.Errorf("Bad LADN DNN value, ofs=%d, dnnLen=%d, len(b)=%d", ofs, dnnLen, len(b))
		}
		if err := dnn.UnmarshalBinary(b[ofs : ofs+uint16(dnnLen)]); err != nil {
			return errors.Wrap(err, "LADNInd UnmarshalBinary()")
		}
		i.DNNs = append(i.DNNs, dnn)
		ofs += uint16(dnnLen)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *LADNInd) MarshalBinary() ([]byte, error) {
	var b []byte
	for j := range i.DNNs {
		mb, err := i.DNNs[j].MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "LADNInd MarshalBinary()")
		}
		b = append(b, uint8(len(mb)))
		b = append(b, mb...)
	}
	return b, nil
}
