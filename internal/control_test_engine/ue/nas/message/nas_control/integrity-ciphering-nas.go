package nas_control

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/security"
)

func EncodeNasPduWithSecurity(ue *context.UEContext, pdu []byte, securityHeaderType uint8, securityContextAvailable, newSecurityContext bool) ([]byte, error) {
	m := nas.NewMessage()
	err := m.PlainNasDecode(&pdu)
	if err != nil {
		return nil, err
	}
	m.SecurityHeader = nas.SecurityHeader{
		ProtocolDiscriminator: nasMessage.Epd5GSMobilityManagementMessage,
		SecurityHeaderType:    securityHeaderType,
	}
	return NASEncode(ue, m, securityContextAvailable, newSecurityContext)
}

func NASEncode(ue *context.UEContext, msg *nas.Message, securityContextAvailable bool, newSecurityContext bool) (payload []byte, err error) {
	var sequenceNumber uint8
	if ue == nil {
		err = fmt.Errorf("amfUe is nil")
		return
	}
	if msg == nil {
		err = fmt.Errorf("Nas message is empty")
		return
	}

	if !securityContextAvailable {
		return msg.PlainNasEncode()
	} else {
		if newSecurityContext {
			ue.UeSecurity.ULCount.Set(0, 0)
			ue.UeSecurity.DLCount.Set(0, 0)
		}

		sequenceNumber = ue.UeSecurity.ULCount.SQN()
		payload, err = msg.PlainNasEncode()
		if err != nil {
			return
		}

		// TODO: Support for ue has nas connection in both accessType
		// make ciphering of NAS message.
		if err = security.NASEncrypt(ue.UeSecurity.CipheringAlg, ue.UeSecurity.KnasEnc, ue.UeSecurity.ULCount.Get(), security.Bearer3GPP,
			security.DirectionUplink, payload); err != nil {
			return
		}

		// add sequence number
		payload = append([]byte{sequenceNumber}, payload[:]...)
		mac32 := make([]byte, 4)

		mac32, err = security.NASMacCalculate(ue.UeSecurity.IntegrityAlg, ue.UeSecurity.KnasInt, ue.UeSecurity.ULCount.Get(), security.Bearer3GPP, security.DirectionUplink, payload)
		if err != nil {
			return
		}

		// Add mac value
		payload = append(mac32, payload[:]...)
		// Add EPD and Security Type
		msgSecurityHeader := []byte{msg.SecurityHeader.ProtocolDiscriminator, msg.SecurityHeader.SecurityHeaderType}
		payload = append(msgSecurityHeader, payload[:]...)

		// Increase UL Count
		ue.UeSecurity.ULCount.AddOne()
	}
	return
}
