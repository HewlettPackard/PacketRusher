/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func PDUSessionReleaseCommand(ue *context.UEContext, smContext context.SmContext, cause uint8) (msg []byte, err error) {
	nasMsg := buildPDUSessionReleaseCommand(ue, smContext, cause)
	nasPdu, err := nasMsg.PlainNasEncode()
	if err != nil {
		return nil, err
	}
	nasMsg, err = buildDLNASTransport(ue, nasPdu, uint8(smContext.GetPduSessionId()))
	if err != nil {
		return nil, err
	}
	return codec.Encode(ue, nasMsg)
}

func buildPDUSessionReleaseCommand(ue *context.UEContext, smContext context.SmContext, cause uint8) *nas.Message {
	m := nas.NewMessage()
	m.GsmMessage = nas.NewGsmMessage()
	m.GsmHeader.SetMessageType(nas.MsgTypePDUSessionReleaseCommand)
	m.GsmHeader.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSSessionManagementMessage)
	m.PDUSessionReleaseCommand = nasMessage.NewPDUSessionReleaseCommand(0x0)
	pDUSessionReleaseCommand := m.PDUSessionReleaseCommand

	pDUSessionReleaseCommand.SetMessageType(nas.MsgTypePDUSessionReleaseCommand)
	pDUSessionReleaseCommand.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSSessionManagementMessage)
	pDUSessionReleaseCommand.SetPDUSessionID(uint8(smContext.GetPduSessionId()))

	pDUSessionReleaseCommand.SetPTI(smContext.GetPti())
	pDUSessionReleaseCommand.SetCauseValue(cause)

	return m
}
