package pfcpType

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	OuterHeaderCreationGtpUUdpIpv4 uint16 = 1
	OuterHeaderCreationGtpUUdpIpv6 uint16 = 1 << 1
	OuterHeaderCreationUdpIpv4     uint16 = 1 << 2
	OuterHeaderCreationUdpIpv6     uint16 = 1 << 3
)

type OuterHeaderCreation struct {
	OuterHeaderCreationDescription uint16
	Teid                           uint32
	Ipv4Address                    net.IP
	Ipv6Address                    net.IP
	PortNumber                     uint16
}

func (o *OuterHeaderCreation) MarshalBinary() (data []byte, err error) {
	octet5 := uint8(o.OuterHeaderCreationDescription & Mask8)
	var GtpU, Udp, Ipv4, Ipv6 bool
	GtpU = utob(octet5&BitMask1) || utob(octet5&BitMask2)
	Udp = utob(octet5&BitMask3) || utob(octet5&BitMask4)
	Ipv4 = utob(octet5&BitMask1) || utob(octet5&BitMask3)
	Ipv6 = utob(octet5&BitMask2) || utob(octet5&BitMask4)
	if !GtpU && !Udp {
		return []byte(""), fmt.Errorf("At least one bit of outer header description field shall be set")
	}

	var idx uint16 = 0
	// Octet 5 to 6
	data = make([]byte, 2)
	binary.LittleEndian.PutUint16(data[idx:], o.OuterHeaderCreationDescription)
	idx = idx + 2

	// Octet m to (m+3)
	if GtpU {
		data = append(data, make([]byte, 4)...)
		binary.BigEndian.PutUint32(data[idx:], o.Teid)
		idx = idx + 4
	}

	// Octet p to (p+3)
	if Ipv4 {
		if o.Ipv4Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv4 address shall be present")
		}
		data = append(data, o.Ipv4Address.To4()...)
		idx = idx + net.IPv4len
	}

	// Octet q to (q+15)
	if Ipv6 {
		if o.Ipv6Address.IsUnspecified() {
			return []byte(""), fmt.Errorf("IPv6 address shall be present")
		}
		data = append(data, o.Ipv6Address.To16()...)
		idx = idx + net.IPv6len
	}

	// Octet r to (r+1)
	if Udp {
		data = append(data, make([]byte, 2)...)
		binary.BigEndian.PutUint16(data[idx:], o.PortNumber)
	}

	return data, nil
}

func (o *OuterHeaderCreation) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 6
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	o.OuterHeaderCreationDescription = binary.LittleEndian.Uint16(data[idx:])
	idx = idx + 2

	octet5 := uint8(o.OuterHeaderCreationDescription & Mask8)
	var GtpU, Udp, Ipv4, Ipv6 bool
	GtpU = utob(octet5&BitMask1) || utob(octet5&BitMask2)
	Udp = utob(octet5&BitMask3) || utob(octet5&BitMask4)
	Ipv4 = utob(octet5&BitMask1) || utob(octet5&BitMask3)
	Ipv6 = utob(octet5&BitMask2) || utob(octet5&BitMask4)
	if !GtpU && !Udp {
		return fmt.Errorf("None of outer header description field is set")
	}

	// Octet m to (m+3)
	if GtpU {
		if length < idx+4 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		o.Teid = binary.BigEndian.Uint32(data[idx:])
		idx = idx + 4
	}

	// Octet p to (p+3)
	if Ipv4 {
		if length < idx+net.IPv4len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		o.Ipv4Address = net.IP(data[idx : idx+net.IPv4len])
		idx = idx + net.IPv4len
	}

	// Octet q to (q+15)
	if Ipv6 {
		if length < idx+net.IPv6len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		o.Ipv6Address = net.IP(data[idx : idx+net.IPv6len])
		idx = idx + net.IPv6len
	}

	// Octet
	if Udp {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		o.PortNumber = binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
