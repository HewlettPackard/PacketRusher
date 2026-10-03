// SPDX-License-Identifier: Apache-2.0
package builder

import (
	nasIE "github.com/free5gc/nas/ie"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/stretchr/testify/require"
	codec "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc/context"
	"net"
	"testing"
)

func TestNGAPSetupTransferNegotiatedIPv6Family(t *testing.T) {
	for _, tc := range []struct {
		nas  uint8
		ngap aper.Enumerated
	}{{nasIE.PDUSessType_IPv6, ie.PDUSessionTypePresentIpv6}, {nasIE.PDUSessType_IPv4v6, ie.PDUSessionTypePresentIpv4v6}} {
		session := new(context.SessionContext)
		session.NewSessionContext()
		sm := context.NewSmContext(1)
		sm.SetPduSessionType(tc.nas)
		sm.SetSessionRule(session.GetSessionRules()[0])
		sm.SetDefQosQFI(1)
		wire, err := buildPDUSessionResourceSetuprequestTransfert(*sm, net.ParseIP("127.0.0.1"))
		require.NoError(t, err)
		transfer := new(ie.PDUSessionResourceSetupRequestTransfer)
		require.NoError(t, codec.Unmarshal(wire, transfer))
		found := false
		for _, field := range transfer.ProtocolIEs.List {
			if field.PDUSessionType != nil {
				found = true
				require.Equal(t, tc.ngap, field.PDUSessionType.Value)
			}
		}
		require.True(t, found)
	}
}
