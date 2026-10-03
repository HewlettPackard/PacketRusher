// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"github.com/free5gc/ngap/ie"
	"math"
	"testing"
)

func TestVirtualUEIdentitySupportsLoadTestIDs(t *testing.T) {
	for _, id := range []int64{1, 256, 257, 100000, math.MaxInt64} {
		c := &ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{RRCContainer: &ie.RRCContainer{Value: VirtualUERrc(id)}}
		got, err := VirtualUEID(c)
		if err != nil || got != id {
			t.Fatalf("ID %d: %d, %v", id, got, err)
		}
	}
	legacy := &ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{IndexToRFSP: &ie.IndexToRFSP{Value: 256}}
	if got, err := VirtualUEID(legacy); err != nil || got != 256 {
		t.Fatalf("legacy identity: %d %v", got, err)
	}
	for _, c := range []*ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{nil, {}, {RRCContainer: &ie.RRCContainer{Value: []byte("PRUE")}, IndexToRFSP: &ie.IndexToRFSP{Value: 1}}, {RRCContainer: &ie.RRCContainer{Value: VirtualUERrc(0)}}, {RRCContainer: &ie.RRCContainer{Value: VirtualUERrc(-1)}}} {
		if _, err := VirtualUEID(c); err == nil {
			t.Fatal("invalid virtual UE identity accepted")
		}
	}
}
