// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"context"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"testing"
	"time"

	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/stretchr/testify/require"
)

func TestNGSetupRejectionReturnsErrorWithoutFatalExit(t *testing.T) {
	for _, kind := range []string{"invalid-response", "failure"} {
		t.Run(kind, func(t *testing.T) {
			node := createTestGNBContext()
			amf := node.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
			if kind == "invalid-response" {
				HandlerNgSetupResponse(amf, node, &message.NGSetupResponse{ServedGUAMIList: &ie.ServedGUAMIList{}})
			} else {
				HandlerNgSetupFailure(amf, node, &message.NGSetupFailure{})
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := amf.WaitActive(ctx, node.Done())
			require.Error(t, err)
			require.NotErrorIs(t, err, context.DeadlineExceeded, "decoded rejection must wake startup without waiting for its deadline")
			require.Equal(t, gnbContext.Inactive, amf.GetState())
		})
	}
}
