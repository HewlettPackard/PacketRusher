// SPDX-License-Identifier: Apache-2.0
package codec

import (
	"reflect"
	"testing"

	message "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/testutil/naswire"
)

func TestNativeOptionalIECompatibilityInMockDecoder(t *testing.T) {
	// Check the codec contract for every restored MM/SM field. The production
	// dispatch tests separately check message direction and security negotiation.
	for _, fixture := range naswire.OptionalIEs {
		t.Run(fixture.Name, func(t *testing.T) {
			plain := naswire.Hex(fixture.Wire)
			parsed, err := DecodePlainNasNoIntegrityCheck(plain)
			require.NoError(t, err)
			require.False(t, reflect.ValueOf(parsed).Elem().FieldByName(fixture.Field).IsNil())
			retained, err := parsed.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, plain, retained)
			if plain[0] == 0x2e {
				directUE := newSecurityUE()
				_, directVerified, directErr := Decode(directUE, plain, false)
				require.NoError(t, directErr)
				require.False(t, directVerified, "plain SM carries no integrity-protected header")
			}
			nested := plain[0] == 0x2e
			if nested {
				plain = naswire.WrapSM(plain)
			}
			ue := newSecurityUE()
			security := ue.GetSecurityContext()
			beforeUL, beforeDL := security.GetULCount(), security.GetDLCount()
			packet := naswire.Protect(plain, security.GetKnasEnc(), security.GetKnasInt(), 7, 0, 2)
			bad := append([]byte(nil), packet...)
			bad[2] ^= 0x80
			_, verified, err := Decode(ue, bad, false)
			require.Error(t, err)
			require.False(t, verified)
			require.Equal(t, beforeUL, security.GetULCount())
			require.Equal(t, beforeDL, security.GetDLCount())
			parsed, verified, err = Decode(ue, packet, false)
			require.NoError(t, err)
			require.True(t, verified)
			afterUL := security.GetULCount()
			require.Equal(t, uint32(7), afterUL.Get())
			require.Equal(t, beforeDL, security.GetDLCount())
			if nested {
				parsed, err = message.Parse(parsed.(*message.DLNASTransport).PayloadCntr.Contents, nil)
				require.NoError(t, err)
			}
			require.False(t, reflect.ValueOf(parsed).Elem().FieldByName(fixture.Field).IsNil())
		})
	}
}

func TestMalformedOptionalIECannotCommitMockReceiveCounter(t *testing.T) {
	for _, wire := range []string{"7e0068010000", "7e0042010134040201f1", "7e004201017a00050001f100", "7e0042010160032000", "7e004201016002200060042000", "7e00680100052e0101c31f2402aa"} {
		ue := newSecurityUE()
		security := ue.GetSecurityContext()
		before := security.GetULCount()
		plain := naswire.Hex(wire)
		_, err := DecodePlainNasNoIntegrityCheck(plain)
		require.Error(t, err)
		packet := naswire.Protect(plain, security.GetKnasEnc(), security.GetKnasInt(), 7, 0, 2)
		_, verified, err := Decode(ue, packet, false)
		require.Error(t, err)
		require.False(t, verified)
		require.Equal(t, before, security.GetULCount())
	}
}
