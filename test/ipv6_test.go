/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

// A UE configured for IPv4v6 PDU sessions registers and is given both an IPv4
// address and an IPv6 interface identifier, of which it makes its link-local address.
func TestIPv4v6PDUSession(t *testing.T) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:9491"), netip.MustParseAddrPort("127.0.0.1:2156"),
		[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38416")}}})
	conf.Ue.PDUSessionType = config.PDUSessionIPv4v6

	// logrus cannot remove a hook, so swap the set out and restore it afterwards.
	hooks := log.StandardLogger().ReplaceHooks(make(log.LevelHooks))
	t.Cleanup(func() { log.StandardLogger().ReplaceHooks(hooks) })
	logs := logtest.NewLocal(log.StandardLogger())

	fiveGC, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).Build()
	require.NoError(t, err)
	t.Cleanup(func() { _ = fiveGC.Close() })
	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(1, conf, &wg)
	t.Cleanup(func() {
		for _, gnb := range gnbs {
			gnb.Terminate()
		}
	})
	time.Sleep(1 * time.Second)

	subscriber := context.SecurityContext{}
	subscriber.SetMsin(tools.IncrementMsin(1, conf.Ue.Msin))
	subscriber.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	subscriber.SetAbba([]uint8{0x00, 0x00})
	fiveGC.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, subscriber)
	simulation := tools.SimulateSingleUE(tools.UESimulationConfig{UeId: 1, Gnbs: gnbs, Cfg: conf, NumPduSessions: 1}, &wg)
	t.Cleanup(func() { stopTestSimulations(t, []*tools.UESimulation{simulation}) })

	// The core's first session has 10.0.0.2, also as its interface identifier.
	logged := func(message string) bool {
		for _, entry := range logs.AllEntries() {
			if entry.Message == message {
				return true
			}
		}
		return false
	}
	require.Eventually(t, func() bool {
		return logged("[UE][NAS] PDU session 1 address: 10.0.0.2") &&
			logged("[UE][NAS] PDU session 1 IPv6 link-local address: fe80::a00:2")
	}, 30*time.Second, 10*time.Millisecond, "the UE must decode both addresses of its PDU session")
}
