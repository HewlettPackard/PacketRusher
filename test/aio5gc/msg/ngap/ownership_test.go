// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"fmt"
	"testing"

	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/stretchr/testify/require"
	core "my5G-RANTester/test/aio5gc/context"
	tools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
)

func TestEncodedReleaseRequiresOwningAssociation(t *testing.T) {
	cfg := tools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:0"), netip.MustParseAddrPort("127.0.0.1:2152"), nil)
	var fgc core.Aio5gc
	require.NoError(t, fgc.Init(cfg, "196673", "test", nil, nil))
	t.Cleanup(func() { require.NoError(t, fgc.Close()) })
	first, second := new(core.GNBContext), new(core.GNBContext)
	ue := fgc.GetAMFContext().NewUEForGNB(1, first)
	other := fgc.GetAMFContext().NewUEForGNB(1, second)
	sm := core.NewSmContext(1)
	sm.GetState().Set(core.Active)
	require.NoError(t, ue.AddSmContext(sm))
	_, err := core.ReleasePDUSession(ue, 1)
	require.NoError(t, err)
	response := &message.PDUSessionResourceReleaseResponse{AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: 1}, PDUSessionResourceReleasedListRelRes: &ie.PDUSessionResourceReleasedListRelRes{List: []ie.PDUSessionResourceReleasedItemRelRes{{PDUSessionID: &ie.PDUSessionID{Value: 1}, PDUSessionResourceReleaseResponseTransfer: &aper.OctetString{0}}}}}
	wire, err := response.MarshalBinary()
	require.NoError(t, err)
	require.ErrorContains(t, Dispatch(wire, second, &fgc), "another gNB association")
	untouched, err := ue.GetSmContext(1)
	require.NoError(t, err)
	require.Same(t, sm, untouched)
	require.Equal(t, core.InactivePending, sm.GetState().Current())
	require.NotEqual(t, ue.GetAmfNgapId(), other.GetAmfNgapId())
	require.NoError(t, Dispatch(wire, first, &fgc))
	_, err = ue.GetSmContext(1)
	require.Error(t, err)
}
func TestHookErrorsDoNotFallThroughToDefaultProtocol(t *testing.T) {
	var fgc core.Aio5gc
	fgc.SetNgapHooks([]func(message.Message, *core.GNBContext, *core.Aio5gc) (bool, error){func(message.Message, *core.GNBContext, *core.Aio5gc) (bool, error) {
		return false, fmt.Errorf("policy rejected message")
	}})
	wire, err := (&message.UEContextReleaseComplete{AMFUENGAPID: &ie.AMFUENGAPID{Value: 0}, RANUENGAPID: &ie.RANUENGAPID{Value: 1}}).MarshalBinary()
	require.NoError(t, err)
	require.ErrorContains(t, Dispatch(wire, new(core.GNBContext), &fgc), "policy rejected message")
	require.NoError(t, fgc.Close())
}
