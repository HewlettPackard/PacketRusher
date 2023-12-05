/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sm_5gs

import (
	"bytes"
	"fmt"
	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func GetPduSessionReleaseComplete(pduSessionId uint8) (nasPdu []byte) {
	m := nas.NewMessage()
	m.GsmMessage = nas.NewGsmMessage()
	m.GsmHeader.SetMessageType(nas.MsgTypePDUSessionReleaseComplete)

	pduSessionReleaseComplete := nasMessage.NewPDUSessionReleaseComplete(0)
	pduSessionReleaseComplete.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSSessionManagementMessage)
	pduSessionReleaseComplete.SetMessageType(nas.MsgTypePDUSessionReleaseComplete)
	pduSessionReleaseComplete.PDUSessionID.SetPDUSessionID(pduSessionId)
	pduSessionReleaseComplete.PTI.SetPTI(0x01)

	m.GsmMessage.PDUSessionReleaseComplete = pduSessionReleaseComplete

	data := new(bytes.Buffer)
	err := m.GsmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}