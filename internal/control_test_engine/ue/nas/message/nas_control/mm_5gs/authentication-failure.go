/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
)

func AuthenticationFailure(cause, eapMsg string, param []byte) []byte {
	msg := &nas.AuthFailure{Cause5GMM: &ie.Cause5GMM{Value: ie.Cause5GMM_MACFailure}}
	if cause == "SQN failure" {
		msg.Cause5GMM.Value = ie.Cause5GMM_SynchFailure
		msg.AuthFailureParam = &ie.AuthFailureParam{Value: param}
	}
	return encodePlain(msg)
}
