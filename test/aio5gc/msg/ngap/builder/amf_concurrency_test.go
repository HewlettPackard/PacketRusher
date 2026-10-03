// SPDX-License-Identifier: Apache-2.0
package builder

import (
	"sync"
	"testing"

	"github.com/free5gc/nas/ie"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/test/aio5gc/context"
)

// NG Setup/reassociation can run while another association registers a UE.
// Builders need immutable AMF configuration, not a copy of its live UE pool.
func TestAMFBuildersConcurrentWithUEAllocation(t *testing.T) {
	plmn := models.PlmnId{Mcc: "208", Mnc: "93"}
	served := models.PlmnIdNid{Mcc: "208", Mnc: "93"}
	amf := new(context.AMFContext)
	amf.NewAmfContext("amf.5gc.3gppnetwork.org", "196673",
		[]models.Nrf_NFMgmt_PlmnSnssai{{PlmnId: &plmn, SNssaiList: []models.ExtSnssai{{Sst: 1, Sd: "010203"}}}},
		[]models.Guami{{PlmnId: &served, AmfId: "196673"}}, 100, nil, nil)
	ue := amf.NewUE(12)
	ue.SetSecurityCapability(&ie.UESecCapability{Length: 2, EA2_128_5G: true, IA2_128_5G: true})
	ue.SetSecurityContext(new(context.SecurityContext))
	want, err := NGSetupResponse(amf)
	require.NoError(t, err)
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 1000; i++ {
			amf.NewUE(int64(100 + i))
		}
	}()
	t.Cleanup(workers.Wait)
	close(start)
	for i := 0; i < 1000; i++ {
		got, err := NGSetupResponse(amf)
		require.NoError(t, err)
		require.Equal(t, want, got, "UE allocation must not alter NG Setup bytes")
		initial, err := buildInitialContextSetupRequest(nil, ue, amf)
		require.NoError(t, err)
		require.Equal(t, int64(12), initial.RANUENGAPID.Value)
		require.Equal(t, uint8(1), initial.AllowedNSSAI.List[0].SNSSAI.SST.Value[0])
	}
	workers.Wait()
}
