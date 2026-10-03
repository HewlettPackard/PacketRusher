# Native local 5GC test fixture

`aio5gc` is an in-process AMF and session manager for PacketRusher tests. It runs actual SCTP associations, typed NGAP v1.2 and NAS v1.3 messages, 5G AKA, integrity-protected NAS, and IPv4/IPv6 session signaling. It has no forwarding UPF and does not establish interoperability with external free5GC/Open5GS installations. Dataplane tests use their own explicitly modeled peers.

## Test author API

Use `testkit` for ordinary registration/session scenarios. It owns the core, production gNBs and UE simulations; provisioning errors are returned. Kernel-assigned SCTP endpoints eliminate process-derived port arithmetic. Fixed endpoint configurations remain supported for retry, reassociation and address allocation tests.

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
builder := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig())
fixture, err := testkit.Start(ctx, builder, 1)
require.NoError(t, err)
t.Cleanup(func() { require.NoError(t, fixture.Close()) })
require.NoError(t, fixture.Provision(1))
simulation, err := fixture.StartUE(tools.UESimulationConfig{
    UeId: 1, NumPduSessions: 1,
})
require.NoError(t, err)
attachment, err := simulation.Execute(ctx, "wait", "")
require.NoError(t, err) // Real UE actor accepted registration and every PDU.
require.Equal(t, []uint8{1}, attachment.ActivePDUSessions)
```

`fixture.Config` and `fixture.Core.Config()` contain effective bound AMF ports. `core.Snapshot().Associations` exposes the actual accepted local/remote SCTP endpoints. Configure inputs before building; do not mutate a published fixture's configuration. `Start` rejects an already canceled context before binding. Node construction still calls the legacy synchronous `InitGnb`, whose finite dial/setup retry loop is not canceled by the caller's context and can call `log.Fatal` when retries/configuration fail. Use a valid local fixture config; a context deadline does not currently bound that construction step.

## Observation and gates

`core.Snapshot()` contains copied UE/PDU records and transition counts, sorted by AMF UE ID. `core.Wait(ctx, predicate)` checks a snapshot and subscribes to the next change atomically; no polling sleep is needed. `Entries[Authenticated]` and `Sessions[id].Entries[Active]` replace bespoke callback-counter maps. Session records retain the last observed transition after the live session is removed. An `Active` core entry occurs **before** the encoded PDU accept is sent; confirm client readiness independently through `simulation.Execute`. Assert `snapshot.Errors` when success is expected: decoder, scenario-hook, handler, encoder and send failures propagate through native dispatch.

`testkit.Gate` pauses a real decoded protocol phase without preventing core shutdown:

```go
gate := testkit.NewGate()
builder.WithNGAPDispatcherHook(func(msg message.Message, peer *core.GNBContext, fgc *core.Aio5gc) (bool, error) {
    if _, ok := msg.(*message.InitialUEMessage); ok {
        if err := gate.Wait(fgc.Context()); err != nil { return true, err }
    }
    return false, nil // Continue the native handler after opening.
})
// Wait on gate.Reached() with a test deadline; call gate.Open() to continue.
```

Raw NAS/NGAP hooks remain available for scenario-specific native handover messages. A `true` result consumes a message; a hook error stops dispatch and is observed instead of falling through to normal protocol. FSM callbacks receive entry, exit and explicit events; counting only `fsm.EntryEvent` avoids double counting. Builder input maps/configs are copied, and duplicate policy registrations return a build error. User callbacks run outside registry/runtime/observation locks and must honor `core.Context()` if they block.

## Ownership and shutdown

Each core owns its registries, TMSI/AMF UE-ID allocators, hooks and observation stream. Each accepted gNB association owns one immutable connection. Core UE identity is AMF UE ID plus the owning association/RAN UE ID; matching numeric IDs from another peer cannot release a session or change NAS security counters. Scenario handover policy explicitly calls `ue.BindGNB(target, targetRANID)`. `FindGnbById` returns the live association pointer, never a copied mutex-bearing object. Identity/config/subscriber getters used as values are detached copies.

`Core.Go` admits a worker before spawning it. `Core.Own` registers an idempotent resource retirement function, so completed readers remove both their closer and association entry. `Stop` cancels policy gates before closing resources; `CloseContext` rejects new work and joins all admitted core workers/closers, or returns the caller's deadline without claiming completion. `Close` uses a ten-second bound. Concurrent closes observe the same shutdown result. A Linux poller-owned nonblocking listener closes a blocked accept without a wakeup dial. `Listener.Accepting()` lets transport tests prove that accept reached its empty-queue wait.

Fixture cleanup stops core policy, explicitly terminates every node, sends Kill to simulations, and joins the constructor/actor wait group and core workers. Production node signal waiters observe node termination and unregister their signal notification. Legacy gNB association/reassociation routines are not all enrolled in that wait group; node termination closes their sockets and prevents further association publication. This is distinct from the stronger joined-worker contract of `Core.CloseContext`.

Switch-off deregistration may overtake a session release ACK. Force release moves `InactivePending` to `Inactive` before retirement. A late ACK is accepted only for a known inactive retirement on the owning UE; an unknown ID or a new active session remains an error.

## Validation

Pure codec/security/ownership checks: `go test -race ./test/aio5gc/...`.
Native lifecycle and all real fixtures: `go test -race ./test ./test/aio5gc/... -count=1 -timeout=180s` in a disposable Linux network namespace with loopback up and SCTP available. The tests retain NGAP golden bytes, NAS tampering/integrity and independent uplink/downlink counters, strict registration/automatic loop metrics, encoded NG/Xn handover completion, and the fixed-port multi-AMF/retry allocation assertions.
