package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSuciFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sucis.txt")
	content := "# comment\n01 99f907 0000 00 00 0000001002\n\n0x0199f9070000000099999999 99 # trailing\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	sucis, err := LoadSuciFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sucis) != 2 || len(sucis[0]) != 13 || len(sucis[1]) != 13 {
		t.Fatalf("unexpected result: %x", sucis)
	}

	bad := filepath.Join(t.TempDir(), "bad.txt")
	os.WriteFile(bad, []byte("zz\n"), 0o600)
	if _, err := LoadSuciFile(bad); err == nil {
		t.Fatal("expected error for non-hex line")
	}
}
