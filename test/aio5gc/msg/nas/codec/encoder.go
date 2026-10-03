/** SPDX-License-Identifier: Apache-2.0 */
package codec

import (
	"fmt"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
)

func Encode(ue *context.UEContext, msg nas.Message, st nas.SecHdrType) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("NAS message is nil")
	}
	if st == nas.SecHdrTypePlainNas {
		return msg.MarshalBinary()
	}
	if ue == nil || ue.GetSecurityContext() == nil {
		return nil, fmt.Errorf("NAS security context is nil")
	}
	if st == nas.SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx {
		sc := ue.GetSecurityContext().NASSecurityContext()
		sc.UplinkCount.Set(0, 0)
		sc.DownlinkCount.Set(0, 0)
	}
	return nas.Marshal(msg, ue.GetSecurityContext().NASSecurityContext(), st)
}
