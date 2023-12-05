package nas

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/tools"
	nasHandler "my5G-RANTester/test/aio5gc/msg/nas/handler"
	"strconv"

	"github.com/free5gc/nas"
	log "github.com/sirupsen/logrus"
)

func Dispatch(nasPDU *ngapType.NASPDU, ueContext *context.UEContext, fgc *context.Aio5gc) {
	payload := nasPDU.Value
	m := new(nas.Message)
	m.SecurityHeaderType = nas.GetSecurityHeaderType(payload) & 0x0f
	var msg *nas.Message
	var err error

	if m.SecurityHeaderType != nas.SecurityHeaderTypePlainNas {
		var integrityProtected bool
		msg, integrityProtected, err = tools.Decode(ueContext, payload, false)
		if !integrityProtected {
			log.Fatal(errors.New("[5GC][NAS] message integrity could not be verified"))
		}

	} else {
		msg, err = tools.DecodePlainNasNoIntegrityCheck(payload)
	}

	// Hook for changing 5GC behaviour
	hook := fgc.GetNasHook()
	if hook != nil {
		handled, err := hook(msg, ueContext, fgc)
		if err != nil {
			log.Fatal(err.Error())
		} else if handled {
			return
		}
	}

	switch msg.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		log.Info("[5GC][NAS] Received Registration Request")
		err = nasHandler.RegistrationRequest(msg, fgc, ueContext)

	case nas.MsgTypeAuthenticationResponse:
		log.Info("[5GC][NAS] Received Authentication Response")
		err = nasHandler.AuthenticationResponse(msg, fgc, ueContext)

	case nas.MsgTypeSecurityModeComplete:
		log.Info("[5GC][NAS] Received Security Mode Complete")
		err = nasHandler.SecurityModeComplete(msg, fgc, ueContext)

	case nas.MsgTypeRegistrationComplete:
		log.Info("[5GC][NAS] Received Registration Complete")
		nasHandler.RegistrationComplete(msg, fgc, ueContext)

	case nas.MsgTypeULNASTransport:
		log.Info("[5GC][NAS] Received UL NAS Transport")
		err = nasHandler.UlNasTransport(msg, fgc, ueContext)

	default:
		err = errors.New("[5GC][NAS] unrecognised nas message type: " + strconv.Itoa(int(msg.GmmHeader.GetMessageType())))
	}
	if err != nil {
		log.Fatal(err)
	}
}
