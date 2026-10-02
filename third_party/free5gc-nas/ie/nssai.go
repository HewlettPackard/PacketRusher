package ie

import (
	"bytes"

	"github.com/pkg/errors"
)

// NSSAI is detailed in 9.11.3.37 NSSAI, 24.501
type NSSAI struct {
	// Name, uint8, Bits, Octet
	SNSSAIs []SNSSAI
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NSSAI) UnmarshalBinary(b []byte) error {
	ofs := uint16(0) // offset
	for int(ofs) < len(b) {
		snssai := SNSSAI{}
		snssaiLen := b[ofs]
		ofs++
		if ofs+uint16(snssaiLen) > uint16(len(b)) {
			i.SNSSAIs = nil
			return errors.Errorf("Bad SNSSAI value, ofs=%d, snssaiLen=%d, len(b)=%d",
				ofs, snssaiLen, len(b))
		}
		if err := snssai.UnmarshalBinary(b[ofs : ofs+uint16(snssaiLen)]); err != nil {
			return errors.Wrap(err, "NSSAI snssai UnmarshalBinary")
		}
		i.SNSSAIs = append(i.SNSSAIs, snssai)
		ofs += uint16(snssaiLen)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NSSAI) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	var n int

	for idx, snssai := range i.SNSSAIs {
		mb, err := snssai.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "NSSAI snssais MarshalBinary")
		}
		err = buf.WriteByte(byte(len(mb)))
		if err != nil {
			return nil, errors.Wrapf(err, "encode length of SNSSAI[%d]", idx)
		}
		n, err = buf.Write(mb)
		if err != nil || n != len(mb) {
			return nil, errors.Wrapf(err, "encode SNSSAI[%d]", idx)
		}
	}
	return buf.Bytes(), nil
}
