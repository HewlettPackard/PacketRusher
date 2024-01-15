/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	"encoding/binary"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"reflect"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/security"
	log "github.com/sirupsen/logrus"
)

func DecodePlainNasNoIntegrityCheck(payload []byte) (*nas.Message, error) {
	const SecurityHeaderTypeMask uint8 = 0x0f

	if payload == nil {
		return nil, fmt.Errorf("nas payload is empty")
	}

	msg := new(nas.Message)
	msg.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & SecurityHeaderTypeMask
	if msg.SecurityHeaderType == nas.SecurityHeaderTypeIntegrityProtectedAndCiphered ||
		msg.SecurityHeaderType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext {
		return nil, fmt.Errorf("nas payload is ciphered")
	}

	if msg.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {
		// remove security Header
		payload = payload[7:]
	}

	err := msg.PlainNasDecode(&payload)
	return msg, err
}

func Encode(ue *context.UEContext, msg *nas.Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("NAS Message is nil")
	}

	// Plain NAS message

	if ue == nil || !(ue.GetSecurityContext() != nil) {
		if msg.GmmMessage == nil {
			return nil, fmt.Errorf("msg.GmmMessage is nil")
		}
		switch msgType := msg.GmmHeader.GetMessageType(); msgType {
		case nas.MsgTypeIdentityRequest:
			if msg.GmmMessage.IdentityRequest == nil {
				return nil,
					fmt.Errorf("identity Request (type unknown) is requierd security, but security context is not available")
			}
			if identityType := msg.GmmMessage.IdentityRequest.SpareHalfOctetAndIdentityType.GetTypeOfIdentity(); identityType !=
				nasMessage.MobileIdentity5GSTypeSuci {
				return nil,
					fmt.Errorf("identity Request (%d) is requierd security, but security context is not available", identityType)
			}
		case nas.MsgTypeAuthenticationRequest:
		case nas.MsgTypeAuthenticationResult:
		case nas.MsgTypeAuthenticationReject:
		case nas.MsgTypeRegistrationReject:
		case nas.MsgTypeDeregistrationAcceptUEOriginatingDeregistration:
		case nas.MsgTypeServiceReject:
		default:
			return nil, fmt.Errorf("NAS message type %d is requierd security, but security context is not available", msgType)
		}
		pdu, err := msg.PlainNasEncode()
		return pdu, err
	} else {
		// Security protected NAS Message
		// a security protected NAS message must be integrity protected, and ciphering is optional
		needCiphering := false
		switch msg.SecurityHeader.SecurityHeaderType {
		case nas.SecurityHeaderTypeIntegrityProtected:
			log.Debug("[5GC][NAS] Encoding Security header type: Integrity Protected")
		case nas.SecurityHeaderTypeIntegrityProtectedAndCiphered:
			log.Debug("[5GC][NAS] Encoding Security header type: Integrity Protected And Ciphered")
			needCiphering = true
		case nas.SecurityHeaderTypeIntegrityProtectedWithNew5gNasSecurityContext:
			log.Debug("[5GC][NAS] Encoding Security header type: Integrity Protected With New 5G Security Context")
			count := new(security.Count)
			count.Set(0, 0)
			ue.GetSecurityContext().SetULCount(*count)
			count = new(security.Count)
			count.Set(0, 0)
			ue.GetSecurityContext().SetDLCount(*count)
		default:
			return nil, fmt.Errorf("wrong security header type: 0x%0x", msg.SecurityHeader.SecurityHeaderType)
		}

		// encode plain nas first
		payload, err := msg.PlainNasEncode()
		if err != nil {
			return nil, fmt.Errorf("plain NAS encode error: %+v", err)
		}
		dlCount := ue.GetSecurityContext().GetDLCount()
		if needCiphering {
			log.Debugf("Encrypt NAS message (algorithm: %+v, DLCount: 0x%0x)", ue.GetSecurityContext().GetCipheringAlg(), ue.GetSecurityContext().GetDLCount())

			if err = security.NASEncrypt(ue.GetSecurityContext().GetCipheringAlg(), ue.GetSecurityContext().GetKnasEnc(), dlCount.Get(),
				security.Bearer3GPP, security.DirectionDownlink, payload); err != nil {
				return nil, fmt.Errorf("encrypt error: %+v", err)
			}
		}

		// add sequece number
		payload = append([]byte{dlCount.SQN()}, payload[:]...)
		mac32, err := security.NASMacCalculate(ue.GetSecurityContext().GetIntegrityAlg(), ue.GetSecurityContext().GetKnasInt(), dlCount.Get(),
			security.Bearer3GPP, security.DirectionDownlink, payload)
		if err != nil {
			return nil, fmt.Errorf("MAC calcuate error: %+v", err)
		}
		// Add mac value
		payload = append(mac32, payload[:]...)

		// Add EPD and Security Type
		msgSecurityHeader := []byte{msg.SecurityHeader.ProtocolDiscriminator, msg.SecurityHeader.SecurityHeaderType}
		payload = append(msgSecurityHeader, payload[:]...)

		// Increase DL Count
		dlCount.AddOne()
		ue.GetSecurityContext().SetDLCount(dlCount)
		return payload, nil
	}
}

func Decode(ue *context.UEContext, payload []byte, initialMessage bool) (msg *nas.Message, integrityProtected bool, err error) {
	if ue == nil {
		return nil, false, fmt.Errorf("[5GC][NAS] error while decoding: ue is nil")
	}
	if payload == nil {
		return nil, false, fmt.Errorf("[5GC][NAS] error while decoding: NAS payload is empty")
	}
	if len(payload) < 2 {
		return nil, false, fmt.Errorf("[5GC][NAS] error while decoding: NAS payload is too short")
	}

	ulCountNew := ue.GetSecurityContext().GetULCount()

	msg = new(nas.Message)
	msg.ProtocolDiscriminator = payload[0]
	msg.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & 0x0f
	if len(payload) < (1 + 1 + 4 + 1 + 3) {
		return nil, false, fmt.Errorf("[5GC][NAS] error while decoding: NAS payload is too short")
	}
	// information to check integrity and ciphered.

	// sequence number.
	sequenceNumber := payload[6]
	msg.SequenceNumber = sequenceNumber

	// mac verification.
	receivedMac32 := payload[2:6]
	msg.MessageAuthenticationCode = binary.BigEndian.Uint32(receivedMac32)

	// remove security Header except for sequence Number
	payload = payload[6:]

	ciphered := false
	switch msg.SecurityHeaderType {

	case nas.SecurityHeaderTypeIntegrityProtected:
		log.Info("[5GC][NAS] Message with integrity")

	case nas.SecurityHeaderTypeIntegrityProtectedAndCiphered:
		log.Info("[5GC][NAS] Message with integrity and ciphered")
		ciphered = true

	case nas.SecurityHeaderTypeIntegrityProtectedWithNew5gNasSecurityContext:
		log.Info("[5GC][NAS] Message with integrity and with NEW 5G NAS SECURITY CONTEXT")
		ulCountNew.Set(0, 0)

	case nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext:
		log.Info("[5GC][NAS] Message with integrity, ciphered and with NEW 5G NAS SECURITY CONTEXT")
		ciphered = true
		ulCountNew.Set(0, 0)
	default:
		return nil, false, fmt.Errorf("[5GC][NAS] wrong security header type: 0x%0x", msg.SecurityHeader.SecurityHeaderType)
	}

	if ciphered && (ue.GetSecurityContext() == nil) {
		return nil, false, fmt.Errorf("[5GC][NAS] message is ciphered, but UE Security Context is not Available")
	}

	if ue.GetSecurityContext() != nil {
		if ulCountNew.SQN() > sequenceNumber {
			ulCountNew.SetOverflow(ulCountNew.Overflow() + 1)
		}
		ulCountNew.SetSQN(sequenceNumber)

		var mac32 []byte
		mac32, err = security.NASMacCalculate(ue.GetSecurityContext().GetIntegrityAlg(), ue.GetSecurityContext().GetKnasInt(), ulCountNew.Get(),
			security.Bearer3GPP, security.DirectionUplink, payload)
		if err != nil {
			return nil, false, fmt.Errorf("MAC calcuate error: %+v", err)
		}

		if !reflect.DeepEqual(mac32, receivedMac32) {
			log.Info("[5GC][NAS] MAC verification failed(received: ", receivedMac32, ", expected: ", mac32, ")")
		} else {
			integrityProtected = true
		}
	} else {
		log.Info("[5GC][NAS] UE Security Context is not Available, so skip MAC verify")
	}

	if ciphered {
		if !integrityProtected {
			return nil, false, fmt.Errorf("NAS message is ciphered, but MAC verification failed")
		}
		// decrypt payload without sequence number (payload[1])
		if err = security.NASEncrypt(ue.GetSecurityContext().GetCipheringAlg(), ue.GetSecurityContext().GetKnasEnc(), ulCountNew.Get(), security.Bearer3GPP,
			security.DirectionUplink, payload[1:]); err != nil {
			return nil, false, fmt.Errorf("decrypt error: %+v", err)
		}
	}

	// remove sequece Number
	payload = payload[1:]

	err = msg.PlainNasDecode(&payload)

	if err != nil {
		return nil, false, err
	}

	msgTypeText := func() string {
		if msg.GmmMessage == nil {
			return "Non GMM message"
		} else {
			return fmt.Sprintf(" message type %d", msg.GmmHeader.GetMessageType())
		}
	}
	errNoSecurityContext := func() error {
		return fmt.Errorf("UE Security Context is not Available, %s", msgTypeText())
	}
	errWrongSecurityHeader := func() error {
		return fmt.Errorf("wrong security header type: 0x%0x, %s", msg.SecurityHeader.SecurityHeaderType, msgTypeText())
	}
	errMacVerificationFailed := func() error {
		return fmt.Errorf("MAC verification failed, %s", msgTypeText())
	}

	if msg.GmmMessage == nil {
		if ue.GetSecurityContext() == nil {
			return nil, false, errNoSecurityContext()
		}
		if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
			return nil, false, errWrongSecurityHeader()
		}
		if !integrityProtected {
			return nil, false, errMacVerificationFailed()
		}
	} else {
		switch msg.GmmHeader.GetMessageType() {
		case nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration, nas.MsgTypeRegistrationRequest:
			if initialMessage {
				if msg.SecurityHeaderType == nas.SecurityHeaderTypeIntegrityProtectedAndCiphered ||
					msg.SecurityHeaderType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext {
					return nil, false, errWrongSecurityHeader()
				}
			} else {
				if ue.GetSecurityContext() != nil {
					if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
						return nil, false, errWrongSecurityHeader()
					}
					if !integrityProtected {
						return nil, false, errMacVerificationFailed()
					}
				}
			}
		case nas.MsgTypeServiceRequest:
			if initialMessage {
				if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtected {
					return nil, false, errWrongSecurityHeader()
				}
			} else {
				if ue.GetSecurityContext() == nil {
					return nil, false, errNoSecurityContext()
				}
				if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
					return nil, false, errWrongSecurityHeader()
				}
				if !integrityProtected {
					return nil, false, errMacVerificationFailed()
				}
			}
		case nas.MsgTypeIdentityResponse:
			mobileIdentityContents := msg.IdentityResponse.MobileIdentity.GetMobileIdentityContents()
			if len(mobileIdentityContents) >= 1 &&
				nasConvert.GetTypeOfIdentity(mobileIdentityContents[0]) == nasMessage.MobileIdentity5GSTypeSuci {
				// Identity is SUCI
				if ue.GetSecurityContext() != nil {
					if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
						return nil, false, errWrongSecurityHeader()
					}
					if !integrityProtected {
						return nil, false, errMacVerificationFailed()
					}
				}
			} else {
				// Identity is not SUCI
				if ue.GetSecurityContext() == nil {
					return nil, false, errNoSecurityContext()
				}
				if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
					return nil, false, errWrongSecurityHeader()
				}
				if !integrityProtected {
					return nil, false, errMacVerificationFailed()
				}
			}
		case nas.MsgTypeAuthenticationResponse,
			nas.MsgTypeAuthenticationFailure,
			nas.MsgTypeSecurityModeReject,
			nas.MsgTypeDeregistrationAcceptUETerminatedDeregistration:
			if ue.GetSecurityContext() != nil {
				if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
					return nil, false, errWrongSecurityHeader()
				}
				if !integrityProtected {
					return nil, false, errMacVerificationFailed()
				}
			}
		case nas.MsgTypeSecurityModeComplete:
			if ue.GetSecurityContext() == nil {
				return nil, false, errNoSecurityContext()
			}
			if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext {
				return nil, false, errWrongSecurityHeader()
			}
			if !integrityProtected {
				return nil, false, errMacVerificationFailed()
			}
		default:
			if ue.GetSecurityContext() == nil {
				return nil, false, errNoSecurityContext()
			}
			if msg.SecurityHeaderType != nas.SecurityHeaderTypeIntegrityProtectedAndCiphered {
				return nil, false, errWrongSecurityHeader()
			}
			if !integrityProtected {
				return nil, false, errMacVerificationFailed()
			}
		}
	}

	if integrityProtected {
		ue.GetSecurityContext().SetULCount(ulCountNew)
	}
	return msg, integrityProtected, nil
}
