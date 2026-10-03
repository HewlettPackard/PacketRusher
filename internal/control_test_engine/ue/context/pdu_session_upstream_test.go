/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Retain the upstream lower/upper-boundary regressions with the standardized
// 1..15 range; identifier 16 is reserved in the native NAS codec.
func TestUePduSessionIdOutOfRangeIsRefused(t *testing.T) {
	ue := &UEContext{}

	for _, id := range []uint8{0, 16, 17, 200} {
		_, err := ue.GetPduSession(id)
		assert.Error(t, err, "get, id %d", id)
		assert.Error(t, ue.DeletePduSession(id), "delete, id %d", id)
	}
}

func TestUeFifteenthPduSessionCanBeFoundAndReleased(t *testing.T) {
	ue := &UEContext{}
	for i := 0; i < 15; i++ {
		_, err := ue.CreatePDUSession()
		require.NoError(t, err)
	}

	pduSession, err := ue.GetPduSession(15)
	require.NoError(t, err)
	assert.Equal(t, uint8(15), pduSession.Id)

	require.NoError(t, ue.DeletePduSession(15))
	assert.Nil(t, ue.PduSession[14])
}
