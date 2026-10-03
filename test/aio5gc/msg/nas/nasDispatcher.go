/** SPDX-License-Identifier: Apache-2.0 */
package nas

import (
	"fmt"
	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/ngap/ie"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
	handler "my5G-RANTester/test/aio5gc/msg/nas/handler"
)

func Dispatch(pdu *ie.NASPDU, ue *context.UEContext, fgc *context.Aio5gc, gnb *context.GNBContext) error {
	ue.LockProtocol()
	defer ue.UnlockProtocol()
	if pdu == nil {
		return fmt.Errorf("[5GC][NAS] Missing NAS PDU")
	}
	payload := pdu.Value
	st := nas.GetSecHdrType(payload)
	var msg nas.Message
	var err error
	switch ue.GetState().Current() {
	case context.Authenticated, context.Registered:
		if st == nas.SecHdrTypePlainNas {
			return fmt.Errorf("[5GC][NAS] Plain NAS in authenticated state")
		}
		var verified bool
		msg, verified, err = codec.Decode(ue, payload, false)
		if err != nil || !verified {
			return fmt.Errorf("[5GC][NAS] Integrity verification failed: %v", err)
		}
	default:
		if st != nas.SecHdrTypePlainNas {
			return fmt.Errorf("[5GC][NAS] Protected NAS before authentication")
		}
		msg, err = codec.DecodePlainNasNoIntegrityCheck(payload)
	}
	if err != nil || msg == nil {
		return fmt.Errorf("[5GC][NAS] Decode failed: %v", err)
	}
	if hook := fgc.GetNasHook(msg.MsgType()); hook != nil {
		handled, hookErr := hook(msg, ue, gnb, fgc)
		if hookErr != nil {
			return fmt.Errorf("NAS scenario hook: %w", hookErr)
		}
		if handled {
			return nil
		}
	}
	amf, session := fgc.GetAMFContext(), fgc.GetSessionContext()
	switch m := msg.(type) {
	case *nas.RegReq:
		err = handler.RegistrationRequest(m, amf, ue, gnb)
	case *nas.IdRsp:
		err = handler.IdentityResponse(m, amf, ue, gnb)
	case *nas.AuthRsp:
		err = handler.AuthenticationResponse(m, gnb, ue, amf)
	case *nas.SecModeComplete:
		err = handler.SecurityModeComplete(m, amf, ue, gnb)
	case *nas.RegComplete:
		err = handler.RegistrationComplete(m, gnb, ue, amf)
	case *nas.ULNASTransport:
		err = handler.UlNasTransport(m, gnb, ue, session)
	case *nas.CfgUpdateComplete:
	case *nas.DeregReqUEOrig:
		err = handler.UEOriginatingDeregistration(m, amf, ue, gnb)
	default:
		err = fmt.Errorf("[5GC][NAS] Unsupported message %s", msg.MsgType())
	}
	return err
}
