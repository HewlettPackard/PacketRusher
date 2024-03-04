/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package scenario

type Task struct {
	TaskType   TaskType
	Delay      int
	Hang       bool // hang simulation in this state until ctrl-c is received
	Parameters struct {
		GnbId string
	}
}

type TaskType int32

const (
	AttachToGNB TaskType = iota
	Registration
	Deregistration
	NewPDUSession
	// DestroyPDUSession
	Terminate
	Kill
	Idle
	NGAPHandover
	XNHandover
)
