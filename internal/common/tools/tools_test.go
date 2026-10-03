package tools

import (
	"testing"

	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
)

// The factory and selector canonicalize the same configured hexadecimal range,
// including a return to the first gNB after a handover.
func TestUEGnbSelectionCanonicalizesConfiguredID(t *testing.T) {
	for _, configuredID := range []string{"00000a", "00000A"} {
		t.Run(configuredID, func(t *testing.T) {
			plmn := config.PlmnList{GnbId: configuredID}
			gnbs := make(map[string]*gnbContext.GNBContext)
			for offset := 0; offset < 2; offset++ {
				id, err := plmn.GNBIDAt(offset)
				if err != nil {
					t.Fatal(err)
				}
				gnbs[id] = nil
			}
			for ueID := 1; ueID <= 3; ueID++ {
				sim := UESimulationConfig{UeId: ueID, Gnbs: gnbs,
					Cfg: config.Config{GNodeB: config.GNodeB{PlmnList: plmn}}}
				for offset := 0; offset < 4; offset++ {
					want := []string{"00000A", "00000B"}[(ueID-1+offset)%2]
					if got := sim.gnbID(offset); got != want {
						t.Errorf("UE %d offset %d: gNB %q, want %q", ueID, offset, got, want)
					}
				}
			}
		})
	}
}

func TestUEGnbSelectionAndHandoverSequence(t *testing.T) {
	cases := []struct {
		name   string
		ueID   int
		gnbIDs []string
		want   []string
	}{
		{"single gNB", 1, []string{"000008"}, []string{"000008", "000008", "000008"}},
		{"first dedicated UE", 1, []string{"000008", "000009", "00000A", "00000B"}, []string{"000008", "000009", "00000A", "00000B", "000008"}},
		{"second dedicated UE", 2, []string{"000008", "000009", "00000A", "00000B"}, []string{"000009", "00000A", "00000B", "000008", "000009"}},
		{"last dedicated UE", 4, []string{"000008", "000009", "00000A", "00000B"}, []string{"00000B", "000008", "000009", "00000A", "00000B"}},
		{"more UEs than gNBs", 7, []string{"000008", "000009", "00000A"}, []string{"000008", "000009", "00000A", "000008"}},
		{"handover with two gNBs", 1, []string{"000008", "000009"}, []string{"000008", "000009", "000008", "000009"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sim := UESimulationConfig{UeId: tc.ueID, Gnbs: make(map[string]*gnbContext.GNBContext)}
			sim.Cfg = config.Config{GNodeB: config.GNodeB{PlmnList: config.PlmnList{GnbId: "000008"}}}
			for _, id := range tc.gnbIDs {
				sim.Gnbs[id] = nil
			}
			for offset, want := range tc.want {
				got := sim.gnbID(offset)
				if got != want {
					t.Errorf("offset %d: gNB %s, want %s", offset, got, want)
				}
				if _, exists := sim.Gnbs[got]; !exists {
					t.Errorf("selected nonexistent gNB %s", got)
				}
			}
		})
	}
}
