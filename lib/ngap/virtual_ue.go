// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/free5gc/ngap/ie"
	"math"
)

// The simulator has no radio RRC implementation. Its transparent RRC container
// carries the virtual UE identity between its gNBs without misusing RFSP's
// constrained 1..256 index as an unbounded load-test UE identifier.
func VirtualUERrc(id int64) []byte {
	b := make([]byte, 12)
	copy(b, "PRUE")
	binary.BigEndian.PutUint64(b[4:], uint64(id))
	return b
}

func VirtualUEID(c *ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer) (int64, error) {
	if c != nil && c.RRCContainer != nil && bytes.HasPrefix(c.RRCContainer.Value, []byte("PRUE")) {
		if len(c.RRCContainer.Value) != 12 {
			return 0, fmt.Errorf("invalid virtual UE identity container length")
		}
		id := binary.BigEndian.Uint64(c.RRCContainer.Value[4:])
		if id == 0 || id > math.MaxInt64 {
			return 0, fmt.Errorf("invalid virtual UE ID")
		}
		return int64(id), nil
	}
	// Accept the previous PacketRusher container for existing small scenarios.
	if c != nil && c.IndexToRFSP != nil && c.IndexToRFSP.Value >= 1 && c.IndexToRFSP.Value <= 256 {
		return c.IndexToRFSP.Value, nil
	}
	return 0, fmt.Errorf("transparent container has no virtual UE identity")
}
