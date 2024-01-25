// /**
//   - SPDX-License-Identifier: Apache-2.0
//   - © Copyright 2024 Hewlett Packard Enterprise Development LP
//     */
package handler

// import (
// 	"errors"
// 	"fmt"
// 	"my5G-RANTester/test/aio5gc/context"

// 	"github.com/free5gc/nas"
// 	log "github.com/sirupsen/logrus"
// )

// func UpdateState(current *context.UeState, next context.UeState) error {
// 	var updatable bool
// 	c := *current
// 	switch c.(type) {
// 	case *AuthenticationInitiated:
// 		updatable = updatableFromAuthenticationInitiated(next)
// 	case *Deregistrated:
// 		updatable = updatableFromDeregistrated(next)
// 	case *DeregistratedInitiated:
// 		updatable = updatableFromDeregistratedInitiated(next)
// 	case *Registred:
// 		updatable = updatableFromRegistred(next)
// 	case *SecurityContextAvailable:
// 		updatable = updatableFromSecurityContextAvailable(next)
// 	default:
// 	}
// 	if !updatable {
// 		return errors.New("Cannot change state from " + c.ToString() + "to" + next.ToString())
// 	}
// 	*current = next
// 	return nil
// }

// func updatableFromAuthenticationInitiated(next context.UeState) bool {
// 	switch next.(type) {
// 	case *DeregistratedInitiated:
// 	case *SecurityContextAvailable:
// 	default:
// 		return false
// 	}
// 	return true
// }

// func updatableFromDeregistrated(next context.UeState) bool {
// 	switch next.(type) {
// 	case *AuthenticationInitiated:
// 	default:
// 		return false
// 	}
// 	return true
// }

// func updatableFromDeregistratedInitiated(next context.UeState) bool {
// 	switch next.(type) {
// 	case *Deregistrated:
// 	default:
// 		return false
// 	}
// 	return true
// }

// func updatableFromRegistred(next context.UeState) bool {
// 	switch next.(type) {
// 	case *DeregistratedInitiated:
// 	case *Deregistrated:
// 	default:
// 		return false
// 	}
// 	return true
// }

// func updatableFromSecurityContextAvailable(next context.UeState) bool {
// 	switch next.(type) {
// 	case *DeregistratedInitiated:
// 	case *Registred:
// 	default:
// 		return false
// 	}
// 	return true
// }

// type AuthenticationInitiated struct{}

// func (a *AuthenticationInitiated) AuthenticationResponse(msg *nas.Message, gnb *context.GNBContext, ueContext *context.UEContext, fgc *context.Aio5gc) error {
// 	return AuthenticationResponse(msg, gnb, ueContext, fgc)
// }
// func (a *AuthenticationInitiated) RegistrationComplete(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for AuthenticationInitiated UE")
// }
// func (a *AuthenticationInitiated) RegistrationRequest(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	log.Warn("[5GC][NAS] Unexpected message: received RegistrationRequest for AuthenticationInitiated UE")
// 	return RegistrationRequest(msg, ueContext, gnb, fgc)
// }
// func (a *AuthenticationInitiated) SecurityModeComplete(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received SecurityModeComplete for AuthenticationInitiated UE")
// }
// func (a *AuthenticationInitiated) UEOriginatingDeregistration(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return UEOriginatingDeregistration(msg, ueContext, gnb, fgc)
// }
// func (a *AuthenticationInitiated) UlNasTransport(*nas.Message, *context.GNBContext, *context.UEContext, *context.SessionContext) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UlNasTransport for AuthenticationInitiated UE")
// }
// func (a *AuthenticationInitiated) ToString() string {
// 	return "AuthenticationInitiated"
// }

// type Deregistrated struct{}

// func (d *Deregistrated) AuthenticationResponse(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for Deregistrated UE")
// }
// func (d *Deregistrated) RegistrationComplete(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for Deregistrated UE")
// }
// func (d *Deregistrated) RegistrationRequest(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return RegistrationRequest(msg, ueContext, gnb, fgc)
// }
// func (d *Deregistrated) SecurityModeComplete(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received SecurityModeComplete for Deregistrated UE")
// }
// func (d *Deregistrated) UEOriginatingDeregistration(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UEOriginatingDeregistration for Deregistrated UE")
// }
// func (d *Deregistrated) UlNasTransport(*nas.Message, *context.GNBContext, *context.UEContext, *context.SessionContext) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UlNasTransport for Deregistrated UE")
// }
// func (d *Deregistrated) ToString() string {
// 	return "Deregistrated"
// }

// type DeregistratedInitiated struct{}

// func (d *DeregistratedInitiated) AuthenticationResponse(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) RegistrationComplete(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) RegistrationRequest(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received SecurityModeComplete for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) SecurityModeComplete(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received SecurityModeComplete for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) UEOriginatingDeregistration(*nas.Message, *context.UEContext, *context.GNBContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UEOriginatingDeregistration for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) UlNasTransport(*nas.Message, *context.GNBContext, *context.UEContext, *context.SessionContext) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UlNasTransport for DeregistratedInitiated UE")
// }
// func (d *DeregistratedInitiated) ToString() string {
// 	return "DeregistratedInitiated"
// }

// type Registred struct{}

// func (r *Registred) AuthenticationResponse(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for Registred UE")
// }
// func (r *Registred) RegistrationComplete(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for Registred UE")
// }
// func (r *Registred) RegistrationRequest(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	log.Warn("[5GC][NAS] Unexpected message: received RegistrationRequest for Registred UE")
// 	//TODO: Send to succesful RegRequest part
// 	return RegistrationRequest(msg, ueContext, gnb, fgc)
// }
// func (r *Registred) SecurityModeComplete(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received SecurityModeComplete for Registred UE")
// }
// func (r *Registred) UEOriginatingDeregistration(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return UEOriginatingDeregistration(msg, ueContext, gnb, fgc)
// }
// func (r *Registred) UlNasTransport(msg *nas.Message, gnb *context.GNBContext, ueContext *context.UEContext, sm *context.SessionContext) error {
// 	return UlNasTransport(msg, gnb, ueContext, sm)
// }
// func (r *Registred) ToString() string {
// 	return "Registred"
// }

// type SecurityContextAvailable struct{}

// func (s *SecurityContextAvailable) AuthenticationResponse(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received AuthenticationResponse for SecurityContextAvailable UE")
// }
// func (s *SecurityContextAvailable) RegistrationComplete(*nas.Message, *context.GNBContext, *context.UEContext, *context.Aio5gc) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received RegistrationComplete for SecurityContextAvailable UE")
// }
// func (s *SecurityContextAvailable) RegistrationRequest(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	log.Warn("[5GC][NAS] Unexpected message: received RegistrationRequest for SecurityContextAvailable UE")
// 	return RegistrationRequest(msg, ueContext, gnb, fgc)
// }
// func (s *SecurityContextAvailable) SecurityModeComplete(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return SecurityModeComplete(msg, ueContext, gnb, fgc)
// }
// func (s *SecurityContextAvailable) UEOriginatingDeregistration(msg *nas.Message, ueContext *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) error {
// 	return UEOriginatingDeregistration(msg, ueContext, gnb, fgc)
// }
// func (s *SecurityContextAvailable) UlNasTransport(*nas.Message, *context.GNBContext, *context.UEContext, *context.SessionContext) error {
// 	return fmt.Errorf("[5GC][NAS] Unexpected message: received UlNasTransport for SecurityContextAvailable UE")
// }
// func (s *SecurityContextAvailable) ToString() string {
// 	return "SecurityContextAvailable"
// }
