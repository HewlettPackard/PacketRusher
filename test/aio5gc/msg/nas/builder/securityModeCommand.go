/** SPDX-License-Identifier: Apache-2.0 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/common/auth"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func SecurityModeCommand(ue *context.UEContext) ([]byte, error) {
	integrity, ciphering := auth.SelectAlgorithms(ue.GetSecurityCapability())
	ue.GetSecurityContext().SetCipheringAlg(ciphering)
	ue.GetSecurityContext().SetIntegrityAlg(integrity)
	ue.GetSecurityContext().DerivateAlgKey()
	return codec.Encode(ue, buildSecurityModeCommand(ue), nas.SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx)
}
func buildSecurityModeCommand(ue *context.UEContext) *nas.SecModeCmd {
	ksi := ue.GetNgKsi()
	requestIMEISV := uint8(0)
	if ue.GetPei() == "" {
		requestIMEISV = 1
	}
	return &nas.SecModeCmd{
		SelectedNASSecAlgos: &ie.NASSecAlgos{CipheringAlgo: ie.AlgCiphering(ue.GetSecurityContext().GetCipheringAlg()), MsgIntAlgo: ie.AlgIntegrity(ue.GetSecurityContext().GetIntegrityAlg())},
		Ngksi:               &ksi, ReplayedUESecCapabilities: ue.GetSecurityCapability(),
		IMEISVReq: &ie.IMEISVReq{Value: requestIMEISV}, Additional5GSecInfo: &ie.Additional5GSecInfo{RINMR: true},
	}
}
