// SPDX-License-Identifier: Apache-2.0
//go:build linux

package context

import (
	"encoding/binary"
	"errors"
	"fmt"
	goRuntime "runtime"
	"syscall"
	"unsafe"

	"github.com/ishidawataru/sctp"
)

type nativeStatusBuffer struct {
	bytes  [256]byte
	length uint32
}

// The SCTP dependency's uintptr API does not mark pointer arguments as escaping.
// Allocate outside its caller's stack and retain ownership through the syscall.
//
//go:noinline
func newNativeStatusBuffer() *nativeStatusBuffer {
	return &nativeStatusBuffer{length: 256}
}

// SCTPAssociationState reads Linux's native sctp_status. Only its fixed-width
// header is decoded; the complete buffer also accommodates sctp_paddrinfo.
func (gnb *GNBContext) SCTPAssociationState() (uint32, error) {
	conn := gnb.GetSCTPConn()
	if conn == nil {
		return 0, fmt.Errorf("no SCTP association")
	}
	status := newNativeStatusBuffer()
	_, _, err := conn.Getsockopt(sctp.SCTP_STATUS, uintptr(unsafe.Pointer(&status.bytes[0])), uintptr(unsafe.Pointer(&status.length)))
	goRuntime.KeepAlive(status)
	if err != nil {
		// Linux removes a completed one-to-one association before its accepted
		// descriptor is retired. STATUS then returns EINVAL and getpeername
		// reports no peer. Require the socket to remain valid/bound as well;
		// an unknown option error or an invalid/reused descriptor is not proof.
		if errors.Is(err, syscall.EINVAL) && conn.LocalAddr() != nil && conn.RemoteAddr() == nil {
			return 1, nil // SCTP_CLOSED
		}
		return 0, err
	}
	if status.length < 8 {
		return 0, fmt.Errorf("truncated SCTP association status: %d", status.length)
	}
	return binary.NativeEndian.Uint32(status.bytes[4:8]), nil
}

// RetiredNGSetup retains a failed reply and positively observed kernel shutdown.
// A startup peer may abandon its association before a buffered request arrives.
// This is transport retirement, not successful NG Setup or a UE procedure.
type RetiredNGSetup struct {
	State uint32
	Cause error
}

func (e *RetiredNGSetup) Error() string {
	return fmt.Sprintf("NG Setup peer retired in SCTP state %d: %v", e.State, e.Cause)
}
func (e *RetiredNGSetup) Unwrap() error { return e.Cause }

// ClassifyRetiredNGSetup leaves unknown/live association failures unchanged.
// It is deliberately used only for NG Setup; NAS/PDU send failures remain errors.
func (gnb *GNBContext) ClassifyRetiredNGSetup(err error) error {
	if err == nil {
		return nil
	}
	state, statusErr := gnb.SCTPAssociationState()
	if statusErr == nil && (state == 1 || state >= 5 && state <= 8) {
		return &RetiredNGSetup{State: state, Cause: err}
	}
	return err
}
