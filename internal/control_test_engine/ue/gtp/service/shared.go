/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package service

import (
	"os"
	"strconv"
)

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
// Override with PR_SETUP_SLOTS if a host needs it lower.
var setupSlots = make(chan struct{}, setupConcurrency())

// acquireSetupSlot blocks until this UE may plumb its tunnel.
func acquireSetupSlot() {
	setupSlots <- struct{}{}
}

func releaseSetupSlot() {
	<-setupSlots
}

// setupConcurrency reports how many tunnels may be plumbed at once.
func setupConcurrency() int {
	if v := os.Getenv("PR_SETUP_SLOTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	return 32
}

// verifyRules reports whether each PDR should be read back after it is installed.
// Off by default: over 7236 UEs it found 0 absent rules and 0 retries, and it costs
// two netlink round trips per UE. Set PR_VERIFY_RULES=1 to turn it back on when
// diagnosing a datapath that is silently dropping.
func verifyRules() bool {
	return os.Getenv("PR_VERIFY_RULES") == "1"
}
