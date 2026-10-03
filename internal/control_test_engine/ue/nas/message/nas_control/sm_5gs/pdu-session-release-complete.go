/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sm_5gs

import nas "github.com/free5gc/nas/message"

func GetPduSessionReleaseComplete(id uint8) []byte {
	return encodePlain(&nas.PDUSessRelComplete{PDUSessId: id, PTI: 1})
}
