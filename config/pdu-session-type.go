// SPDX-License-Identifier: Apache-2.0
package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/free5gc/nas/ie"
)

// PDUSessionType selects the requested inner IP family; zero preserves IPv4.
type PDUSessionType uint8

func (t PDUSessionType) NASValue() uint8 {
	if t == 0 {
		return ie.PDUSessType_IPv4
	}
	return uint8(t)
}

func (t *PDUSessionType) UnmarshalYAML(_ context.Context, unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	switch strings.ToLower(value) {
	case "", "ipv4":
		*t = PDUSessionType(ie.PDUSessType_IPv4)
	case "ipv6":
		*t = PDUSessionType(ie.PDUSessType_IPv6)
	case "ipv4v6":
		*t = PDUSessionType(ie.PDUSessType_IPv4v6)
	default:
		return fmt.Errorf("ue.pdusessiontype %q must be IPv4, IPv6, or IPv4v6", value)
	}
	return nil
}
