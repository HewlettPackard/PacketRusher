package nasMsgHandler

import (
	"my5G-RANTester/test/amf/context"
)

func PDUSessionEstablishmentAccept(ue *context.UEContext) (msg []byte, err error) {

	return buildSessionEstablishmentAccept(*ue)
}

func buildSessionEstablishmentAccept(ue context.UEContext) (msg []byte, err error) {

	return
}
