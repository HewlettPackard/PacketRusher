package ie

import (
	"encoding"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOpaqueNativeValuesValidateLengthAndOwnTheirBuffers(t *testing.T) {
	cases := []struct {
		name string
		ie   interface {
			encoding.BinaryMarshaler
			encoding.BinaryUnmarshaler
		}
		min, max int
	}{
		{"emergency", &EmergNumList{}, 3, 48},
		{"extended emergency", &ExtendedEmergNumList{}, 4, 65535},
		{"SOR", &SORTransparentCntr{}, 17, 65535},
		{"operator categories", &OperatorDefinedAccessCategoryDefs{}, 0, 8320},
		{"EPS status", &EPSBearerCtxStatus{}, 2, 2},
		{"EPS algorithms", &EPSNASSecAlgos{}, 1, 1},
		{"S1 capabilities", &S1UESecCapability{}, 2, 5},
		{"additional information", &AdditionalInfo{}, 1, 255},
		{"mapped EPS", &MappedEPSBearerCtxs{}, 4, 65535},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.min > 0 {
				require.Error(t, tc.ie.UnmarshalBinary(make([]byte, tc.min-1)))
			}
			require.Error(t, tc.ie.UnmarshalBinary(make([]byte, tc.max+1)))
			wire := make([]byte, tc.max)
			wire[0] = 0x5a
			require.NoError(t, tc.ie.UnmarshalBinary(wire))
			wire[0] ^= 0xff
			encoded, err := tc.ie.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, byte(0x5a), encoded[0], "decoder must own its value bytes")
			encoded[0] ^= 0xff
			again, err := tc.ie.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, byte(0x5a), again[0], "encoder must return a separate buffer")
		})
	}
}

func TestSingleOctetCompatibilityFieldsRejectMalformedLengths(t *testing.T) {
	fields := []encoding.BinaryUnmarshaler{&MICOInd{}, &NSSAIInclusionMode{}, &Non3GppNWProvidedPolicies{}, &SMSInd{}, &AlwaysonPDUSessInd{}, &AllowedSSCMode{}, &CongestionReattemptIndicator5GSM{}}
	for _, field := range fields {
		require.Error(t, field.UnmarshalBinary(nil))
		require.Error(t, field.UnmarshalBinary([]byte{1, 2}))
	}
}

func TestUCS2NetworkNamePreservesUnicodeAndRejectsMalformedText(t *testing.T) {
	var name NwName
	wire := []byte{0x90, 0x00, 0x46, 0x00, 0x52, 0x00, 0xe9}
	require.NoError(t, name.UnmarshalBinary(wire))
	require.Equal(t, "FRé", name.TextStr)
	encoded, err := name.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, wire, encoded)
	for _, bad := range [][]byte{{0x90, 0x00}, {0x91, 0x00, 0x46}, {0x90, 0xd8, 0x00}} {
		require.Error(t, name.UnmarshalBinary(bad))
	}
	name.CodeScheme, name.TextStr = CodeScheme_UCS2, "😀"
	_, err = name.MarshalBinary()
	require.Error(t, err)
}

func TestSUCIProfilesRejectTruncatedNativeKeyAndMAC(t *testing.T) {
	for _, scheme := range []uint8{ECIESSchemeProfileA, ECIESSchemeProfileB} {
		keyLen := 32
		if scheme == ECIESSchemeProfileB {
			keyLen = 33
		}
		identity := MobileId5GS{TypeOfId: IdType_5GS_SUCI, SUPIFormat: uint8(SupiIMSI), ProtectionSchemeId: scheme, CipherVal: []byte{1}, ECCEphPubKey: make([]byte, keyLen), MACTag: make([]byte, 8)}
		identity.PlmnId.MCC, identity.PlmnId.MNC = "208", "93"
		good, err := identity.MarshalBinary()
		require.NoError(t, err)
		require.Len(t, good, 8+keyLen+1+8)
		identity.ECCEphPubKey = identity.ECCEphPubKey[:keyLen-1]
		_, err = identity.MarshalBinary()
		require.Error(t, err)
		identity.ECCEphPubKey = make([]byte, keyLen)
		identity.MACTag = identity.MACTag[:7]
		_, err = identity.MarshalBinary()
		require.Error(t, err)
	}
}

func TestN1SMPayloadValueLengthAndReuse(t *testing.T) {
	var payload PayloadCntr
	require.Error(t, payload.UnmarshalBinary(nil, PayloadCntrType_N1SMInfo))
	require.Error(t, payload.UnmarshalBinary(make([]byte, 65536), PayloadCntrType_N1SMInfo))
	require.NoError(t, payload.UnmarshalBinary([]byte{1, 2}, PayloadCntrType_N1SMInfo))
	require.NoError(t, payload.UnmarshalBinary([]byte{3}, PayloadCntrType_N1SMInfo))
	require.Equal(t, []byte{3}, payload.Contents, "reusing an IE must replace the previous value")
	payload.Contents = nil
	_, err := payload.MarshalBinary()
	require.Error(t, err)
}
