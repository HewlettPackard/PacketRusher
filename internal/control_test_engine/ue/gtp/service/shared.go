/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 * © Copyright 2026 Valentin D'Emmanuele
 */
package service

// setupSlots bounds how many tunnels are plumbed at once. Unbounded parallelism
// fails in the netlink layer with "bad file descriptor"; the 500 ms registration
// floor used to hide that by admitting at most two UEs a second.
//
// A single mutex is the wrong cure. SetupGtpInterface runs on the goroutine that
// handles that UE's messages (see gnbMsgHandler), so serialising it backs up the
// gNB's dispatcher: the simulator stops reading, the AMF's sends go unacked, and
// the association eventually times out. That failure looks exactly like a core
// problem and is not one. A small number of slots keeps the netlink layer inside
// what it tolerates while leaving the dispatcher free to make progress.
// Raised from 4 once rule creation stopped opening a socket per call: the bound
// existed to contain that churn, not because the kernel objected to the work.
var setupSlots = make(chan struct{}, 32)

// acquireSetupSlot blocks until this UE may plumb its tunnel.
func acquireSetupSlot() {
	setupSlots <- struct{}{}
}

func releaseSetupSlot() {
	<-setupSlots
}
