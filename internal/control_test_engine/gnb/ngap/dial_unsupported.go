// SPDX-License-Identifier: Apache-2.0
//go:build !linux

package ngap

import (
	"context"
	"fmt"
	"github.com/ishidawataru/sctp"
	"net/netip"
)

func dialSCTPContext(ctx context.Context, local, remote netip.AddrPort) (*sctp.SCTPConn, error) {
	return nil, fmt.Errorf("native SCTP startup requires Linux")
}
