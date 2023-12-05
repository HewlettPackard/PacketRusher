/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/tools"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	log "github.com/sirupsen/logrus"
)

func ConfigurationUpdateCommand(ue *context.UEContext, nwName *context.NetworkName) ([]byte, error) {

	nasMsg := buildConfigurationUpdateCommand(nwName)
	return tools.Encode(ue, nasMsg)
}

func buildConfigurationUpdateCommand(networkName *context.NetworkName) *nas.Message {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeConfigurationUpdateCommand)

	m.SecurityHeader = nas.SecurityHeader{
		ProtocolDiscriminator: nasMessage.Epd5GSMobilityManagementMessage,
		SecurityHeaderType:    nas.SecurityHeaderTypeIntegrityProtectedAndCiphered,
	}

	configurationUpdateCommand := nasMessage.NewConfigurationUpdateCommand(0)
	configurationUpdateCommand.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	configurationUpdateCommand.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	configurationUpdateCommand.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	configurationUpdateCommand.SetMessageType(nas.MsgTypeConfigurationUpdateCommand)

	if networkName != nil {
		// Full network name
		if networkName.Full != "" {
			fullNetworkName := nasConvert.FullNetworkNameToNas(networkName.Full)
			configurationUpdateCommand.FullNameForNetwork = &fullNetworkName
			configurationUpdateCommand.FullNameForNetwork.SetIei(nasMessage.ConfigurationUpdateCommandFullNameForNetworkType)
		} else {
			log.Info("[5GC][NAS] ConfigurationUpdateCommand: Require Full Network Name, but got nothing.")
		}
		// Short network name
		if networkName.Short != "" {
			shortNetworkName := nasConvert.ShortNetworkNameToNas(networkName.Short)
			configurationUpdateCommand.ShortNameForNetwork = &shortNetworkName
			configurationUpdateCommand.ShortNameForNetwork.SetIei(nasMessage.ConfigurationUpdateCommandShortNameForNetworkType)
		} else {
			log.Info("[5GC][NAS] ConfigurationUpdateCommand: Require Short Network Name, but got nothing.")
		}
	}

	m.GmmMessage.ConfigurationUpdateCommand = configurationUpdateCommand

	return m
}
