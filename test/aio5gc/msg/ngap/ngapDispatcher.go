/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package ngap

import (
	"fmt"
	"github.com/free5gc/ngap/message"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/ngap/handler"
)

func Dispatch(buf []byte, gnb *context.GNBContext, fgc *context.Aio5gc) {
	msg, err := message.Parse(buf)
	if err != nil {
		log.Errorf("[5GC][NGAP] Decode failed: %v", err)
		return
	}
	handled := false
	for _, hook := range fgc.GetNgapHooks() {
		done, err := hook(msg, gnb, fgc)
		if err != nil {
			log.Error(err)
		}
		handled = handled || done
	}
	if handled {
		return
	}
	switch m := msg.(type) {
	case *message.NGSetupRequest:
		err = handler.NGSetupRequest(m, gnb, fgc)
	case *message.InitialUEMessage:
		err = handler.InitialUEMessage(m, gnb, fgc)
	case *message.UplinkNASTransport:
		err = handler.UplinkNASTransport(m, gnb, fgc)
	case *message.InitialContextSetupResponse:
		err = handler.InitialContextSetupResponse(m, fgc)
	case *message.PDUSessionResourceSetupResponse:
		err = handler.PDUSessionResourceSetup(m, fgc)
	case *message.PDUSessionResourceReleaseResponse:
		err = handler.PDUSessionResourceRelease(m, fgc)
	case *message.UEContextReleaseComplete:
		err = handler.UEContextReleaseComplete(m, fgc)
	default:
		err = fmt.Errorf("unsupported NGAP message %T", msg)
	}
	if err != nil {
		log.Errorf("[5GC][NGAP] Handle failed: %v", err)
	}
}
