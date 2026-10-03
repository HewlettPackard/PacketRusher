// SPDX-License-Identifier: Apache-2.0
package config

import (
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPDUSessionTypeConfig(t *testing.T) {
	for _, tc := range []struct {
		value    string
		expected uint8
	}{{"", 1}, {"IPv4", 1}, {"IPv6", 2}, {"IPv4v6", 3}, {"ipv4V6", 3}} {
		t.Run(tc.value, func(t *testing.T) {
			var cfg Config
			document := "ue: {}"
			if tc.value != "" {
				document = "ue:\n  pdusessiontype: " + tc.value
			}
			require.NoError(t, yaml.UnmarshalWithOptions([]byte(document), &cfg, yaml.Strict()))
			require.Equal(t, tc.expected, cfg.Ue.PDUSessionType.NASValue())
		})
	}
	for _, invalid := range []string{"IPv5", "Ethernet", "42"} {
		var cfg Config
		require.Error(t, yaml.UnmarshalWithOptions([]byte("ue:\n  pdusessiontype: "+invalid), &cfg, yaml.Strict()))
	}
}
