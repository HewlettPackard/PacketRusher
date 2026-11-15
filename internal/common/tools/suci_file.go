package tools

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// LoadSuciFile reads prepared SUCIs: one hex-encoded 5GS mobile identity
// (without IEI or length) per line, sent verbatim, valid or not. Spaces, colons and a 0x prefix are ignored,
// as are blank lines and # comments.
func LoadSuciFile(path string) ([][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sucis [][]byte
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(strings.ToLower(line), "0x")
		line = strings.NewReplacer(" ", "", ":", "").Replace(line)
		raw, err := hex.DecodeString(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		if len(raw) == 0 {
			return nil, fmt.Errorf("%s:%d: empty identity", path, n)
		}
		sucis = append(sucis, raw)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(sucis) == 0 {
		return nil, fmt.Errorf("%s: no SUCIs found", path)
	}
	return sucis, nil
}
