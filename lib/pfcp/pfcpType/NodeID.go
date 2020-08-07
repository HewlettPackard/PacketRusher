package pfcpType

import (
	"fmt"
	"math/bits"
	"net"
)

const (
	NodeIdTypeIpv4Address uint8 = iota
	NodeIdTypeIpv6Address
	NodeIdTypeFqdn
)

type NodeID struct {
	NodeIdType  uint8 // 0x00001111
	NodeIdValue []byte
}

func (n *NodeID) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(n.NodeIdType) > 4 {
		return []byte(""), fmt.Errorf("Node ID type shall not be greater than 4 bits binary integer")
	}
	data = append([]byte(""), byte(n.NodeIdType))

	// Octet 6 to o
	if n.NodeIdType == 0 && len(n.NodeIdValue) != net.IPv4len {
		return []byte(""), fmt.Errorf("Length of node ID data shall be 4 Octet if node ID is an IPv4 address")
	}
	if n.NodeIdType == 1 && len(n.NodeIdValue) != net.IPv6len {
		return []byte(""), fmt.Errorf("Length of node ID data shall be 16 Octet if node ID is an IPv6 address")
	}
	data = append(data, n.NodeIdValue...)

	return data, nil
}

func (n *NodeID) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	n.NodeIdType = uint8(data[idx]) & Mask4
	idx = idx + 1

	// Octet 6 to o
	if n.NodeIdType == 0 {
		if length < idx+net.IPv4len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		n.NodeIdValue = data[idx : idx+net.IPv4len]
		idx = idx + net.IPv4len
	} else if n.NodeIdType == 1 {
		if length < idx+net.IPv6len {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		n.NodeIdValue = data[idx : idx+net.IPv6len]
		idx = idx + net.IPv6len
	} else {
		n.NodeIdValue = data[idx:]
		idx = idx + uint16(len(n.NodeIdValue))
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}

func (n *NodeID) ResolveNodeIdToIp() net.IP {
	switch n.NodeIdType {
	case NodeIdTypeIpv4Address, NodeIdTypeIpv6Address:
		return net.IP(n.NodeIdValue)
	case NodeIdTypeFqdn:
		ns, _ := net.LookupHost(string(n.NodeIdValue))
		return net.ParseIP(ns[0])
	default:
		return net.IPv4zero
	}
}
