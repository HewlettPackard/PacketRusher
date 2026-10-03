/** SPDX-License-Identifier: Apache-2.0 */
package builder

import (
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
)

func DownlinkNASTransport(nas []byte, ue *context.UEContext) ([]byte, error) {
	msg, err := BuildDownlinkNASTransport(nas, ue)
	if err != nil {
		return nil, err
	}
	return msg.MarshalBinary()
}
func BuildDownlinkNASTransport(nas []byte, ue *context.UEContext) (*message.DownlinkNASTransport, error) {
	return &message.DownlinkNASTransport{AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: ue.GetRanNgapId()}, NASPDU: &ie.NASPDU{Value: nas}}, nil
}
