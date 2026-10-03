// SPDX-License-Identifier: Apache-2.0
package context

import (
	stdcontext "context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestIndependentCoreIDsAndRegistryOwnership(t *testing.T) {
	var first, second AMFContext
	require.Equal(t, int64(0), first.NewUE(1).GetAmfNgapId())
	require.Equal(t, int64(0), second.NewUE(1).GetAmfNgapId())
	require.Equal(t, int32(1), first.TmsiAllocate())
	require.Equal(t, int32(1), second.TmsiAllocate())
	old, newPeer := new(GNBContext), new(GNBContext)
	require.NoError(t, first.AddGnb("peer", old))
	require.NoError(t, first.AddGnb("peer", newPeer))
	first.RemoveGnb("peer", old)
	found, err := first.GetGnb("peer")
	require.NoError(t, err)
	require.Same(t, newPeer, found)
	require.Zero(t, second.GNBCount())
	first.RemoveGnb("peer", newPeer)
	require.Zero(t, first.GNBCount())
}
func TestGNBIdentitySemanticLookupAndValueIsolation(t *testing.T) {
	var amf AMFContext
	gnb := new(GNBContext)
	identity := models.GlobalRanNodeId{PlmnId: &models.PlmnId{Mcc: "208", Mnc: "93"}, GNbId: &models.GNbId{BitLength: 24, GNBValue: "000008"}}
	gnb.SetGlobalRanNodeID(identity)
	require.NoError(t, amf.AddGnb("peer", gnb))
	identity.PlmnId.Mcc = "999"
	equivalent := models.GlobalRanNodeId{PlmnId: &models.PlmnId{Mcc: "208", Mnc: "93"}, GNbId: &models.GNbId{BitLength: 24, GNBValue: "000008"}}
	found, err := amf.FindGnbById(equivalent)
	require.NoError(t, err)
	require.Same(t, gnb, found)
	snapshot := gnb.GetGlobalRanNodeID()
	snapshot.GNbId.GNBValue = "ffffff"
	require.Equal(t, "000008", gnb.GetGlobalRanNodeID().GNbId.GNBValue)
}
func TestSubscriberCopiesOwnSecretsAndRegistrationCounters(t *testing.T) {
	var amf AMFContext
	security := SecurityContext{msin: "120", abba: []byte{1, 2}, kgnb: []byte{3}, NH: []byte{4}}
	security.SetAuthSubscription("key", "opc", "op", "8000", "000001")
	require.NoError(t, amf.Provision(models.Snssai{Sst: 1}, security))
	security.authenticationSubs.SequenceNumber.Sqn = "ffffff"
	security.abba[0] = 9
	first, err := amf.FindProvisionedData("120")
	require.NoError(t, err)
	copy := first.GetSecurityContext()
	require.Equal(t, "000001", copy.GetAuthSubscription().SequenceNumber.Sqn)
	require.Equal(t, byte(1), copy.abba[0])
	copy.authenticationSubs.SequenceNumber.Sqn = "bbbbbb"
	copy.abba[0] = 8
	copy.kgnb[0] = 8
	copy.NH[0] = 8
	second, err := amf.FindProvisionedData("120")
	require.NoError(t, err)
	other := second.GetSecurityContext()
	require.Equal(t, "000001", other.authenticationSubs.SequenceNumber.Sqn)
	require.Equal(t, []byte{1, 2}, other.abba)
	require.Equal(t, []byte{3}, other.kgnb)
	require.Equal(t, []byte{4}, other.NH)
	getter := other.GetAuthSubscription()
	getter.SequenceNumber.Sqn = "aaaaaa"
	require.Equal(t, "000001", other.GetAuthSubscription().SequenceNumber.Sqn)
}
func TestOwnedWorkerCancellationAndFiniteClose(t *testing.T) {
	var core Aio5gc
	var ended atomic.Bool
	require.True(t, core.Go(func() { <-core.Context().Done(); ended.Store(true) }))
	require.NoError(t, core.Close())
	require.True(t, ended.Load())
	require.False(t, core.Go(func() { t.Error("late worker ran") }))
	var blocked Aio5gc
	release := make(chan struct{})
	require.True(t, blocked.Go(func() { <-release }))
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, blocked.CloseContext(ctx), stdcontext.DeadlineExceeded)
	close(release)
	require.NoError(t, blocked.Close())
	require.NoError(t, blocked.Close())
}
func TestResourceRetiresExactlyOnceAndPreservesCloseError(t *testing.T) {
	var core Aio5gc
	var calls atomic.Int32
	want := errors.New("close proof")
	retire := core.Own(func() error { calls.Add(1); return want })
	require.Equal(t, 1, core.ResourceCount())
	require.ErrorIs(t, retire(), want)
	require.ErrorIs(t, retire(), want)
	require.Zero(t, core.ResourceCount())
	require.Equal(t, int32(1), calls.Load())
	core.Own(func() error { return want })
	require.ErrorIs(t, core.Close(), want)
	require.ErrorIs(t, core.Close(), want)
}
func TestCoreObservationsAreDetachedAndWaitCannotLoseWakeup(t *testing.T) {
	var core Aio5gc
	machine, err := initUeFSM(core.observedCallbacks(nil, false))
	require.NoError(t, err)
	core.amfContext.ueFsm = machine
	core.amfContext.onCreate = core.observeCreatedUE
	ue := core.amfContext.NewUE(1)
	ue.SetSecurityContext(&SecurityContext{msin: "120"})
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), time.Second)
	defer cancel()
	var waiters sync.WaitGroup
	waiters.Add(1)
	done := make(chan error, 1)
	go func() {
		defer waiters.Done()
		_, err := core.Wait(ctx, func(s Snapshot) bool { return len(s.UEs) == 1 && s.UEs[0].State == Authenticated })
		done <- err
	}()
	require.NoError(t, machine.SendEvent(ue.GetState(), RegistrationRequest, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger())))
	require.NoError(t, machine.SendEvent(ue.GetState(), AuthenticationSuccess, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger())))
	require.NoError(t, <-done)
	waiters.Wait()
	snapshot := core.Snapshot()
	snapshot.UEs[0].Entries[Authenticated] = 99
	snapshot.UEs[0].Sessions[1] = PDURecord{State: Active}
	require.Equal(t, uint64(1), core.Snapshot().UEs[0].Entries[Authenticated])
	require.Empty(t, core.Snapshot().UEs[0].Sessions)
	require.NoError(t, core.Close())
	_, err = core.Wait(ctx, func(Snapshot) bool { return false })
	require.ErrorIs(t, err, stdcontext.Canceled)
}
