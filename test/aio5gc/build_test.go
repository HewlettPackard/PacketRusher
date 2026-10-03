// SPDX-License-Identifier: Apache-2.0
package aio5gc

import (
	"testing"

	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/util/fsm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	core "my5G-RANTester/test/aio5gc/context"
	tools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
)

func builderConfig() *FiveGCBuilder {
	return new(FiveGCBuilder).WithConfig(tools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:0"), netip.MustParseAddrPort("127.0.0.1:2152"), nil))
}
func TestBuilderReportsDuplicatePolicyRegistration(t *testing.T) {
	cb := func(*fsm.State, fsm.EventType, fsm.ArgsType) {}
	for _, builder := range []*FiveGCBuilder{builderConfig().WithUeCallback(core.Registered, cb).WithUeCallback(core.Registered, cb), builderConfig().WithPDUCallback(core.Active, cb).WithPDUCallback(core.Active, cb), builderConfig().WithNASDispatcherHook(nas.MsgTypeRegReq, nil).WithNASDispatcherHook(nas.MsgTypeRegReq, nil)} {
		fgc, err := builder.Build()
		require.ErrorContains(t, err, "duplicate")
		require.Nil(t, fgc)
	}
}
func TestInitCopiesCallbacksAndHooks(t *testing.T) {
	cfg := builderConfig().config
	var firstCalls, replacedCalls int
	callbacks := fsm.Callbacks{core.AuthenticationInitiated: func(*fsm.State, fsm.EventType, fsm.ArgsType) { firstCalls++ }}
	var coreOne core.Aio5gc
	require.NoError(t, coreOne.Init(cfg, "196673", "first", callbacks, nil))
	require.Len(t, callbacks, 1, "initialization must not fill caller's callback map")
	callbacks[core.AuthenticationInitiated] = func(*fsm.State, fsm.EventType, fsm.ArgsType) { replacedCalls++ }
	ue := coreOne.GetAMFContext().NewUE(1)
	require.NoError(t, ue.GetUeFsm().SendEvent(ue.GetState(), core.RegistrationRequest, fsm.ArgsType{"ue": ue}, logrus.NewEntry(logrus.StandardLogger())))
	require.Equal(t, 1, firstCalls)
	require.Zero(t, replacedCalls)
	hooks := map[nas.MsgType]func(nas.Message, *core.UEContext, *core.GNBContext, *core.Aio5gc) (bool, error){nas.MsgTypeRegReq: func(nas.Message, *core.UEContext, *core.GNBContext, *core.Aio5gc) (bool, error) { return true, nil }}
	coreOne.SetNasHooks(hooks)
	delete(hooks, nas.MsgTypeRegReq)
	require.NotNil(t, coreOne.GetNasHook(nas.MsgTypeRegReq))
	require.NoError(t, coreOne.Close())
}
