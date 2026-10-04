// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import "errors"

var (
	// ErrUnavailable identifies backend-specific load/attachment failure. It is
	// never returned for packet/configuration validation or ownership collision.
	ErrUnavailable = errors.New("eBPF backend unavailable")
	// ErrCleanupIncomplete forbids changing backend while resources remain held.
	ErrCleanupIncomplete = errors.New("eBPF cleanup incomplete")
)
