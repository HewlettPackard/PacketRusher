// SPDX-License-Identifier: Apache-2.0
package mm_5gs_test

import (
	"encoding/binary"
	"testing"

	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
)

func TestSuciOverrideIsSentVerbatim(t *testing.T) {
	// Not a decodable identity: unknown type bits, arbitrary length.
	for _, raw := range [][]byte{{0x17}, {0x6e, 0x5c, 0x01}, {0x01, 0x99, 0xf9, 0x07, 0, 0, 0x07, 0, 0xaa, 0xbb}} {
		ue := &context.UEContext{}
		ue.NewRanUeContext("0000000120", &ie.UESecCapability{Length: 2, EA2_128_5G: true, IA2_128_5G: true}, "", "", "", "8000", "000000000000", "999", "70", sidf.HomeNetworkPublicKey{ProtectionScheme: "0", PublicKeyID: "0"}, "0", "internet", 1, "010203", config.TunnelMode(0), nil, nil, 1)
		require.NoError(t, ue.SetSuciOverride(raw))

		packet := mm_5gs.GetRegistrationRequest(ie.RegType_InitialReg, nil, nil, false, ue)
		length := int(binary.BigEndian.Uint16(packet[4:6]))
		require.Equal(t, raw, packet[6:6+length])
	}
	require.Error(t, (&context.UEContext{}).SetSuciOverride(nil))
}
