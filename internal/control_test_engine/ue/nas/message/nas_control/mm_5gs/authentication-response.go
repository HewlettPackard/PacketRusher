/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"encoding/base64"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
)

func AuthenticationResponse(param []uint8, eapMsg string) []byte {
	msg := &nas.AuthRsp{}
	if len(param) > 0 {
		msg.AuthRspParam = &ie.AuthRspParam{Res: param}
	} else if eapMsg != "" {
		raw, err := base64.StdEncoding.DecodeString(eapMsg)
		if err != nil {
			return nil
		}
		msg.EAPMsg = &ie.EAPMsg{Eap: raw}
	}
	return encodePlain(msg)
}
