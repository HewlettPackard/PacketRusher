/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"bytes"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func ConfigurationUpdateComplete(ue *context.UEContext) ([]byte, error) {
	pdu := getConfigurationUpdateComplete()
	// pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	// if err != nil {
	// 	return nil, fmt.Errorf("error encoding %s IMSI UE NAS Registration Complete Msg", ue.UeSecurity.Supi)
	// }

	return pdu, nil
}

func getConfigurationUpdateComplete() (nasPdu []byte) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeConfigurationUpdateComplete)

	configurationUpdateComplete := nasMessage.NewConfigurationUpdateComplete(0)
	configurationUpdateComplete.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	configurationUpdateComplete.SetSecurityHeaderType(0x00)
	configurationUpdateComplete.SetSpareHalfOctet(0x00)
	configurationUpdateComplete.SetMessageType(nas.MsgTypeConfigurationUpdateComplete)

	m.GmmMessage.ConfigurationUpdateComplete = configurationUpdateComplete

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}