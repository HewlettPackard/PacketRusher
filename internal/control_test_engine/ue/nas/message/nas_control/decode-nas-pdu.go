/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas_control

import nas "github.com/free5gc/nas/message"

func GetNasPduFromPduAccept(dlNas *nas.DLNASTransport) nas.Message {
	if dlNas == nil || dlNas.PayloadCntr == nil {
		return nil
	}
	msg, err := nas.Parse(dlNas.PayloadCntr.Contents, nil)
	if err != nil {
		return nil
	}
	return msg
}
