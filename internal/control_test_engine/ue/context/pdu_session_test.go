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

// The UE keeps sessions 1..16, and the network names them in its 5GSM messages. The
// check was "> 15" alone, so ID 0 wrapped to index 255 and ended the process, and the
// 16th session could be created but never found or released.
func TestUePduSessionIdOutOfRangeIsRefused(t *testing.T) {
	ue := &UEContext{}

	for _, id := range []uint8{0, 17, 200} {
		_, err := ue.GetPduSession(id)
		assert.Error(t, err, "get, id %d", id)
		assert.Error(t, ue.DeletePduSession(id), "delete, id %d", id)
	}
}

func TestUeSixteenthPduSessionCanBeFoundAndReleased(t *testing.T) {
	ue := &UEContext{}
	for i := 0; i < 16; i++ {
		_, err := ue.CreatePDUSession()
		require.NoError(t, err)
	}

	pduSession, err := ue.GetPduSession(16)
	require.NoError(t, err)
	assert.Equal(t, uint8(16), pduSession.Id)

	require.NoError(t, ue.DeletePduSession(16))
	assert.Nil(t, ue.PduSession[15])
}
