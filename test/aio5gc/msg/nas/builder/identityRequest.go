/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package builder

import (
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
)

func IdentityRequest(ue *context.UEContext) (nasPdu []byte, err error) {
	nasMsg, err := buildIdentityRequest(ue)
	if err != nil {
		return nil, err
	}

	return codec.Encode(ue, nasMsg)
}

func buildIdentityRequest(ue *context.UEContext) (nasMsg *nas.Message, err error) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeIdentityRequest)

	if ue.GetState().Is(context.Registered) {
		m.SecurityHeader = nas.SecurityHeader{
			ProtocolDiscriminator: nasMessage.Epd5GSMobilityManagementMessage,
			SecurityHeaderType:    nas.SecurityHeaderTypeIntegrityProtectedAndCiphered,
		}
	}

	identityRequest := nasMessage.NewIdentityRequest(0)
	identityRequest.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	identityRequest.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	identityRequest.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	identityRequest.IdentityRequestMessageIdentity.SetMessageType(nas.MsgTypeIdentityRequest)
	identityRequest.SpareHalfOctetAndIdentityType.SetTypeOfIdentity(0x01)

	m.GmmMessage.IdentityRequest = identityRequest

	return m, nil
}
