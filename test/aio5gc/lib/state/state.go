// /**
//   - SPDX-License-Identifier: Apache-2.0
//   - © Copyright 2024 Hewlett Packard Enterprise Development LP
//     */
package state

import (
	"fmt"
)

type UE interface {
	ToString() string
	UpdatableTo(int) bool
	Is(int) bool
	GetHistory() []string
	Current() int
}

const (
	AuthenticationInitiated = iota
	Deregistrated
	DeregistratedInitiated
	Registred
	SecurityContextAvailable
)

func UpdateUE(currentPointer *UE, next int) error {

	current := *currentPointer

	if current.UpdatableTo(next) {
		previouses := append(current.GetHistory(), current.ToString())
		switch next {
		case AuthenticationInitiated:
			*currentPointer = authenticationInitiated{history: history{previouses: previouses}}
		case Deregistrated:
			*currentPointer = deregistrated{history: history{previouses: previouses}}
		case DeregistratedInitiated:
			*currentPointer = deregistratedInitiated{history: history{previouses: previouses}}
		case Registred:
			*currentPointer = registred{history: history{previouses: previouses}}
		case SecurityContextAvailable:
			*currentPointer = securityContextAvailable{history: history{previouses: previouses}}
		default:
			return fmt.Errorf("UE State Update does not have transition for %v", toString(next))
		}

		return nil
	}

	return fmt.Errorf("UE state transition from %v to %v not allowed", current.ToString(), toString(next))
}

func GetNewDeregistrated() UE {
	return deregistrated{history: history{previouses: []string{}}}
}

func toString(i int) string {
	switch i {
	case AuthenticationInitiated:
		return authenticationInitiated{}.ToString()
	case Deregistrated:
		return deregistrated{}.ToString()
	case DeregistratedInitiated:
		return deregistratedInitiated{}.ToString()
	case Registred:
		return registred{}.ToString()
	case SecurityContextAvailable:
		return securityContextAvailable{}.ToString()
	default:
		return "Unkown UE State"
	}
}

type history struct {
	previouses []string
}

func (s history) GetHistory() []string {
	return s.previouses
}

type authenticationInitiated struct {
	history
}

func (s authenticationInitiated) ToString() string {
	return "AuthenticationInitiated"
}

func (s authenticationInitiated) UpdatableTo(next int) bool {
	switch next {
	case DeregistratedInitiated:
	case SecurityContextAvailable:
	default:
		return false
	}
	return true
}

func (s authenticationInitiated) Is(state int) bool {
	if state == AuthenticationInitiated {
		return true
	}
	return false
}

func (s authenticationInitiated) Current() int {
	return AuthenticationInitiated
}

type deregistrated struct {
	history
}

func (s deregistrated) ToString() string {
	return "Deregistrated"
}

func (s deregistrated) UpdatableTo(next int) bool {
	switch next {
	case AuthenticationInitiated:
	default:
		return false
	}
	return true
}

func (s deregistrated) Is(state int) bool {
	if state == Deregistrated {
		return true
	}
	return false
}

func (s deregistrated) Current() int {
	return Deregistrated
}

type deregistratedInitiated struct {
	history
}

func (s deregistratedInitiated) ToString() string {
	return "DeregistratedInitiated"
}

func (s deregistratedInitiated) UpdatableTo(next int) bool {
	switch next {
	case Deregistrated:
	default:
		return false
	}
	return true
}

func (s deregistratedInitiated) Is(state int) bool {
	if state == DeregistratedInitiated {
		return true
	}
	return false
}

func (s deregistratedInitiated) Current() int {
	return DeregistratedInitiated
}

type registred struct {
	history
}

func (s registred) ToString() string {
	return "Registred"
}

func (s registred) UpdatableTo(next int) bool {
	switch next {
	case DeregistratedInitiated:
	case Deregistrated:
	default:
		return false
	}
	return true
}

func (s registred) Is(state int) bool {
	if state == Registred {
		return true
	}
	return false
}

func (s registred) Current() int {
	return Registred
}

type securityContextAvailable struct {
	history
}

func (s securityContextAvailable) ToString() string {
	return "SecurityContextAvailable"
}

func (s securityContextAvailable) UpdatableTo(next int) bool {
	switch next {
	case DeregistratedInitiated:
	case Registred:
	default:
		return false
	}
	return true
}

func (s securityContextAvailable) Is(state int) bool {
	if state == SecurityContextAvailable {
		return true
	}
	return false
}

func (s securityContextAvailable) Current() int {
	return SecurityContextAvailable
}
