package pfcpType

import (
	"encoding/binary"
	"fmt"
	"net"
)

type FSEID struct {
	V4          bool
	V6          bool
	Seid        uint64
	Ipv4Address net.IP
	Ipv6Address net.IP
}

func (f *FSEID) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5
	tmpUint8 := btou(f.V4)<<1 | btou(f.V6)
	data = append([]byte(""), byte(tmpUint8))
	idx = idx + 1

	// Octet 6 to 13
	data = append(data, make([]byte, 8)...)
	binary.BigEndian.PutUint64(data[idx:], f.Seid)

	// Octet m to (m+3)
	if f.V4 {
		if f.Ipv4Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv4 address shall be present if V4 is set")
		}
		data = append(data, f.Ipv4Address.To4()...)
	}

	// Octet p to (p+15)
	if f.V6 {
		if f.Ipv6Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv6 address shall be present if V6 is set")
		}
		data = append(data, f.Ipv6Address.To16()...)
	}

	if !f.V4 && !f.V6 {
		return []byte(""), fmt.Errorf("At least one of V4 and V6 flags shall be set")
	}

	return data, nil
}

func (f *FSEID) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	f.V4 = utob(uint8(data[idx]) & BitMask2)
	f.V6 = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet 6 to 13
	if length < idx+8 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	f.Seid = binary.BigEndian.Uint64(data[idx:])
	idx = idx + 8

	// Octet m to (m+3)
	if f.V4 {
		if length < idx+net.IPv4len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		f.Ipv4Address = net.IP(data[idx : idx+net.IPv4len])
		idx = idx + net.IPv4len
	}

	// Octet p to (p+15)
	if f.V6 {
		if length < idx+net.IPv6len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		f.Ipv6Address = net.IP(data[idx : idx+net.IPv6len])
		idx = idx + net.IPv6len
	}

	if !f.V4 && !f.V6 {
		return fmt.Errorf("None of V4 and V6 flags is set")
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
