# Local 5GC fixture

`aio5gc` supplies a small in-process AMF and session manager for PacketRusher's SCTP integration tests. It performs 5G AKA, protected NAS registration, IPv4 PDU-session establishment/release and NGAP association handling. It does not exercise an external free5GC/Open5GS installation or forward user-plane traffic.

The fixture uses the native typed messages in free5GC NAS v1.3 and NGAP v1.2. NAS dispatcher hooks accept `message.Message` keyed by `message.MsgType`; NGAP hooks accept `ngap/message.Message`. Return `true` from a hook to consume the message before normal dispatch. Builders construct typed message fields and encode through the upstream codecs.

`FiveGCBuilder.Build()` binds configured listeners before returning and reports bind failures to the caller. A test should register `t.Cleanup(func() { _ = core.Close() })` immediately after building. `Close()` is idempotent, rejects new workers, closes listeners and accepted associations, and waits for active readers. Register any listener created directly with `core.RegisterCloser(listener.Close)` before starting `service.Serve`. Each accepted association owns its connection, so replies from an old association cannot be redirected to its replacement.

FSM callbacks run for entry, exit and explicit events. Count registrations or activations on `fsm.EntryEvent` and protect shared callback assertions with a mutex. UE/PDU iteration visits a snapshot; callbacks may remove contexts without retaining the collection lock. Confirming one PDU-session release leaves other sessions intact.

Run codec, security, session and shutdown regressions with `go test -race ./test/aio5gc/...`. Run SCTP scenarios with `go test -race ./test -timeout 120s`; the host must support SCTP and permit local listening sockets. The migration retains byte-for-byte NGAP golden packets from the previous implementation, checks protected NAS round trips and tampering rejection, and covers independent NAS uplink/downlink counters.
