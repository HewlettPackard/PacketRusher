package config

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/goccy/go-yaml"
)

// IPv4Port is a tuple of IPv4 address and port number.
type IPv4Port struct {
	netip.AddrPort
}

var _ yaml.InterfaceUnmarshalerContext = &IPv4Port{}

// UnmarshalYAML implements yaml.InterfaceUnmarshalerContext interface.
func (v *IPv4Port) UnmarshalYAML(ctx context.Context, unmarshal func(interface{}) error) error {
	var value struct {
		IP   string `yaml:"ip"`   // IPv4 address or hostname
		Port uint16 `yaml:"port"` // port number
	}
	if e := unmarshal(&value); e != nil {
		return e
	}

	ips, e := net.DefaultResolver.LookupNetIP(ctx, "ip4", value.IP)
	if e != nil || len(ips) == 0 {
		return fmt.Errorf("cannot resolve %s as IPv4 address: %w", value.IP, e)
	}

	v.AddrPort = netip.AddrPortFrom(ips[0].Unmap(), value.Port)
	return nil
}

// WithNextAddr returns a copy of IPv4Port with the next IPv4 address.
func (v IPv4Port) WithNextAddr() (copy IPv4Port) {
	copy.AddrPort = netip.AddrPortFrom(v.Addr().Next(), v.Port())
	return
}

// WithPort returns a copy of IPv4Port with the specified port number.
func (v IPv4Port) WithPort(port uint16) (copy IPv4Port) {
	copy.AddrPort = netip.AddrPortFrom(v.Addr(), port)
	return
}
