// SPDX-License-Identifier: Apache-2.0
package sm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRequestedPDUSessionFamilyOnWire(t *testing.T) {
	for _, family := range []uint8{ie.PDUSessType_IPv4, ie.PDUSessType_IPv6, ie.PDUSessType_IPv4v6} {
		packet := GetPduSessionEstablishmentRequest(7, family)
		require.Equal(t, []byte{0x2e, 7, 1, 0xc1, 0xff, 0xff, 0x90 | family}, packet[:7])
		message, err := nas.Parse(packet, nil)
		require.NoError(t, err)
		require.Equal(t, family, message.(*nas.PDUSessEstReq).PDUSessType.Value)
	}
	require.Equal(t, GetPduSessionEstablishmentRequest(7, 1), GetPduSessionEstablishmentRequest(7))
	require.Nil(t, GetPduSessionEstablishmentRequest(7, 4))
}
