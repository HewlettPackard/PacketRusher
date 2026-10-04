# aio5gc

In-process mock of a 5G core (AMF and session management) for the integration tests in `test/`. It answers
NG Setup, registers UEs with 5G AKA and protected NAS, and establishes and releases their PDU sessions; it
has no user plane.

A test builds it with `FiveGCBuilder`, may hook NGAP and NAS messages or UE and PDU session state changes,
and stops it with `Close()`.
