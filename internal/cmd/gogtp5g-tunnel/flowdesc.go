package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"unsafe"

	"github.com/free5gc/go-gtp5gnl"
	"github.com/khirono/go-nl"
)

// IPFilterRule <- action s dir s proto s 'from' s src 'to' s dst
// action <- 'permit' / 'deny'
// dir <- 'in' / 'out'
// proto <- 'ip' / digit
// src <- addr s ports?
// dst <- addr s ports?
// addr <- 'any' / 'assigned' / cidr
// cidr <- ipv4addr ('/' digit)?
// ipv4addr <- digit ('.' digit){3}
// ports <- port (',' port)*
// port <- (digit '-' digit) / digit
// digit <- [1-9][0-9]+
// s <- ' '+

func ParseFlowDesc(s string) (nl.AttrList, error) {
	var attrs nl.AttrList
	token := strings.Fields(s)
	pos := 0

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	switch token[pos] {
	case "permit":
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.FLOW_DESCRIPTION_ACTION,
			Value: nl.AttrU8(gtp5gnl.SDF_FILTER_PERMIT),
		})
	default:
		return nil, fmt.Errorf("unknown action %v", token[pos])
	}
	pos++

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	switch token[pos] {
	case "in":
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.FLOW_DESCRIPTION_DIRECTION,
			Value: nl.AttrU8(gtp5gnl.SDF_FILTER_IN),
		})
	case "out":
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.FLOW_DESCRIPTION_DIRECTION,
			Value: nl.AttrU8(gtp5gnl.SDF_FILTER_OUT),
		})
	default:
		return nil, fmt.Errorf("unknown direction %v", token[pos])
	}
	pos++

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	var proto uint8
	switch token[pos] {
	case "ip":
		proto = 0xff
	default:
		v, err := strconv.ParseUint(token[pos], 10, 8)
		if err != nil {
			return nil, err
		}
		proto = uint8(v)
	}
	attrs = append(attrs, nl.Attr{
		Type:  gtp5gnl.FLOW_DESCRIPTION_PROTOCOL,
		Value: nl.AttrU8(proto),
	})
	pos++

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	if token[pos] != "from" {
		return nil, fmt.Errorf("not match 'from'")
	}
	pos++

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	src, err := ParseFlowDescIPNet(token[pos])
	if err != nil {
		return nil, err
	}
	pos++
	attrs = append(attrs, nl.Attr{
		Type:  gtp5gnl.FLOW_DESCRIPTION_SRC_IPV4,
		Value: nl.AttrBytes(src.IP),
	})
	attrs = append(attrs, nl.Attr{
		Type:  gtp5gnl.FLOW_DESCRIPTION_SRC_MASK,
		Value: nl.AttrBytes(src.Mask),
	})

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	sports, err := ParseFlowDescPorts(token[pos])
	if err == nil {
		b := EncodePorts(sports)
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.FLOW_DESCRIPTION_SRC_PORT,
			Value: nl.AttrBytes(b),
		})
		pos++
	}

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	if token[pos] != "to" {
		return nil, fmt.Errorf("not match 'to'")
	}
	pos++

	if pos >= len(token) {
		return nil, fmt.Errorf("too few fields %v", len(token))
	}
	dst, err := ParseFlowDescIPNet(token[pos])
	if err != nil {
		return nil, err
	}
	pos++
	attrs = append(attrs, nl.Attr{
		Type:  gtp5gnl.FLOW_DESCRIPTION_DEST_IPV4,
		Value: nl.AttrBytes(dst.IP),
	})
	attrs = append(attrs, nl.Attr{
		Type:  gtp5gnl.FLOW_DESCRIPTION_DEST_MASK,
		Value: nl.AttrBytes(dst.Mask),
	})

	if pos < len(token) {
		dports, err := ParseFlowDescPorts(token[pos])
		if err == nil {
			b := EncodePorts(dports)
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.FLOW_DESCRIPTION_DEST_PORT,
				Value: nl.AttrBytes(b),
			})
		}
	}

	return attrs, nil
}

func ParseFlowDescIPNet(s string) (*net.IPNet, error) {
	if s == "any" {
		return &net.IPNet{
			IP:   net.IPv6zero,
			Mask: net.CIDRMask(0, 128),
		}, nil
	}
	_, ipnet, err := net.ParseCIDR(s)
	if err == nil {
		return ipnet, nil
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid address %v", s)
	}
	v4 := ip.To4()
	if v4 != nil {
		ip = v4
	}
	n := len(ip) * 8
	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(n, n),
	}, nil
}

func ParseFlowDescPorts(s string) ([][]uint16, error) {
	var vals [][]uint16
	for _, port := range strings.Split(s, ",") {
		digit := strings.SplitN(port, "-", 2)
		switch len(digit) {
		case 1:
			v, err := strconv.ParseUint(digit[0], 10, 16)
			if err != nil {
				return nil, err
			}
			vals = append(vals, []uint16{uint16(v)})
		case 2:
			start, err := strconv.ParseUint(digit[0], 10, 16)
			if err != nil {
				return nil, err
			}
			end, err := strconv.ParseUint(digit[1], 10, 16)
			if err != nil {
				return nil, err
			}
			vals = append(vals, []uint16{uint16(start), uint16(end)})
		default:
			return nil, fmt.Errorf("invalid port: %q", port)
		}
	}
	return vals, nil
}

func EncodePorts(ports [][]uint16) []byte {
	b := make([]byte, len(ports)*4)
	off := 0
	for _, p := range ports {
		x := (*uint32)(unsafe.Pointer(&b[off]))
		switch len(p) {
		case 1:
			*x = uint32(p[0])<<16 | uint32(p[0])
		case 2:
			*x = uint32(p[0])<<16 | uint32(p[1])
		}
		off += 4
	}
	return b
}
