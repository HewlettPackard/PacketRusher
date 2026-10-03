// SPDX-License-Identifier: Apache-2.0
package mm_5gs_test

import (
	"crypto/ecdh"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/free5gc/nas/ie"
	message "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	corecontext "my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func TestProductionSUCIProfilesSurviveRegistrationIdentityAndEmbeddedWire(t *testing.T) {
	vectors := []struct {
		name, scheme, public, private string
		curve                         ecdh.Curve
		ephemeralLength               int
	}{
		{"ProfileA", "1", "5a8d38864820197c3394b92613b20b91633cbd897119273bf8e4a6f4eec0a650", "c53c22208b61860b06c62e5406a7b330c2b577aa5558981510d128247d38bd1d", ecdh.X25519(), 32},
		{"ProfileB", "2", "0472DA71976234CE833A6907425867B82E074D44EF907DFB4B3E21C1C2256EBCD15A7DED52FCBB097A4ED250E036C7B9C8C7004C4EEDC4F068CD7BF8D3F900E3B4", "F1AB1074477EBCC7F554EA1C5FC368B1616730155E0041AC447D6301975FECDA", ecdh.P256(), 33},
	}
	for _, vector := range vectors {
		for _, msin := range []string{"001002086", "0010020862"} {
			t.Run(vector.name+"/"+msin, func(t *testing.T) {
				publicBytes, err := hex.DecodeString(vector.public)
				require.NoError(t, err)
				public, err := vector.curve.NewPublicKey(publicBytes)
				require.NoError(t, err)
				privateBytes, err := hex.DecodeString(vector.private)
				require.NoError(t, err)
				private, err := vector.curve.NewPrivateKey(privateBytes)
				require.NoError(t, err)
				keys := []sidf.HomeNetworkPrivateKey{{ProtectionScheme: vector.scheme, PrivateKey: private, PublicKey: public}}
				ue := &context.UEContext{}
				ue.NewRanUeContext(msin, &ie.UESecCapability{Length: 2, EA2_128_5G: true, IA2_128_5G: true}, "", "", "", "8000", "000000000000", "208", "93", sidf.HomeNetworkPublicKey{ProtectionScheme: vector.scheme, PublicKeyID: "1", PublicKey: public}, "0", "internet", 1, "010203", config.TunnelMode(0), nil, nil, 1)
				expected := "imsi-20893" + msin
				check := func(identity *ie.MobileId5GS) {
					t.Helper()
					supi, err := sidf.ToSupi(identity.SUCIStr(), keys)
					require.NoError(t, err, "SIDF must independently authenticate and deconceal the transmitted SUCI")
					require.Equal(t, expected, supi)
					require.Len(t, identity.ECCEphPubKey, vector.ephemeralLength)
					require.Len(t, identity.CipherVal, (len(msin)+1)/2)
					require.Len(t, identity.MACTag, 8)
				}
				registration := mm_5gs.GetRegistrationRequest(ie.RegType_InitialReg, nil, nil, true, ue)
				response := mm_5gs.IdentityResponse(ue)
				for name, packet := range map[string][]byte{"registration": registration, "identity": response} {
					t.Run(name, func(t *testing.T) {
						require.NotEmpty(t, packet)
						identityOffset := 5
						if name == "registration" {
							identityOffset = 6
						}
						identityLength := int(binary.BigEndian.Uint16(packet[identityOffset-2 : identityOffset]))
						require.Equal(t, 8+vector.ephemeralLength+(len(msin)+1)/2+8, identityLength, "SUCI must have no appended bytes")
						raw := packet[identityOffset : identityOffset+identityLength]
						var decoded ie.MobileId5GS
						require.NoError(t, decoded.UnmarshalBinary(raw))
						check(&decoded)
						// The actual aio5gc plaintext decoder receives the production builder wire.
						parsed, err := codec.DecodePlainNasNoIntegrityCheck(packet)
						require.NoError(t, err)
						if name == "registration" {
							request := parsed.(*message.RegReq)
							check(request.MobileId5GS)
							capability, err := request.Capability5GMM.MarshalBinary()
							require.NoError(t, err)
							require.Equal(t, []byte{7}, capability, "N3 data transfer capability must retain baseline wire byte")
							require.True(t, request.Capability5GMM.N3Data)
						} else {
							check(parsed.(*message.IdRsp).MobileId)
						}
					})
				}
				// Security Mode Complete embeds a fresh production Registration Request,
				// passing through native identity encoding again before NAS protection.
				security := new(corecontext.SecurityContext)
				security.SetCipheringAlg(2)
				security.SetIntegrityAlg(2)
				core := new(corecontext.UEContext)
				core.SetSecurityContext(security)
				protected, err := mm_5gs.SecurityModeComplete(ue, 1)
				require.NoError(t, err)
				parsed, verified, err := codec.Decode(core, protected, false)
				require.NoError(t, err)
				require.True(t, verified)
				complete := parsed.(*message.SecModeComplete)
				require.NotNil(t, complete.NASMsgCntr)
				embedded, err := codec.DecodePlainNasNoIntegrityCheck(complete.NASMsgCntr.Contents)
				require.NoError(t, err)
				check(embedded.(*message.RegReq).MobileId5GS)
			})
		}
	}
}
