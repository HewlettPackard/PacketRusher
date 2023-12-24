/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

type UEMessage struct {
	GNBPduSessions   [16]*GnbPDUSession
	GnbIp            string
	GNBRx            chan UEMessage
	GNBTx            chan UEMessage
	IsNas            bool
	Nas              []byte
	ConnectionClosed bool
	PrUeId           int64
	Mcc              string
	Mnc              string
	UEContext        *GNBUe
}
