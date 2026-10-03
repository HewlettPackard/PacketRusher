package ie

import (
	"net"

	"github.com/pkg/errors"
)

const (
	SvcLvlAAAddr_IPv4   uint8 = 1
	SvcLvlAAAddr_IPv6   uint8 = 2
	SvcLvlAAAddr_IPv4v6 uint8 = 3
	SvcLvlAAAddr_FQDN   uint8 = 4
)

// SvcLvlAAServerAddr is detailed in 9.11.2.12 Service-level-AA server address, 24.501
type SvcLvlAAServerAddr struct {
	// Name, uint8, Bits, Octet
	ipv4 string // 8 -> 1 ,   4 -> 7
	ipv6 string // 8 -> 1 ,   4 -> 19
	fqdn string // 8 -> 1 ,   4 -> ...
}

func (i *SvcLvlAAServerAddr) GetAddr() (string, string, string) {
	return i.ipv4, i.ipv6, i.fqdn
}

func (i *SvcLvlAAServerAddr) SetAddr(addr string) error {
	ip := net.ParseIP(addr)
	if ip != nil {
		if ip.To4() != nil {
			i.ipv4 = addr
		} else if ip.To16() != nil {
			i.ipv6 = addr
		} else {
			return errors.Errorf("unsupported IP address type: %s", addr)
		}
	} else if isFQDN(addr) {
		i.fqdn = addr
	} else {
		return errors.Errorf("invalid address: %s", addr)
	}
	return nil
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAAServerAddr) UnmarshalBinary(b []byte) error {
	if len(b) < 1 || len(b) > 256 {
		return errors.Errorf("bad SvcLvlAAServerAddr IE length(%d)", len(b))
	}

	i.ipv4 = ""
	i.ipv6 = ""
	i.fqdn = ""

	switch b[0] {
	case SvcLvlAAAddr_IPv4:
		if len(b) != 1+4 {
			return errors.Errorf("bad SvcLvlAAServerAddr IE length(%d)-ipv4", len(b))
		}
		i.ipv4 = net.IP(b[1:]).String()
	case SvcLvlAAAddr_IPv6:
		if len(b) != 1+16 {
			return errors.Errorf("bad SvcLvlAAServerAddr IE length(%d)-ipv6", len(b))
		}
		i.ipv6 = net.IP(b[1:]).String()
	case SvcLvlAAAddr_IPv4v6:
		if len(b) != 1+20 {
			return errors.Errorf("bad SvcLvlAAServerAddr IE length(%d)-ipv4v6", len(b))
		}
		i.ipv4 = net.IP(b[1:5]).String()
		i.ipv6 = net.IP(b[5:]).String()
	case SvcLvlAAAddr_FQDN:
		i.fqdn = string(b[1:])
		if !isFQDN(i.fqdn) {
			return errors.Errorf("bad SvcLvlAAServerAddr FQDN(%s)", i.fqdn)
		}
	default:
		return errors.Errorf("unknown address type(%d)", b[0])
	}

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAAServerAddr) MarshalBinary() ([]byte, error) {
	var b []byte

	if i.ipv4 != "" && i.ipv6 != "" {
		b = make([]byte, 1+20)
		b[0] = SvcLvlAAAddr_IPv4v6
		copy(b[1:], net.ParseIP(i.ipv4).To4())
		copy(b[5:], net.ParseIP(i.ipv6).To16())
	} else if i.ipv4 != "" {
		b = make([]byte, 1+4)
		b[0] = SvcLvlAAAddr_IPv4
		copy(b[1:], net.ParseIP(i.ipv4).To4())
	} else if i.ipv6 != "" {
		b = make([]byte, 1+16)
		b[0] = SvcLvlAAAddr_IPv6
		copy(b[1:], net.ParseIP(i.ipv6).To16())
	} else if i.fqdn != "" {
		b = make([]byte, 1+len(i.fqdn))
		b[0] = SvcLvlAAAddr_FQDN
		copy(b[1:], i.fqdn)
	} else {
		return nil, errors.New("bad SvcLvlAAServerAddr")
	}

	return b, nil
}
