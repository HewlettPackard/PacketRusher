/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package interface_management

import (
	ngap "github.com/free5gc/ngap/message"
)

func AmfConfigurationUpdateAcknowledge() ([]byte, error) {
	message := BuildAmfConfigurationUpdateAcknowledge()

	return message.MarshalBinary()
}

func BuildAmfConfigurationUpdateAcknowledge() (pdu *ngap.AMFConfigurationUpdateAcknowledge) {
	pdu = &ngap.AMFConfigurationUpdateAcknowledge{}
	return
}
