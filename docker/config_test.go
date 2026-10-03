package docker_test

import (
	"encoding/hex"
	"os"
	"testing"

	"my5G-RANTester/config"

	"github.com/goccy/go-yaml"
)

// Decode the container examples using the same strict schema as config.Load.
// This catches obsolete AMF layouts and unrendered environment placeholders
// without contacting an AMF or requiring Docker or kernel modules.
func TestExampleConfigurations(t *testing.T) {
	for _, file := range []string{"config.yml", "config.bridge.yml"} {
		t.Run(file, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			var cfg config.Config
			if err := yaml.NewDecoder(f, yaml.Strict()).Decode(&cfg); err != nil {
				t.Fatal(err)
			}
			if len(cfg.AMFs) == 0 || !cfg.AMFs[0].Addr().Is4() || cfg.AMFs[0].Port() == 0 {
				t.Fatal("example must contain a usable IPv4 AMF endpoint")
			}
			if !cfg.GNodeB.ControlIF.Addr().Is4() || !cfg.GNodeB.DataIF.Addr().Is4() {
				t.Fatal("example must contain usable N2 and N3 addresses")
			}
			for name, value := range map[string]string{"key": cfg.Ue.Key, "opc": cfg.Ue.Opc} {
				decoded, err := hex.DecodeString(value)
				if err != nil || len(decoded) != 16 {
					t.Fatalf("%s must be a 16-byte hexadecimal credential", name)
				}
			}
			if cfg.Ue.Dnn == "" || cfg.Ue.Msin == "" {
				t.Fatal("example must contain a DNN and MSIN")
			}
		})
	}
}
