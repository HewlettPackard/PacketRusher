/** SPDX-License-Identifier: Apache-2.0 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func ConfigurationUpdateCommand(ue *context.UEContext, name *context.NetworkName) ([]byte, error) {
	return codec.Encode(ue, buildConfigurationUpdateCommand(name), nas.SecHdrTypeIntegrityProtectedAndCiphered)
}
func buildConfigurationUpdateCommand(name *context.NetworkName) *nas.CfgUpdateCmd {
	msg := new(nas.CfgUpdateCmd)
	if name != nil {
		if name.Full != "" {
			msg.FullNameForNw = &ie.NwName{Ext: 1, TextStr: name.Full}
		}
		if name.Short != "" {
			msg.ShortNameForNw = &ie.NwName{Ext: 1, TextStr: name.Short}
		}
	}
	return msg
}
