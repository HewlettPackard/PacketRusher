/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A PDU Session ID is 0..255 on the wire, but only IDs 1..15 are assigned by TS 24.007.
// The range check was "< 1 && > 16", which is never true, so an invalid ID from the AMF
// indexed the session array out of range and ended the process.
func TestGnbPduSessionIdOutOfRangeIsRefused(t *testing.T) {
	for _, id := range []int64{0, 16, 17, 200, -1} {
		ue := &GNBUe{}

		_, err := ue.CreatePduSession(id, "10.0.0.1", "01", "010203", 1, 1, 1, 9, 1, 1)
		assert.Error(t, err, "create, id %d", id)

		_, err = ue.GetPduSession(id)
		assert.Error(t, err, "get, id %d", id)

		assert.Error(t, ue.DeletePduSession(id), "delete, id %d", id)
	}
}

func TestGnbPduSessionIdAtTheBoundsIsAccepted(t *testing.T) {
	for _, id := range []int64{1, 15} {
		ue := &GNBUe{}

		_, err := ue.GetPduSession(id)
		assert.NoError(t, err, "get, id %d", id)
		assert.NoError(t, ue.DeletePduSession(id), "delete, id %d", id)
	}
}
