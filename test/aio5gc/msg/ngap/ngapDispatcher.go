/** SPDX-License-Identifier: Apache-2.0 */
package ngap

import (
	"fmt"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/ngap/handler"
)

func Dispatch(buf []byte, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	msg, err := message.Parse(buf)
	if err != nil {
		return fmt.Errorf("NGAP decode: %w", err)
	}
	handled := false
	for _, hook := range fgc.GetNgapHooks() {
		done, err := hook(msg, gnb, fgc)
		if err != nil {
			return fmt.Errorf("NGAP scenario hook: %w", err)
		}
		handled = handled || done
	}
	if handled {
		return nil
	}
	switch m := msg.(type) {
	case *message.NGSetupRequest:
		err = handler.NGSetupRequest(m, gnb, fgc)
	case *message.InitialUEMessage:
		err = handler.InitialUEMessage(m, gnb, fgc)
	case *message.UplinkNASTransport:
		err = handler.UplinkNASTransport(m, gnb, fgc)
	case *message.InitialContextSetupResponse:
		err = handler.InitialContextSetupResponse(m, fgc, gnb)
	case *message.PDUSessionResourceSetupResponse:
		err = handler.PDUSessionResourceSetup(m, fgc, gnb)
	case *message.PDUSessionResourceReleaseResponse:
		err = handler.PDUSessionResourceRelease(m, fgc, gnb)
	case *message.UEContextReleaseComplete:
		err = handler.UEContextReleaseComplete(m, fgc, gnb)
	default:
		err = fmt.Errorf("unsupported NGAP message %T", msg)
	}
	return err
}
