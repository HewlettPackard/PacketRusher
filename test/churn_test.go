/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/stretchr/testify/require"
)

// A churn cohort member registers, and once deregistered stays parked, even though
// it loops, until it is rearmed; it then registers again. Terminate while parked ends
// it. At each step the state the member reports, which the wave driver acts on, agrees
// with the AMF's. It has no PDU session, so its deregistration does not depend on the
// test 5GC releasing one.
func TestChurnCohortMemberParksUntilRearmed(t *testing.T) {
	conf := amfTools.GenerateDefaultConf(
		netip.MustParseAddrPort("127.0.0.1:9389"),
		netip.MustParseAddrPort("127.0.0.1:2754"),
		[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38428")}}},
	)

	fiveGC, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).Build()
	require.NoError(t, err)
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(1, conf, &wg)
	time.Sleep(time.Second)

	scenario := make(chan procedures.UeTesterMessage)
	member := tools.NewChurnMember()
	ueCfg := tools.UESimulationConfig{
		UeId:             1,
		Gnbs:             gnbs,
		Cfg:              conf,
		ScenarioChan:     scenario,
		RegistrationLoop: true,
		Churn:            member,
	}

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(tools.IncrementMsin(ueCfg.UeId, conf.Ue.Msin))
	securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})
	fiveGC.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext)

	state := func() fsm.StateType {
		var s fsm.StateType
		fiveGC.GetAMFContext().ExecuteForAllUe(func(ue *context.UEContext) { s = ue.GetState().Current() })
		return s
	}
	registered := func() bool { return state() == context.Registered && member.State() == tools.ChurnRegistered }
	deregistered := func() bool { return state() == context.Deregistered && member.State() == tools.ChurnParked }

	tools.SimulateSingleUE(ueCfg, &wg)
	require.Eventually(t, registered, 10*time.Second, 50*time.Millisecond, "the UE should register")

	scenario <- procedures.UeTesterMessage{Type: procedures.Terminate}
	require.Eventually(t, deregistered, 10*time.Second, 50*time.Millisecond, "the UE should deregister")

	// A looping UE that is not in a cohort registers again after
	// TimeBeforeReregistration, here 0. A cohort member must wait for its rearm.
	require.Never(t, registered, 2*time.Second, 50*time.Millisecond, "a parked member must not register before it is rearmed")

	require.Equal(t, tools.ChurnParked, member.State())
	require.True(t, member.Rearm())
	require.Eventually(t, registered, 10*time.Second, 50*time.Millisecond, "the rearmed UE should register again")

	scenario <- procedures.UeTesterMessage{Type: procedures.Terminate}
	require.Eventually(t, deregistered, 10*time.Second, 50*time.Millisecond, "the UE should deregister again")

	// Parked again, Terminate ends it: the send is unbuffered, so it completes only if
	// the parked UE reads it.
	select {
	case scenario <- procedures.UeTesterMessage{Type: procedures.Terminate}:
	case <-time.After(5 * time.Second):
		t.Fatal("a parked member should still answer Terminate")
	}
	select {
	case <-member.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("the member should have returned")
	}
	require.Equal(t, tools.ChurnGone, member.State())
}
