package ie

import (
	"bytes"
	"strings"
	"testing"
)

func TestIsFQDN(t *testing.T) {
	testCases := []struct {
		expected bool
		fqdn     string
	}{
		// Valid FQDNs
		{true, "example.com"},
		{true, "sub.example.com"},
		{true, "sub-domain.example.com"},
		{true, "a.com"},
		{true, "A.com"},                     // upper case
		{true, "example.com."},              // FQDN with trailing dot
		{true, "xn--bcher-kva.example.com"}, // IDN with Punycode
		{true, "a-very-long-label-name-that-is-still-under-63-characters.com"},
		{true, "a23456789012345678902234567890323456789042345678905234567890623" + ".com"}, // Label of 63 characters

		// Invalid FQDNs
		{false, ""},                 // Empty string
		{false, "com"},              // TLD only
		{false, ".com"},             // Starts with a dot
		{false, "example..com"},     // Consecutive dots
		{false, "-example.com"},     // Label starts with a hyphen
		{false, "example-.com"},     // Label ends with a hyphen
		{false, "example.com-"},     // Ends with a hyphen
		{false, "sub-.example.com"}, // Subdomain ends with hyphen
		{false, "sub.example-.com"}, // Subdomain ends with hyphen
		{false, "exa_mple.com"},     // Underscore in the label
		{false, "sub!.example.com"}, // Special character in the label
		{false, "sub@.example.com"}, // Special character in the label
		{false, "s%ub.example.com"}, // Special character in the label
		{false, "s=ub.example.com"}, // Special character in the label
		{false, "a234567890b234567890c234567890d234567890e234567890f234567890g234" + ".com"}, // Label > 63 characters
		{false, strings.Repeat("a.", 126) + "com"},                                           // FQDN > 253 characters
	}

	for _, tc := range testCases {
		t.Run(tc.fqdn, func(t *testing.T) {
			actual := isFQDN(tc.fqdn)
			if actual != tc.expected {
				t.Errorf("expect %q to be %v, but is %v", tc.fqdn, tc.expected, actual)
			}
		})
	}
}

func TestEncodeFQDN(t *testing.T) {
	testCases := []struct {
		name     string
		fqdn     string
		expected []byte
	}{
		{
			name:     "empty string",
			fqdn:     "",
			expected: []byte{0},
		},
		{
			name:     "simple FQDN",
			fqdn:     "internet",
			expected: []byte{8, 'i', 'n', 't', 'e', 'r', 'n', 'e', 't'},
		},
		{
			name:     "subdomain",
			fqdn:     "sub.example.com",
			expected: []byte{3, 's', 'u', 'b', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm'},
		},
		{
			name:     "uppercase converted to lowercase",
			fqdn:     "Example.COM",
			expected: []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm'},
		},
		{
			name: "label with hyphen",
			fqdn: "sub-domain.example.com",
			expected: []byte{
				10, 's', 'u', 'b', '-', 'd', 'o', 'm', 'a', 'i', 'n',
				7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm',
			},
		},
		{
			name:     "label with numbers",
			fqdn:     "sub123.example.com",
			expected: []byte{6, 's', 'u', 'b', '1', '2', '3', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm'},
		},
		{
			name:     "label with leading space",
			fqdn:     " abc.com",
			expected: []byte{4, ' ', 'a', 'b', 'c', 3, 'c', 'o', 'm'},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := encodeFQDN(tc.fqdn)
			if !bytes.Equal(actual, tc.expected) {
				t.Errorf("expected %v for FQDN %q, but got %v", tc.expected, tc.fqdn, actual)
			}
		})
	}
}

func TestDecodeFQDN(t *testing.T) {
	testCases := []struct {
		name     string
		apn      []byte
		expected string
	}{
		{
			name:     "empty APN",
			apn:      []byte{0},
			expected: "",
		},
		{
			name:     "simple FQDN",
			apn:      []byte{8, 'i', 'n', 't', 'e', 'r', 'n', 'e', 't'},
			expected: "internet",
		},
		{
			name:     "subdomain",
			apn:      []byte{3, 's', 'u', 'b', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm'},
			expected: "sub.example.com",
		},
		{
			name:     "uppercase converted to lowercase",
			apn:      []byte{7, 'E', 'X', 'A', 'M', 'P', 'L', 'E', 3, 'C', 'O', 'M'},
			expected: "example.com",
		},
		{
			name: "label with hyphen",
			apn: []byte{
				10, 's', 'u', 'b', '-', 'd', 'o', 'm', 'a', 'i', 'n',
				7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				3, 'c', 'o', 'm',
			},
			expected: "sub-domain.example.com",
		},
		{
			name:     "label with numbers",
			apn:      []byte{6, 's', 'u', 'b', '1', '2', '3', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm'},
			expected: "sub123.example.com",
		},
		{
			name:     "label length exceeds buffer length",
			apn:      []byte{10, 'a', 'b', 'c'},
			expected: "",
		},
		{
			name:     "label with tailing space",
			apn:      []byte{6, 't', 'e', 's', 't', 'a', 'p', 4, 'c', 'o', 'm', ' '},
			expected: "testap.com ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := decodeFQDN(tc.apn)
			if actual != tc.expected {
				t.Errorf("expected %q for APN %v, but got %q", tc.expected, tc.apn, actual)
			}
		})
	}
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	testCases := []string{
		"example.com",
		"sub.example.com",
		"deep.sub.example.com",
		"a.b.c.d.e.f.g.h",
		"test123.example456.com",
		"sub-domain.example.com",
	}

	for _, fqdn := range testCases {
		t.Run(fqdn, func(t *testing.T) {
			encoded := encodeFQDN(fqdn)
			decoded := decodeFQDN(encoded)
			if decoded != strings.ToLower(fqdn) {
				t.Errorf("round trip failed: original=%q, decoded=%q", fqdn, decoded)
			}
		})
	}
}
