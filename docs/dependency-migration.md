# Latest free5GC codec migration

PacketRusher now uses Go 1.26.2 and these free5GC libraries:

| Module | Version |
| --- | --- |
| nas | v1.3.0 |
| ngap | v1.2.0 |
| openapi | v1.3.0 |
| util | v1.4.0 |
| go-gtp5gnl | v1.6.2 |

NAS and NGAP now use their native typed message APIs. Builders marshal through
the native codecs, and dispatchers consume decoded message types directly.
NAS v1.3.0 uses a documented local replacement at
[`third_party/free5gc-nas`](../third_party/free5gc-nas/README.packetrusher.md). It
corrects Profile B SUCI encoding, restores every baseline-supported optional IE
left unimplemented in dispatched responses, adds UCS-2 network names, and rejects
truncated IE envelopes. Unconsumed container IEs retain owned, bounded value
bytes; their inner procedures are not implemented. Newer unsupported IE warnings
remain errors. The fork retains the original license, source version and upstream
tests. Native-codec tests run explicitly in CI because it is a nested module.
Obsolete local ASN.1 transfer definitions have been removed. Authentication
subscriptions use the current OpenAPI UDR models.

NGAP PLMN identities use their own validated codec following
[TS 38.413, section 9.3.3.5](https://www.etsi.org/deliver/etsi_ts/138400_138499/138413/15.04.00_60/ts_138413v150400p.pdf):
MCC digits followed by MNC digits, low nibble first, with a filler before a
two-digit MNC. NAS places the third MNC digit differently; its PLMN encoding is
unchanged. For MCC 999/MNC 070, NGAP carries `99 09 07` and NAS carries
`99 09 70`. Independent literal-wire and authentication tests cover both layouts
and preserve leading zeros and two-/three-digit MNC identities.

Security Mode Commands require outer security header type 3 (integrity protected
with a new 5G NAS security context), as specified by
[3GPP TS 24.501, section 5.4.2.2](https://www.etsi.org/deliver/etsi_ts/124500_124599/124501/16.12.00_60/ts_124501v161200p.pdf).
Plaintext commands and commands protected with an existing-context header are
rejected before committing security state. Other plaintext 5GMM message types retain their prior security-header policy.
Security Mode Command processing derives a candidate NAS security context and
commits it only after integrity verification. The selected algorithms must have
been offered by the UE, and replayed security capabilities must match the
registration request. Invalid MACs and malformed protected headers leave keys,
algorithms and counters unchanged. Uplink and downlink counters are independent.

PDU session IDs are restricted to 1 through 15. Zero is unassigned and higher
values are reserved by [3GPP TS 24.007, section 11.2.3.1b](https://www.etsi.org/deliver/etsi_ts/124000_124099/124007/19.05.00_60/ts_124007v190500p.pdf).
The CLI rejects requests outside this range; context lookups reject invalid IDs
instead of indexing outside the session table.

The aio5gc test core uses the same native codecs. It has synchronous listener
startup, explicit cleanup, and bounded received frames. Golden NGAP packets
generated from the previous PacketRusher codecs, independent AES-CTR/CMAC
protected optional-IE fixtures, and protected NAS round trips
check compatibility independently of simply encoding and decoding with the same
library. Production Registration Request, Identity Response and the registration
container in Security Mode Complete preserve Profile A/B SUCI ciphertext and MAC
through native wire encoding; SIDF independently authenticates and deconceals
those wire identities in regressions. Registration still advertises N3 data
transfer support, preserving the baseline capability byte. Session iteration takes
a snapshot before invoking callbacks, so a
release callback can remove its own session.

Build the entire CLI package, since it contains multiple source files:

```sh
go build -o packetrusher ./cmd
go test -race ./internal/... ./config ./cmd ./lib/...
go test -race ./test/... -timeout=180s
go test -race github.com/free5gc/nas/ie/... github.com/free5gc/nas/message/... github.com/free5gc/nas/internal/...
```

Testing all internal packages also requires GSL development headers and libraries
(`libgsl-dev` on Debian/Ubuntu). Local SCTP mock-core tests require SCTP support.
The CLI itself builds with `CGO_ENABLED=0`.

The local PR draft records test results and remaining runtime validation. These
codec and mock-core checks do not establish interoperability with a live
free5GC/Open5GS deployment or validate kernel gtp5g traffic.
