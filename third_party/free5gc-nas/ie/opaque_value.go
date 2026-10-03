package ie

import "github.com/pkg/errors"

// copyIEValue validates the value length and gives the native IE its own buffer.
// Opaque values do not claim to implement the procedures carried inside them.
func copyIEValue(name string, b []byte, min, max int) ([]byte, error) {
	if len(b) < min || len(b) > max {
		return nil, errors.Errorf("%s IE length %d outside %d..%d", name, len(b), min, max)
	}
	return append([]byte(nil), b...), nil
}
