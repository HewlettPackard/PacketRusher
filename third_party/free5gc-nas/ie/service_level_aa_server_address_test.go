package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSvcLvlAAServerAddr_UnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected *SvcLvlAAServerAddr
	}{
		{
			name: "IPv4 address",
			input: []byte{
				1,
				192, 168, 0, 1,
			},
			expected: &SvcLvlAAServerAddr{ipv4: "192.168.0.1"},
		},
		{
			name: "IPv6 address",
			input: []byte{
				2,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
			},
			expected: &SvcLvlAAServerAddr{ipv6: "::1"},
		},
		{
			name: "IPv4 and IPv6 address",
			input: []byte{
				3,
				192, 168, 0, 1,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
			},
			expected: &SvcLvlAAServerAddr{ipv4: "192.168.0.1", ipv6: "::1"},
		},
		{
			name: "FQDN address",
			input: []byte{
				4,
				104, 101, 108, 108, 111, 46, 99, 111, 109,
			},
			expected: &SvcLvlAAServerAddr{fqdn: "hello.com"},
		},
		{
			name:     "Unknown address type",
			input:    []byte{5},
			expected: nil,
		},
		{
			name: "Invalid length for IPv4 address",
			input: []byte{
				1,
				192, 168, 0,
			},
			expected: nil,
		},
		{
			name: "Invalid length for IPv6 address",
			input: []byte{
				2,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0,
			},
			expected: nil,
		},
		{
			name: "Invalid length for IPv4 and IPv6 address",
			input: []byte{
				3,
				192, 168, 0, 1,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
			expected: nil,
		},
		{
			name:     "Empty input",
			input:    []byte{},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var i SvcLvlAAServerAddr
			err := i.UnmarshalBinary(tc.input)
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &i)
			}
		})
	}
}

func TestSvcLvlAAServerAddr_MarshalBinary(t *testing.T) {
	testCases := []struct {
		name     string
		input    *SvcLvlAAServerAddr
		expected []byte
	}{
		{
			name:     "IPv4 address",
			input:    &SvcLvlAAServerAddr{ipv4: "192.168.0.1"},
			expected: []byte{1, 192, 168, 0, 1},
		},
		{
			name:  "IPv6 address",
			input: &SvcLvlAAServerAddr{ipv6: "::1"},
			expected: []byte{
				2,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "IPv4 and IPv6 address",
			input: &SvcLvlAAServerAddr{ipv4: "192.168.0.1", ipv6: "::1"},
			expected: []byte{
				3,
				192, 168, 0, 1,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:     "FQDN address",
			input:    &SvcLvlAAServerAddr{fqdn: "hello.com"},
			expected: []byte{4, 104, 101, 108, 108, 111, 46, 99, 111, 109},
		},
		{
			name:     "Empty address",
			input:    &SvcLvlAAServerAddr{},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			}
		})
	}
}

func TestSvcLvlAAServerAddr_SetAddr(t *testing.T) {
	testCases := []struct {
		name     string
		addr     string
		expected *SvcLvlAAServerAddr
	}{
		{
			name:     "IPv4 address",
			addr:     "192.168.0.1",
			expected: &SvcLvlAAServerAddr{ipv4: "192.168.0.1"},
		},
		{
			name:     "IPv6 address",
			addr:     "::1",
			expected: &SvcLvlAAServerAddr{ipv6: "::1"},
		},
		{
			name:     "FQDN address",
			addr:     "hello.com",
			expected: &SvcLvlAAServerAddr{fqdn: "hello.com"},
		},
		{
			name:     "Invalid IPv4 - 1",
			addr:     "192.168.0.256",
			expected: nil,
		},
		{
			name:     "Invalid IPv4 - 2",
			addr:     "192.168.0.",
			expected: nil,
		},
		{
			name:     "Invalid IPv6 - 1",
			addr:     ":::1",
			expected: nil,
		},
		{
			name:     "Invalid IPv6 - 2",
			addr:     "::65536",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var i SvcLvlAAServerAddr
			err := i.SetAddr(tc.addr)
			if tc.expected == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &i)
			}
		})
	}
}
