// SPDX-License-Identifier: Apache-2.0
package context

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPDUSessionIdentifiers(t *testing.T) {
	ue := &UEContext{}
	for _, id := range []uint8{0, 16, 255} {
		_, err := ue.GetPduSession(id)
		require.Error(t, err)
		require.Error(t, ue.DeletePduSession(id))
	}
	for id := uint8(1); id <= 15; id++ {
		session, err := ue.CreatePDUSession()
		require.NoError(t, err)
		require.Equal(t, id, session.Id)
		found, err := ue.GetPduSession(id)
		require.NoError(t, err)
		require.Same(t, session, found)
	}
	_, err := ue.CreatePDUSession()
	require.Error(t, err)
	require.NoError(t, ue.DeletePduSession(15))
	replacement, err := ue.CreatePDUSession()
	require.NoError(t, err)
	require.Equal(t, uint8(15), replacement.Id)
}
