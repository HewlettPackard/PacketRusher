/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package test

import (
	"my5G-RANTester/test/aio5gc/context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ForceReleaseAllPDUSession deletes each SM context from inside ExecuteForAllSmContexts.
// That used to hold smContextMtx while DeleteSmContext took it again, so the test 5GC
// hung whenever a UE still had a PDU session when it was force-released.
func TestTestAmfDeletesSmContextsWhileIterating(t *testing.T) {
	ue := (&context.AMFContext{}).NewUE(1)
	require.NoError(t, ue.AddSmContext(context.NewSmContext(1)))
	require.NoError(t, ue.AddSmContext(context.NewSmContext(2)))

	done := make(chan struct{})
	go func() {
		ue.ExecuteForAllSmContexts(func(sm *context.SmContext) {
			_, _ = ue.DeleteSmContext(sm.GetPduSessionId())
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deleting an SM context while iterating over them deadlocked")
	}

	for _, id := range []int32{1, 2} {
		_, err := ue.GetSmContext(id)
		assert.Error(t, err, "SM context %d should have been deleted", id)
	}
}
