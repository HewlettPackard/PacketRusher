package codec

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/security"
	log "github.com/sirupsen/logrus"
)

func Encode(ue *context.UEContext, msg *nas.Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("NAS Message is nil")
	}

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
