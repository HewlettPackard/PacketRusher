package ie

import (
	"strings"
)

// encodeFQDN encodes the given string as the Name Syntax defined
// in RFC 2181, RFC 1035 and RFC 1123.
func encodeFQDN(fqdn string) []byte {
	fqdn = strings.ToLower(fqdn)
	b := make([]byte, len(fqdn)+1)

	offset := 0
	for _, label := range strings.Split(fqdn, ".") {
		l := len(label)
		b[offset] = uint8(l)
		copy(b[offset+1:], label)
		offset += l + 1
	}

	return b
}

// decodeFQDN decodes the given Name Syntax-encoded []byte as
// a string.
func decodeFQDN(b []byte) string {
	var (
		fqdn   []string
		offset int
	)

	maxLen := len(b)
	for offset < maxLen {
		l := int(b[offset])
		if offset+l+1 > maxLen {
			break
		}
		fqdn = append(fqdn, string(b[offset+1:offset+l+1]))
		offset += l + 1
	}

	return strings.ToLower(strings.Join(fqdn, "."))
}

func toBinaryCodedDecimal(val uint8) uint8 {
	return ((val / 10) << 4) + (val % 10)
}

// reverse operation for BCD
func fromBinaryCodedDecimal(val uint8) uint8 {
	return (val>>4)*10 + (val & (0x0f))
}

// Refer to TS 23.040 - 9.1.2.3  Semi-octet representation
func toSemiOctet(val uint8) uint8 {
	return ((val & 0x0F) << 4) | ((val & 0xF0) >> 4)
}

func isFQDN(addr string) bool {
	if len(addr) > 253 || len(addr) == 0 {
		return false
	}
	if addr[len(addr)-1] == '.' {
		addr = addr[:len(addr)-1]
	}
	labels := strings.Split(addr, ".")
	if len(labels) <= 1 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		allDigits := true
		for _, r := range label {
			if (r < 'a' || r > 'z') &&
				(r < 'A' || r > 'Z') &&
				(r < '0' || r > '9') &&
				r != '-' {
				return false
			}
			if r < '0' || r > '9' {
				allDigits = false
			}
		}
		if allDigits {
			return false
		}
	}
	return true
}
