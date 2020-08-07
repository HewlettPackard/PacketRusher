package pfcpType

import (
	"fmt"
	"net"
)

type RemoteGTPUPeer struct {
	V4          bool
	V6          bool
	Ipv4Address net.IP
	Ipv6Address net.IP
}

func (r *RemoteGTPUPeer) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(r.V4)<<1 | btou(r.V6)
	data = append([]byte(""), byte(tmpUint8))

	// Octet m to (m+3)
	if r.V4 {
		if r.Ipv4Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv4 address shall be present if V4 is set")
		}
		data = append(data, r.Ipv4Address.To4()...)
	}

	// Octet p to (p+15)
	if r.V6 {
		if r.Ipv6Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv6 address shall be present if V6 is set")
		}
		data = append(data, r.Ipv6Address.To16()...)
	}

	if (r.V4 && r.V6) || (!r.V4 && !r.V6) {
		return []byte(""), fmt.Errorf("Either V4 and V6 shall be set")
	}

	return data, nil
}

func (r *RemoteGTPUPeer) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	r.V4 = utob(uint8(data[idx]) & BitMask2)
	r.V6 = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet m to (m+3)
	if r.V4 {
		if length < idx+net.IPv4len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		r.Ipv4Address = net.IP(data[idx : idx+net.IPv4len])
		idx = idx + net.IPv4len
	}

	// Octet p to (p+15)
	if r.V6 {
		if length < idx+net.IPv6len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		r.Ipv6Address = net.IP(data[idx : idx+net.IPv6len])
		idx = idx + net.IPv6len
	}

	if (r.V4 && r.V6) || (!r.V4 && !r.V6) {
		return fmt.Errorf("Both or none of V4 and V6 is set")
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
