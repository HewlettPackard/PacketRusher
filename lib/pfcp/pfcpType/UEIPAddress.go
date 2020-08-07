package pfcpType

import (
	"fmt"
	"net"
)

type UEIPAddress struct {
	Ipv6d                    bool
	Sd                       bool
	V4                       bool
	V6                       bool
	Ipv4Address              net.IP
	Ipv6Address              net.IP
	Ipv6PrefixDelegationBits uint8
}

func (u *UEIPAddress) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(u.Ipv6d)<<3 |
		btou(u.Sd)<<2 |
		btou(u.V4)<<1 |
		btou(u.V6)
	data = append([]byte(""), byte(tmpUint8))

	// Octet m to (m+3)
	if u.V4 {
		if u.Ipv4Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv4 address shall be present if V4 is set")
		}
		data = append(data, u.Ipv4Address.To4()...)
	}

	// Octet p to (p+15)
	if u.V6 {
		if u.Ipv6Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv6 address shall be present if V6 is set")
		}
		data = append(data, u.Ipv6Address.To16()...)
	}

	// Octet r
	if u.V6 && u.Ipv6d {
		data = append(data, byte(u.Ipv6PrefixDelegationBits))
	}

	return data, nil
}

func (u *UEIPAddress) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	u.Ipv6d = utob(uint8(data[idx]) & BitMask4)
	u.Sd = utob(uint8(data[idx]) & BitMask3)
	u.V4 = utob(uint8(data[idx]) & BitMask2)
	u.V6 = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet m to (m+3)
	if u.V4 {
		if length < idx+net.IPv4len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		u.Ipv4Address = net.IP(data[idx : idx+net.IPv4len])
		idx = idx + net.IPv4len
	}

	// Octet p to (p+15)
	if u.V6 {
		if length < idx+net.IPv6len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		u.Ipv6Address = net.IP(data[idx : idx+net.IPv6len])
		idx = idx + net.IPv6len
	}

	// Octet r
	if u.V6 && u.Ipv6d {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		u.Ipv6PrefixDelegationBits = uint8(data[idx])
		idx = idx + 1
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
