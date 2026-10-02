package ie

import (
	"github.com/pkg/errors"
)

type IpAddrType uint8

const (
	IPv4   IpAddrType = 0x01
	IPv6   IpAddrType = 0x02
	IPv4v6 IpAddrType = 0x03

	SI6LLA_Absent  uint8 = 0
	SI6LLA_Present uint8 = 1
)

// PDUAddr is detailed in 9.11.4.10 PDU address, 24.501
type PDUAddr struct {
	// Name, uint8, Bits, Octet
	IPv4       []byte
	IPv6IfId   []byte
	SMFIPv6LLA []byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUAddr) UnmarshalBinary(b []byte) error {
	if len(b) < 5 {
		return errors.Errorf("Bad PDUAddr IE length(%d)", len(b))
	}
	si6lla := GetBit4(b[0])
	PDUSessTypeValue := Get3Bits31(b[0])
	ofs := 1 // Offset
	switch IpAddrType(PDUSessTypeValue) {
	case IPv4:
		i.IPv4 = append(i.IPv4, b[ofs:ofs+4]...)
		ofs += 4

	case IPv6:
		if len(b) < ofs+8 {
			return errors.Errorf("Bad PDUAddr IE length(%d)", len(b))
		}
		i.IPv6IfId = append(i.IPv6IfId, b[ofs:ofs+8]...)
		ofs += 8

	case IPv4v6:
		if len(b) < ofs+12 {
			return errors.Errorf("Bad PDUAddr IE length(%d)", len(b))
		}
		i.IPv6IfId = append(i.IPv6IfId, b[ofs:ofs+8]...)
		ofs += 8

		i.IPv4 = append(i.IPv4, b[ofs:ofs+4]...)
		ofs += 4

	default:
		return errors.Errorf("Bad PDUSessTypeValue")
	}
	if SI6LLA_Present == si6lla {
		if len(b) < ofs+16 {
			return errors.Errorf("Bad PDUAddr IE length(%d)", len(b))
		}
		i.SMFIPv6LLA = append(i.SMFIPv6LLA, b[ofs:ofs+16]...)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUAddr) MarshalBinary() ([]byte, error) {
	PDUSessTypeValue := uint8(0)
	ieLen := 1
	if nil != i.SMFIPv6LLA {
		ieLen += 16
	}
	if nil != i.IPv4 {
		ieLen += 4
		PDUSessTypeValue = uint8(IPv4)
	}
	if nil != i.IPv6IfId {
		ieLen += 8
		if nil != i.IPv4 {
			PDUSessTypeValue = uint8(IPv4v6)
		} else {
			PDUSessTypeValue = uint8(IPv6)
		}
	}
	if PDUSessTypeValue == 0 {
		return nil, errors.Errorf("PDUAddr: neither ipv4 nor ipv6IfId exists")
	}

	b := make([]byte, ieLen)

	if i.SMFIPv6LLA != nil {
		b[0] = SetBit4(b[0], SI6LLA_Present)
	}

	b[0] = Set3Bits31(b[0], PDUSessTypeValue)
	ofs := 1 // Offset
	if i.IPv6IfId != nil {
		copy(b[ofs:ofs+8], i.IPv6IfId)
		ofs += 8
	}
	if i.IPv4 != nil {
		copy(b[ofs:ofs+4], i.IPv4)
		ofs += 4
	}
	if i.SMFIPv6LLA != nil {
		copy(b[ofs:ofs+16], i.SMFIPv6LLA)
	}

	return b, nil
}
