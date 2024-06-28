/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

type Task struct {
	TaskType   TaskType
	Delay      int
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
	Terminate
	Kill
	Idle
	ServiceRequest
	NGAPHandover
	XNHandover
	// DestroyPDUSession
)

func (t TaskType) ToStr() string {
	switch t {
	case AttachToGNB:
		return "AttachToGNB"
	case Registration:
		return "Registration"
	case Deregistration:
		return "Deregistration"
	case NewPDUSession:
		return "NewPDUSession"
	// case DestroyPDUSession:
	// return ""
	case Terminate:
		return "Terminate"
	case Kill:
		return "Kill"
	case Idle:
		return "Idle"
	case ServiceRequest:
		return "ServiceRequest"
	case NGAPHandover:
		return "NGAPHandover"
	case XNHandover:
		return "XNHandover"
	default:
		return "Undefined"
	}
}
