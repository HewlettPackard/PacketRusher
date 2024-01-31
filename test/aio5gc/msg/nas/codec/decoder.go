/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package codec

import (
	"encoding/binary"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"reflect"

	"github.com/free5gc/nas"
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

	if integrityProtected {
		ue.GetSecurityContext().SetULCount(ulCountNew)
	}
	return msg, integrityProtected, nil
}
