package pfcpType

import (
	"fmt"
	"free5gc/lib/util_3gpp"
	"math/bits"
	"net"
)

type UserPlaneIPResourceInformation struct {
	Assosi          bool
	Assoni          bool
	Teidri          uint8 // 0x00011100
	V6              bool
	V4              bool
	TeidRange       uint8
	Ipv4Address     net.IP
	Ipv6Address     net.IP
	NetworkInstance util_3gpp.Dnn
	SourceInterface uint8 // 0x00001111
}

func (u *UserPlaneIPResourceInformation) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(u.Teidri) > 3 {
		return []byte(""), fmt.Errorf("TEIDRI shall not be greater than 3 bits binary integer")
	}
	tmpUint8 := btou(u.Assosi)<<6 |
		btou(u.Assoni)<<5 |
		u.Teidri<<2 |
		btou(u.V6)<<1 |
		btou(u.V4)
	data = append([]byte(""), byte(tmpUint8))

	// Octet 6
	if u.Teidri != 0 {
		data = append(data, byte(u.TeidRange))
	}

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

	if !u.V4 && !u.V6 {
		return []byte(""), fmt.Errorf("At least one of V4 and V6 flags shall be set")
	}

	// Octet k to l
	if u.Assoni {
		AssoniBuf, _ := u.NetworkInstance.MarshalBinary()
		data = append(data, AssoniBuf...)
	}

	// Octet r
	if u.Assosi {
		if bits.Len8(u.SourceInterface) > 4 {
			return []byte(""), fmt.Errorf("Source interface shall not be greater than 4 bits binary integer")
		}
		data = append(data, byte(u.SourceInterface))
	}

	return data, nil
}

func (u *UserPlaneIPResourceInformation) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	u.Assosi = utob(uint8(data[idx]) & BitMask7)
	u.Assoni = utob(uint8(data[idx]) & BitMask6)
	u.Teidri = uint8(data[idx]) >> 2 & Mask3
	u.V6 = utob(uint8(data[idx]) & BitMask2)
	u.V4 = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet 6
	if u.Teidri != 0 {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		u.TeidRange = uint8(data[idx])
		idx = idx + 1
	}

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

	if !u.V4 && !u.V6 {
		return fmt.Errorf("None of V4 and V6 flags is set")
	}

	// Octet r
	if u.Assosi {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		u.SourceInterface = data[length-1] & Mask4
		data = data[:length-1]
	}

	// Octet k to l
	if u.Assoni {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		err := u.NetworkInstance.UnmarshalBinary(data[idx:])
		if err != nil {
			return err
		}
	}

	return nil
}
