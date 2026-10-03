/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"net/netip"

	nasType "github.com/free5gc/nas/ie"
)

type UEMessage struct {
	GNBPduSessions    [16]*GnbPDUSession
	GnbIp             netip.Addr
	GtpDevice         *gtp.Device // the gNB's shared GTP-U device, when it has one
	GNBRx             chan UEMessage
	GNBTx             chan UEMessage
	GNBInboundChannel chan UEMessage
	IsNas             bool
	Nas               []byte
	ConnectionClosed  bool
	ConnectionLost    chan struct{} // closed only when this logical connection fails
	PrUeId            int64
	Tmsi              *nasType.MobileId5GS
	Mcc               string
	Mnc               string
	UEContext         *GNBUe
	IsHandover        bool
	Idle              bool
	FetchPagedUEs     bool
	PagedUEs          []PagedUE
}
